package queue

import (
	"github.com/kr/beanstalk"
	"log"
	"time"
)

type Queue struct {
	address      string
	name         string
	messageQueue chan []byte
}

func New(address, queueName string, messageQueue chan []byte) *Queue {
	q := &Queue{
		address,
		queueName,
		messageQueue,
	}
	go q.run()
	return q
}

func (q *Queue) connect() *beanstalk.Conn {
	for {
		beanstalk, err := beanstalk.Dial("tcp", q.address)
		if err != nil {
			log.Printf("Could not connect to queue server %s: %v", q.address, err)
			time.Sleep(1 * time.Second)
			continue
		}
		log.Printf("Connected to queue server at %s", q.address)
		return beanstalk
	}
}

func (q *Queue) run() {
	var Queue *beanstalk.Conn
	Queue = q.connect()
	defer Queue.Close()
	tube := &beanstalk.TubeSet{Queue, make(map[string]bool)}
	tube.Name["default"] = false
	tube.Name[q.name] = true
	for {
		id, body, err := tube.Reserve(10)
		if err == nil {
			Queue.Delete(id)
			// Silently discard messages if chan is full
			if len(q.messageQueue) < cap(q.messageQueue) {
				q.messageQueue <- body
			}
		}
	}
}
