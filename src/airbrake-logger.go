package main

import (
	"flag"
	"github.com/kr/beanstalk"
	"log"
	"os"
	"os/signal"
	"sender"
	"server"
	"time"
)

var (
	queueServer      = flag.String("beanstalk", "0.0.0.0:11300", "Address:port of Beanstalkd server")
	tcpServerAddress = flag.String("listen", "", "Address:port to listen for messages")
	queueName        = flag.String("queue", "Airbrake", "Queue name to listen")
	url              = flag.String("url", "https://api.airbrake.io/notifier_api/v2/notices", "Url to post log messages")
	rateLimit        = flag.Int("rate-limit", 25, "Upper limit messages per minute to send")
	queueLength      = flag.Int("queue-len", 10000, "Length of message queue")
	sleepDelay       time.Duration
	messageQueue     chan []byte
	messageSender    *sender.Sender
	tcpServer        *server.Server
)

func init() {
	flag.Parse()
	sleepDelay = 600 / time.Duration(*rateLimit) * time.Second / 10 // One tenth of second resolution
	messageQueue = make(chan []byte, *queueLength)
	messageSender = sender.New(*url, messageQueue)
	if *tcpServerAddress != "" {
		tcpServer = server.New(*tcpServerAddress, messageQueue)
	}
}

func main() {
	var osSignalsChannel chan os.Signal
	osSignalsChannel = make(chan os.Signal, 1)
	signal.Notify(osSignalsChannel, os.Interrupt, os.Kill)
	var Queue *beanstalk.Conn
	Queue = queueConnect()
	defer Queue.Close()
	tube := &beanstalk.TubeSet{Queue, make(map[string]bool)}
	tube.Name["default"] = false
	tube.Name[*queueName] = true
	for {
		id, body, err := tube.Reserve(10)
		if err == nil {
			log.Printf("Got log entry %d", id)
			Queue.Delete(id)
			messageQueue <- body
		}
		select {
		case <-osSignalsChannel:
			log.Print("Exiting")
			return
		default:
			time.Sleep(sleepDelay)
		}
	}
}

func queueConnect() *beanstalk.Conn {
	for {
		log.Print("Connecting to queue server...")
		beanstalk, err := beanstalk.Dial("tcp", *queueServer)
		if err != nil {
			log.Printf("Could not connect: %v", err)
			time.Sleep(1 * time.Second)
			continue
		}
		log.Printf("Connected")
		return beanstalk
	}
}
