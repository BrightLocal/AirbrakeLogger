package sender

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"time"
)

type Sender struct {
	url            string
	messageQueue   chan []byte
	delay          time.Duration
	messageCounter uint
	errorsCounter  uint
	startTime      time.Time
}

func New(url string, rate int, messageQueue chan []byte) *Sender {
	s := &Sender{
		url:            url,
		messageQueue:   messageQueue,
		delay:          600 / time.Duration(rate) * time.Second / 10, // One tenth of second resolution
		messageCounter: 0,
		errorsCounter:  0,
		startTime:      time.Now(),
	}
	go s.sender()
	go s.countLogger()
	return s
}

func (s *Sender) sender() {
	log.Print("Starting sender")
	for {
		s.send(<-s.messageQueue)
		time.Sleep(s.delay)
	}
}

func (s *Sender) send(xml []byte) {
	s.messageCounter++
	var msg string
	err := json.Unmarshal(xml, &msg)
	if err != nil {
		log.Printf("Error parsing xml: %s", err)
		log.Printf("Erroneous payload: %s", xml)
		return
	}
	resp, err := http.Post(s.url, "text/xml", bytes.NewBuffer([]byte(msg)))
	if err != nil {
		log.Printf("Error sending post request: %s", err)
		s.errorsCounter++
	} else {
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			log.Printf("Message not accepted: %s", resp.Status)
			s.errorsCounter++
			return
		}
	}
}

func (s *Sender) countLogger() {
	for range time.Tick(1 * time.Minute) {
		log.Printf(
			"Messages: %d (%0.4f/min). Errors: %d (%0.4f/min).",
			s.messageCounter,
			float64(s.messageCounter)/time.Now().Sub(s.startTime).Minutes(),
			s.errorsCounter,
			float64(s.errorsCounter)/time.Now().Sub(s.startTime).Minutes(),
		)
	}
}
