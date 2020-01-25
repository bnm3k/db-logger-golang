# Using postgres & leveldb for logging in Golang

Golang's standard library provides the handy [log](https://golang.org/pkg/log/) package which suffices for most simple logging cases. Usually, the logs are output to the standard streams (stdout or stderr). From there, they can be redirected as needed when launching the app via terminal. Alternatively, the log outputs can also be written directly to a file. Still, I was curious as to what it would take to use a relational database (like Postgres) as a destination for the log outputs without changing too much what other programmers and even parts of the program expect of the log Package and its functions. Therefore when reading this post, please do take it as more of exploratory rather than didactic or some sort of sagely best-practices.



### Overview of the log Package

The log package gives us two options: we can either create our own logger manually, or we can use its default standard logger. We'll opt to create our own logger assuming that different parts of the program will require their own specialized logger (e.g. for logging errors only or for logging informational messages only). Furthermore, everything we do with a customized logger can be extended to work with the default logger provided.



### Creating a logger

Using [`log.New`](https://golang.org/src/log/log.go?s=2897:2953#L52), we can create our own custom loggers. `log.New` has the following signature:

```go
func New(out io.Writer, prefix string, flag int) *Logger
```



Let's start with the `out` parameter. The key thing to note, is that the argument doesn't necessarily have to be a file or one of the standard streams such as `os.Stdout`- all it has to be, or rather do, is implement the `io.Writer` interface. Here is the `io.Writer` interface:

```go
type Writer interface {
    Write(p []byte) (n int, err error)
}
```

This then gives us a lot of freedom. For our use case, we can implement our own interface that sort of 'redirects' all log outputs to postgres.



The next parameter in the `log.New` function is the `prefix` parameter. This gives us a basic way to create leveled logging, that is, different loggers can use different prefixes to set themselves apart. 



Finally, we have the `flag` parameter which adds additional prefixes to the log output. These additional prefixes can contain the date and/or the time the log was created plus  (if required) the associated filename and/or filepath from which the log was generated. In order to set the `flag`correctly, we have to use the constants that the package provides us.



### Parsing and structuring the log output (preliminaries)

With all the flags that give maximum information in the prefix set (Ldate | Ltime | Lmicroseconds | Llongfile), plus adding the prefix 'ERROR', the log ouput has the following format:

```
ERROR 2009/01/23 01:23:23.123123 /a/b/c/d.go:23: some error message
```

We could dump the output as is to postgres, using the `text` data type for storage, and call it a day. However, since we're going through all these to use Postgres, we might as well take full advantage of it and enforce a structure for the log storage.

In order to do so, first, we must parse the log string. This looks like a job for the mighty `regex`.



We can already see the structures that we want to extract, such as the date, the prefix, the associated file, and most importantly, the actual log message.



I'm more comfortable using javascript for regex stuff so that's where I tinkered for a while before settling on the pattern to use Note that, since a user can set their own flags, (e.g. getting rid of the filepath since for example, when logging a http request, the filepath is irrelevant), I had to take this into account and make the related patterns optional:

```javascript
const parseLog = (() => {
    const r = /^(\w+)\s+(\d{4}\/\d{2}\/\d{2}\s)?(\d{2}:\d{2}:\d{2}(\.\d+)?\s)?(.*\.go:\d+:\s)?([^]*)/;
    return logStr => {
        let match = r.exec(logStr);
        return match
            ? {
                  prefix: match[1],
                  date: match[2] && match[2].trim(),
                  time: match[3] && match[3].trim(),
                  file: match[5] && match[5].trim(),
                  payload: match[6]
              }
            : null;
    };
})();
```

`parseLog` closes over the `r` pattern since I didn't want the pattern to be recompiled every time the function is called - though this might very much be unnecessary, I probably should check the relevant MDN docs but haven't.

The regex can be broken down as follows:

1. `/^`: the beginning, standard stuff

2. `(\w+)\s+`: matches the prefix which is expected to be alphanumeric characters only plus a bit of space. When we go back to Golang, we must find a way to enforce this, for example, when creating the logger

3. `(\d{4}/\d{2}/\d{2}\s)?`: this matches the date portion of the log output plus a bit of space. However, the date can be ommitted therefore the match is made optional.

4. `(\d{2}:\d{2}:\d{2}(.\d+)?\s)?`:  matches the time portion, the microseconds may or may not be provided. As with the date, we must also take into account that the time can be ommitted

5. `(.*\.go:\d+:\s)?`: matches the file part. From golang's documentation of the log output, we know that regardless of whether the full path or just the file name is provided, a colon is appended at the end. I'm also assuming that all files have the `.go` extension. This is a 'known unknown'. There are probably some  '[unknown unknown](https://en.wikipedia.org/wiki/There_are_known_knowns)' assumptions that I'm making in this regex pattern that might result in errors later on, but for now, these assumptions remain in the realm of the unknown unknowns.

6. `([^]*)`: Finally, we use this match to capture the actual payload of the log output. It's supposed to match every character including a newline character.
   
   

Using `parseLog` with the sample log output provided earlier, we get:

```javascript
{ 
    prefix: 'ERROR',
    date: '2009/01/23',
    time: '01:23:23.123123',
    file: '/a/b/c/d.go:23:',
    payload: 'some error message' 
}
```

Satisfied with the javascript `parseLog` as it is for now, the next step was to translate it to Go. This was a bit tricky for me since up to that point, I'd never used regular expressions in Go so I had to spend some time working through Go's regex package.



### Parsing and structuring the log outputs (Golang implementation)

I usually find myself front-loading a lot of the key design decisions when using Go, which is great to some extent (most of the times) since I'm still going to have to think about and formalize such matters at some point either way. But sometimes it leads to premature over-abstraction. In javascript though, I often find myself freestyling until I arrive at what I want; it's only by forcing myself lately to use TDD that I've started front-loading design decisions in js too.



Back to logging: I've opted to sort of encapsulate the regex pattern into a struct with its own `type` to allow for coupling associated methods (such as `parseLog`) and also allow for different regexes to be used depending on the logger: again a hunch tells me this might be over-abstraction...



We'll have a `type customOut` which will encapsulate the regex pattern as so:

```go
type customOut struct {
	r *regexp.Regexp
}
```



Therefore, we'll have to supply some means for initialzing `customOut` with the default regex pattern. I used the back-ticks since when using the usual double quotes for strings, I have to escape all the backslashes in the regex pattern which is cumbersome and adds unnecessary noise. I've also used `MustCompile` since it's more terse and I am not dealing with a dynamic pattern. Lastly, in order to capture the payload, I've changed the pattern from `([^]*)` to `([\w\W]*)` since the former throws an error for some reason:

```go
func newCustomOut() *customOut {
	return &customOut{
		r: regexp.MustCompile(`^(\w+)\s+(\d{4}\/\d{2}\/\d{2}\s)?(\d{2}:\d{2}:\d{2}(\.\d+)?\s)?(.*\.go:\d+:\s)?([\w\n]*)`),
	}
}
```



Separately, we'll also have a `type` that captures parsed logs:

```go
type parsedLog struct {
	Prefix  string
	LogTime time.Time
	File    string
	Payload string
}
```



Back to `customOut`, we'll add the following method for parsing the logs:

```go
func (c *customOut) parseLog(str string) (parsedLog, error) {
	var err error = nil
	var pl parsedLog

	matches := c.r.FindStringSubmatch(str)
	if matches != nil {
		var logTime time.Time
		logTime, err = parseLogTime(matches[2], matches[3])
		pl = parsedLog{
			Prefix:  matches[1],
			LogTime: logTime,
			File:    strings.TrimSpace(matches[5]),
			Payload: matches[6],
		}
	} else {
		err = ErrInvalidLog
	}

	if err != nil {
		return parsedLog{}, ErrInvalidLog
	}
	return pl, nil
}
```



Since I opted to store the date and time into a `time.Time` variable rather than a `string` variable, I have to convert them. Hence the `parseLogTime` function. It gives us a bit of flexibility but postgres is already great at parsing date and time strings into timestamp so this might be unnecessary work on the application's part. Without further ado, here's the `parseLogTime` function. Note that we account for the extra space that logger adds after the date and time values, back in javascript, we used `trim` to get rid of such space characters. Alternatively, we could have used extra groups to match out just the date and time portions without space in the regex but I opted for otherwise since it made the regex much harder to inspect by eye.

```go
func parseLogTime(dateVal, timeVal string) (time.Time, error) {
	now := time.Now()
	var t time.Time
	var err error = nil
	if dateVal == "" && timeVal == "" {
		// No date val. No time val
		return now, nil
	} else if dateVal == "" {
		// Only time val provided"
		y, m, d := now.Date()
		dtValStr := fmt.Sprintf("%v/%02d/%02d %s", y, m, d, timeVal)
		t, err = time.Parse("2006/01/02 15:04:05.999999 ", dtValStr)
	} else if timeVal == "" {
		// Only date val provided"
		t, err = time.Parse("2006/01/02 ", dateVal)
	} else {
		// Both date val and time val provided
		dtValStr := fmt.Sprintf("%s%s", dateVal, timeVal)
		t, err = time.Parse("2006/01/02 15:04:05.999999 ", dtValStr)
	}

	return t, err
}
```



Back in `parseLog`, we also have the `ErrInvalidLog` error just in case something goes wrong and we need to give feedback:

```go
//ErrInvalidLog ..
var ErrInvalidLog = errors.New("Invalid Log. Unable to Parse")
```



So far, all the types and methods have been private since, if this is to be repackaged into a reusable package, the user shouldn't have to care about the how the log is parsed, all they'd require is a `logger` equivalent to what Go's standard library provides.



### Setting up Postgres

Before going any further, since the ultimate goal is to store the logs in Postgres, it's best to think about how the table(s) should be designed. 



Courtesy of how `parseLogTime` is structured, we know that we'll always have a `LogTime` regardless of whether the user adds the date/time flags- the rest of the values though might be omitted. With that in mind, our log table probably needs a primary key. At first, I thought of using the `LogTime` value as the primary key for each entry, since we also get indexing for free which will come in handy when querying the logs. However, even if it's highly improbable, it is quite possible that two different logs might end up having the same log time and one of them will have to be discarded (due to the uniqueness constraint for primary keys). Another alternative is to use a synthetic key (e.g. an incrementing integer) in combination, or even in leau of the timestamp. But, I opted to forgo having a primary key altogether until such a need arose - e.g. if I need to use some collumn in the table as a foreign key.



Another aspect that needs to be considered is which type to use for the `LogTime` value. Postgres provides two types for timestamps, `timestamp` and `timestamptz`. With `timestamp`, we simply take the log time as it is and store it.  While choosing one or the other, we have to take into consideration the fact that our application and the postgres server might be running in two different timezones, or even simply that postgres is configured to a different timezone. For now, we'll go with `timestamp`, and just as with the primary key, consider `timestamptz` when the need arises.



All in all, the table definition ends up being as follows:

```sql
create table log(
    prefix varchar(15),
    log_time timestamp,
    file text,
    payload text
);
```



Finally, since we expect that we'll be doing a lot of range queries based on the `log_time`, we probably should add an index to it but again, we'll delay that decision until it's necessitated. There are also a couple of other nitty gritties like initializing the actual database but those go without saying.



### Writing logs to postgres

Back in our Go application, we need to glue both the logger and postgres together.



We'll offload the labor of setting up a connection to Postgres to the logger user rather than setting it up ourselves within the logger constructor. This adds a lot of flexibility. It also allows the same `*sql.DB*` instance to be reused across the application. There are also other benefits, for example, if we give the end-user the responsibility of crafting the sql insertion statement, they can shift to some other relational database such as MySQL (which uses a different syntax for parameters), or they can use a different name for the logs tables instead of `log`.  For now, let's narrow down our focus until everything is working fine.



Since `customOut` is already being used to encapsulate the regex pattern, it can also be used to encapsulate the `db` as follows:

```go
type customOut struct {
	r          *regexp.Regexp
	db         *sql.DB
}
```



We then use the following function to create instances of the logger:

```go
func NewCustomLogger(prefix string, flag int, db *sql.DB) (*log.Logger, error) {
	match, err := regexp.MatchString("^\\w+$", prefix)
	if err != nil || match == false {
		return nil, ErrInvalidPrefix
	}
	cOut := newCustomOut(db)
	return log.New(cOut, prefix+"\t", flag), nil

}
```



The prefix is constrained to alphanumeric characters only (no spaces, tabs or special characters and symbols). This is because the regex in `parseLog` already assumes so and if we were to leave out this check, it would result in certain errors and malformed outputs depending on the prefix. 



An additional `ErrInvalidPrefix` is included to make it clear to the caller of the function:

```go
var ErrInvalidPrefix = errors.New("Invalid Prefix")
```



We also have to modify the `newCustomOut` method so as to add the `db` parameter:

```go
func newCustomOut(db *sql.DB) *customOut {
	return &customOut{
		r: regexp.MustCompile(`^(\w+)\s+(\d{4}\/\d{2}\/\d{2}\s)?(\d{2}:\d{2}:\d{2}(\.\d+)?\s)?(.*\.go:\d+:\s)?([\w\n]*)`),
		db:  db,
	}
}
```



Finally, the pièce de résistance, the last piece of the puzzle: implementing the `io.Writer` interface in `customOut`:

```go
func (c *customOut) Write(log []byte) (n int, err error) {
	pl, err := c.parseLog(string(log))
	if err != nil {
		fmt.Println(err)
	} else {
	    insertStmt := "insert into log(prefix, created_at, file, payload) values ($1, $2, $3, $4)"
		_, err = c.db.Exec(insertStmt, pl.Prefix, pl.LogTime, pl.File, pl.Payload)
		if err != nil {
			fmt.Println(err)
		}
	}
	return len(log), err
}
```



Voila! We can now pass `cOut` directly to `log.New`. Keep in mind though, when we run `c.db.Exec`, we probably do need to do something more intelligent when an error is returned.



### Querying logs

The whole point of storing the logs in Postgres should be the query flexibility and endless options we get out of the box. We could use `listen/notify`  plus triggers and other 'stuff' to set up a poor man's monitoring system. We could also add full text indexing on the payload and start edging in on [Elasticsearch](https://www.elastic.co/products/log-monitoring) territory. For now, we'll settle for simply querying all the log outputs that occured in the last 24 hours:

```sql
select prefix, log_time, file, payload 
from log 
where log_time >= now() - '1 day'::interval
```



Declaring and using intervals in postgres sql is so expressive that it's quite easy to modify the above query to instead retrieve the logs from the past 1 week:

```sql
select prefix, log_time, file, payload 
from log 
where log_time >= now() - '1 week'::interval
```



Another advantage of using intervals rather than handrolling our own calculations is that, under the hood, postgres takes care of a lot of edge cases that come with dealing with date/time data whenever we use intervals.



We can also use the prefix to retrieve only certain kinds of log messages, eg Errors:

```sql
select prefix, log_time, file, payload
from log
where log_time >= now() - '1 Week'::interval and prefix = 'ERROR'
```

The querying options are endless and if we need to incorporate some additional dimension of our logs, such as the process number, we simply add a new collumn to our table and use it in our queries.



### Increasing write throughput

As it stands, any part of our application that uses the custom Logger has to wait for postgres to confirm the insertion. This ropes in additional latency (particularly compared to simply logging to stdout or even to a file).



Now, so far, I've insisted on delaying optimizations and extensions (such as adding an index to the log_time column) until they are severely needed. Therefore, this section sort of betrays any pragmatism I hitherto held. Still, I thought it might be a fun undertaking just for the sake of it. 



However, if I'm to be fully thorough in increasing throughput, I should have probably carried out benchmark tests to see just how much latency I've shaven off and tried out other approaches, which I didn't. So take this section (and even the entire article for that matter) with a grain of salt, see if it's worth it and tinker it to your own needs.



### Using an embedded database instead (leveldb)

For simple applications though, using Postgres for logging is a bit excessive. Even if both the application and Postgres are running on the same machine, Postgres still has to run as an entire server on a different process. Moreoever, if our instance of Postgres fails or is shut down, or een the connection is messed up, our application, (with the logging code we have so far), also fails along. 



We could add a failover strategy such as redirecting to stdout or a file (or even a csv which will be easier to bulk insert into postgres later on). 



Or, we could simply use an embedded database which runs in the same process as our application. 



Let's use an embedded database. Since we're already using a relational database, the option that would require the least amount of modification is [sqlite](https://www.sqlite.org/index.html). Still, I wanted to try out something more fun, maybe to the tune of [leveldb](https://en.wikipedia.org/wiki/LevelDB). 



Like sqlite, leveldb is 'embeddable'. Unlike sqlite, leveldb is a key-value nosql database. Therefore, we don't really need to parse and structure our log output, we could just dump it as it is into leveldb. However, we do need to think about which key to use. Such a key not only has to be unique, it also has to facilitate efficient querying of the log output. I decided to use a concatenation of both the log prefix and the timestamp as the key:

```go
//add key here
```

I could (and should) also append a random value such as a [short-id](https://github.com/teris-io/shortid) since two separte logs could potentially have the same prefix and timestamp but that's let's pretend that'll never happen.



Additionally, since the log output is already being parsed into a struct when we were working with Postgres, it might as well be stored as json, all we have to do is add those tag-thingies and make all the struct fields exportable:

```go
// new struct field
```



Finally, all that's left is insertion into leveldb

```go
// inserting into leveldb
```



The great thing about relational databases is that they give us a lot of flexibility and options when it comes to querying as we have seen. Furthermore, their rich array of data types allow us to encode more aspects of our data. With key-value stores though, all we have to work with are keys, for which leveldb only sees as opaque byte arrays, regardless of what they encode. Since leveldb uses lexicographical order to sort the keys, we have to keep in mind some of the assumptions inherent in the key format used above. For one, it's expected that querying will be limited to a select prefix. Secondly, given that lexicographic order is not the same as temporal or even numeric order, we might (and will) have situation where an earlier timestamp is 'greater' than a more recent timestamp when we do know that the opposite is true. For example, under lexicographic order, the following relation between the timestamps holds true:

```
15" > "1479953943"
```

For the time being, we'll wrongly assume that all timestamps have the same number of digits so that ordering works as expected. 



With all that in mind, querying for logs such as `ERROR` logs output in the last 24 hours is as follows:










































