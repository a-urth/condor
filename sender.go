package main

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"time"
)

const (
	msgSizeLimit = 160
)

type Message struct {
	Originator string `json:"originator"`
	Recipient  string `json:"recipient"`
	Body       string `json:"body"`
}

func (m Message) Validate() error {
	verify := &Verifier{}

	verify.That(len(m.Recipient) > 0, "recipient cannot be empty")
	verify.That(len(m.Originator) > 0, "originator cannot be empty")
	verify.That(len(m.Body) > 0, "body cannot be empty")
	verify.That(
		len(m.Body) <= msgSizeLimit,
		fmt.Sprintf("body cannot be larger than %d bytes", msgSizeLimit),
	)

	return verify.err
}

func (m Message) Key() string {
	return m.Recipient + m.Originator
}

func (m Message) String() string {
	return fmt.Sprintf(
		"Message <Recipient: %q, Originator: %q, Body: %q>",
		m.Recipient, m.Originator, m.Body,
	)
}

type Sender struct {
	ch                chan string
	ticker            *time.Ticker
	finish            chan interface{}
	messages          *messageStore
	messageBirdClient messageBirder
}

type messageStore struct {
	sync.RWMutex
	store map[string][]*Message
}

func NewSender(limit int, period time.Duration, client messageBirder) *Sender {
	tick := time.NewTicker(period)

	s := Sender{
		ch:     make(chan string, limit),
		ticker: tick,
		finish: make(chan interface{}),
		messages: &messageStore{
			sync.RWMutex{},
			map[string][]*Message{},
		},
		messageBirdClient: client,
	}

	go s.send()

	return &s
}

func (s *Sender) Close() {
	log.Println("Stopping sender...")
	s.finish <- nil
}

func (s *Sender) Enque(msg Message) {
	key := msg.Key()

	s.messages.Lock()
	defer s.messages.Unlock()

	if msgs, ok := s.messages.store[key]; ok {
		s.messages.store[key] = append(msgs, &msg)
	} else {
		s.messages.store[key] = []*Message{&msg}
		s.ch <- key
	}
}

func (s *Sender) send() {
	for {
		select {
		case <-s.ticker.C:
			// if there are no messages in the queue yet
			if len(s.ch) == 0 {
				continue
			}

			if key, ok := <-s.ch; ok {
				s.messages.Lock()

				msgs, ok := s.messages.store[key]
				if !ok {
					panic("Message received but no data in store")
				}

				totalLen := 0
				bodies := make([]string, 0, len(msgs))
				// there will be anyway one message less
				newMsgs := make([]*Message, 0, len(msgs)-1)

				// try to merge messages from same originator to same recipient
				// in one sms
				for _, msg := range msgs {
					totalLen += len(msg.Body)
					if totalLen <= msgSizeLimit {
						bodies = append(bodies, msg.Body)
					} else {
						newMsgs = append(newMsgs, msg)
					}
				}

				if len(newMsgs) > 0 {
					s.messages.store[key] = newMsgs
				} else {
					delete(s.messages.store, key)
				}

				s.messages.Unlock()

				// since originator and recipient is same for all messages
				// we can take them from first one
				_, err := s.messageBirdClient.Create(
					msgs[0].Originator,
					[]string{msgs[0].Recipient},
					strings.Join(bodies, "\n"),
					nil,
				)
				// dummy fall back, do nothing on error
				if err != nil {
					log.Printf("Error sending message - %s", err.Error())
				}
			}
		case <-s.finish:
			return
		}
	}
}
