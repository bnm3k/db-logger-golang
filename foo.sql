

select prefix, log_time, file, payload
from log
where log_time >= now() - '24 hours'::interval and prefix = 'ERROR'