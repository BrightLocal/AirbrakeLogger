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

var (
	messageQueue    chan []byte
	messageSender   *sender.Sender
	tcpServer       *server.Server
	beanstalkClient *queue.Queue
)

func init() {
	flag.Parse()
	messageQueue = make(chan []byte, *queueLength)
	log.Printf("Queue length is %d", *queueLength)
	messageSender = sender.New(*url, *rateLimit, messageQueue)
	if *queueServer != "" {
		beanstalkClient = queue.New(*queueServer, *queueName, messageQueue)
	}
	if *tcpServerAddress != "" {
		tcpServer = server.New(*tcpServerAddress, messageQueue)
		log.Printf("Listening on %s", *tcpServerAddress)
	}
}

func main() {
	var osSignalsChannel chan os.Signal = make(chan os.Signal)
	signal.Notify(osSignalsChannel, os.Interrupt, os.Kill)
	<-osSignalsChannel
	log.Print("Exiting")
}
