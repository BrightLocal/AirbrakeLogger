package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"queue"
	"sender"
	"server"
)

var (
	queueServer      = flag.String("beanstalk", "", "Address:port of Beanstalkd server")
	tcpServerAddress = flag.String("listen", "", "Address:port to listen for messages")
	queueName        = flag.String("queue", "Airbrake", "Queue name to listen")
	url              = flag.String("url", "https://api.airbrake.io/notifier_api/v2/notices", "Url to post log messages")
	rateLimit        = flag.Int("rate-limit", 25, "Upper limit messages per minute to send")
	queueLength      = flag.Int("queue-len", 1000, "Length of message queue")
)

func init() {
	flag.Parse()
	if *queueServer == "" && *tcpServerAddress == "" {
		log.Fatal("Use at least one messages source (use beanstalkd or listen flags)!")
	}
	messageQueue := make(chan []byte, *queueLength)
	sender.New(*url, *rateLimit, messageQueue)
	if *queueServer != "" {
		queue.New(*queueServer, *queueName, messageQueue)
	}
	if *tcpServerAddress != "" {
		server.New(*tcpServerAddress, messageQueue)
		log.Printf("Listening on %s", *tcpServerAddress)
	}
	log.Printf("Queue length is %d", *queueLength)
	log.Printf("Rate limit is %d/min", *rateLimit)
}

func main() {
	var osSignalsChannel chan os.Signal = make(chan os.Signal)
	signal.Notify(osSignalsChannel, os.Interrupt, os.Kill)
	<-osSignalsChannel
	log.Print("Exiting")
}
