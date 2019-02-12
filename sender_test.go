package main

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/messagebird/go-rest-api/sms"
)

type createCall struct {
	originator string
	recipients []string
	body       string
}

func compareMessages(t *testing.T, expected Message, got createCall) {
	if expected.Originator != got.originator {
		t.Fatalf(
			"Wrong originator. Expected - %s, got - %s",
			expected.Originator, got.originator,
		)
	}

	if len(got.recipients) != 1 {
		t.Fatalf("Expected to have 1 recipient, got - %d", len(got.recipients))
	}

	if expected.Recipient != got.recipients[0] {
		t.Fatalf(
			"Wrong recipient. Expected - %s, got - %s",
			expected.Recipient, got.recipients[0],
		)
	}

	if expected.Body != got.body {
		t.Fatalf(
			"Wrong body. Expected - %s, got - %s",
			expected.Body, got.body,
		)
	}
}

func TestSender(t *testing.T) {
	createCalls := []createCall{}
	clientMock := messageBirdMock{
		CreateFunc: func(originator string, recipients []string, body string, msgParams *sms.Params) (*sms.Message, error) {
			call := createCall{
				originator: originator,
				recipients: recipients,
				body:       body,
			}
			createCalls = append(createCalls, call)
			return new(sms.Message), nil
		},
	}

	sender := NewSender(10, 1*time.Second, &clientMock)
	defer sender.Close()

	t.Run("positive - single message", func(t *testing.T) {
		defer func() {
			createCalls = []createCall{}
		}()

		msg := Message{
			Originator: "mockedOriginator",
			Recipient:  "mockedRecipient",
			Body:       "mockedBody",
		}

		sender.Enque(msg)

		time.Sleep(1*time.Second + 500*time.Millisecond)

		if len(sender.ch) != 0 {
			t.Fatalf("Expected to have empty sender's queue. Got - %d", len(sender.ch))
		}

		if len(createCalls) != 1 {
			t.Fatalf("Expected to have 1 message, got - %d", len(createCalls))
		}

		compareMessages(t, msg, createCalls[0])
	})

	t.Run("positive - several big messages", func(t *testing.T) {
		defer func() {
			createCalls = []createCall{}
		}()

		msgs := make([]Message, 5)
		for i := 0; i < len(msgs); i++ {
			msg := Message{
				Originator: fmt.Sprintf("mockedOriginator-%d", i),
				Recipient:  fmt.Sprintf("mockedRecipient-%d", i),
				Body:       strings.Repeat("*", 160),
			}
			msgs[i] = msg

			sender.Enque(msg)
		}

		time.Sleep(time.Duration(len(msgs))*time.Second + 500*time.Millisecond)

		if len(sender.ch) != 0 {
			t.Fatalf("Expected to have empty sender's queue. Got - %d", len(sender.ch))
		}

		if len(createCalls) != len(msgs) {
			t.Fatalf(
				"Expected to have %d messages, got - %d",
				len(msgs), len(createCalls),
			)
		}

		for i, msg := range msgs {
			compareMessages(t, msg, createCalls[i])
		}
	})

	t.Run("positive - several small messages merged", func(t *testing.T) {
		msg := Message{
			Originator: "mockedOriginator",
			Recipient:  "mockedRecipient",
			Body:       strings.Repeat("*", 50),
		}

		for i := 0; i < 2; i++ {
			sender.Enque(msg)
		}

		// wait for period
		time.Sleep(2*time.Second + 500*time.Millisecond)

		// enqueue one more message
		anotherMsg := msg
		anotherMsg.Body = "just another message"

		sender.Enque(anotherMsg)

		time.Sleep(1*time.Second + 500*time.Millisecond)

		if len(sender.ch) != 0 {
			t.Fatalf("Expected to have empty sender's queue. Got - %d", len(sender.ch))
		}

		// we expected body to be merged
		msg.Body = strings.Join([]string{msg.Body, msg.Body}, "\n")
		// we expect only 2 messages, since first 2 should've been merged
		expectedMessages := []Message{msg, anotherMsg}

		if len(createCalls) != len(expectedMessages) {
			t.Fatalf(
				"Expected to have %d messages, got - %d",
				len(expectedMessages), len(createCalls),
			)
		}

		for i, msg := range expectedMessages {
			compareMessages(t, msg, createCalls[i])
		}
	})
}
