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
	conn         *beanstalk.Conn
	tubes        *beanstalk.TubeSet
}

func New(address, queueName string, messageQueue chan []byte) *Queue {
	q := &Queue{
		address,
		queueName,
		messageQueue,
		nil,
		nil,
	}
	go q.run()
	return q
}

func (q *Queue) connect() {
	for {
		var err error
		q.conn, err = beanstalk.Dial("tcp", q.address)
		if err != nil {
			log.Printf("Could not connect to queue server %s: %v", q.address, err)
			time.Sleep(5 * time.Second)
			continue
		}
		log.Printf("Connected to queue server at %s", q.address)
		return
	}
}

func (q *Queue) subscribe() {
	q.tubes = &beanstalk.TubeSet{q.conn, make(map[string]bool)}
	q.tubes.Name["default"] = false
	q.tubes.Name[q.name] = true
}

func (q *Queue) run() {
	log.Print("Starting beanstalkd queue watcher")
	q.connect()
	defer func() {
		q.conn.Close()
	}()
	q.subscribe()
	for {
		id, body, err := q.tubes.Reserve(1)
		if err == nil {
			q.conn.Delete(id)
			// Silently discard messages if chan is full
			if len(q.messageQueue) < cap(q.messageQueue) {
				q.messageQueue <- body
			}
		} else if err.Error() != "reserve-with-timeout: timeout" {
			log.Printf("Reconnecting to queue server due to %v", err)
			q.conn.Close()
			time.Sleep(5 * time.Second)
			q.connect()
			q.subscribe()
		} else {
			time.Sleep(50 * time.Millisecond)
		}
	}
}
