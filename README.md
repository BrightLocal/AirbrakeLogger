# airbrake-logger
Command line parameters:
  * ```-beanstalk 0.0.0.0:11300``` -- connect to beanstalkd server. Default: none.
  * ```-queue Airbrake``` -- beanstalkd queue name to watch. Default: ```Airbrake```.
  * ```-listen 0.0.0.0:11311``` -- open TCP socket for messages. Default: none.
  * ```-url https://api.airbrake.io/notifier_api/v2/notices``` -- URL where messages will be posted. Default: ```https://api.airbrake.io/notifier_api/v2/notices```
  * ```-rate-limit 25``` -- limit of message per minute to send. Default ```25```.
  * ```-queue-len 1000``` -- size of internal message queue. Messages not fitting into the queue will be discarded. Default: ```1000```.
