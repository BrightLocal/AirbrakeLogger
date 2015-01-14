package sender

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
)

type Sender struct {
	url          string
	messageQueue chan []byte
}

func New(url string, messageQueue chan []byte) *Sender {
	s := &Sender{
		url,
		messageQueue,
	}
	go s.sender()
	return s
}

func (s *Sender) sender() {
	for {
		s.send(<-s.messageQueue)
	}
}

func (s *Sender) send(xml []byte) {
	var msg string
	err := json.Unmarshal(xml, &msg)
	if err != nil {
		log.Printf("%s", err)
		return
	}
	resp, err := http.Post(s.url, "text/xml", bytes.NewBuffer([]byte(msg)))
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
