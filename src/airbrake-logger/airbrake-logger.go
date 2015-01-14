package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"github.com/kr/beanstalk"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
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
	tcpServer        *server
)

func init() {
	flag.Parse()
	sleepDelay = 600 / time.Duration(*rateLimit) * time.Second / 10 // One tenth of second resolution
	messageQueue = make(chan []byte, *queueLength)
	go sender(messageQueue)
	if *tcpServerAddress != "" {
		tcpServer = &server{}
		tcpServer.runListener(*tcpServerAddress)
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

func sender(c chan []byte) {
	for {
		send(<-c)
	}
}

func send(xml []byte) {
	var msg string
	err := json.Unmarshal(xml, &msg)
	if err != nil {
		log.Printf("%s", err)
		return
	}
	resp, err := http.Post(*url, "text/xml", bytes.NewBuffer([]byte(msg)))
	if err != nil {
		log.Printf("Error sending post request: %s", err)
	} else {
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			log.Printf("Message not accepted: %s", resp.Status)
			return
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

type server struct {
}

func (s *server) runListener(address string) error {
	socket, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}
	go s.listener(socket)
	return nil
}

func (s *server) listener(socket net.Listener) {
	for {
		connection, err := socket.Accept()
		if err != nil {
			log.Print(err)
		}
		go s.handler(connection)
	}
}

func (s *server) handler(connection net.Conn) {
	defer connection.Close()
	msg, err := bufio.NewReader(connection).ReadBytes(0)
	if err != nil {
		log.Print(err)
	}
	messageQueue <- msg[:len(msg)-1]
}
