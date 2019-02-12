package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strings"
	"testing"
)

func failOnError(t *testing.T, err error) {
	if err != nil {
		t.Fatalf("Unexpected error - %s", err.Error())
	}
}

func testRequest(
	t *testing.T,
	uri, method, expectedBody string,
	expectedStatus int,
	reqBody io.Reader,
) {
	client := &http.Client{}

	req, err := http.NewRequest(method, uri, reqBody)
	failOnError(t, err)

	req.Header.Add("Content-Type", "application/json")
	resp, err := client.Do(req)
	failOnError(t, err)
	defer resp.Body.Close()

	if resp.StatusCode != expectedStatus {
		t.Fatalf("Wrong status. Expected - %d, got - %d", expectedStatus, resp.StatusCode)
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	failOnError(t, err)

	if string(respBody) != expectedBody {
		t.Fatalf("Wrong body. Expected - %s, got - %s", expectedBody, respBody)
	}
}

func TestServer(t *testing.T) {
	msgs := []Message{}
	callback := func(msg Message) {
		msgs = append(msgs, msg)
	}

	port := 8000 + rand.Intn(1000)
	uri := fmt.Sprintf("http://localhost:%d/send", port)
	s := NewServer(port, callback)
	s.Start()
	defer s.Close(context.TODO())

	t.Run("negative - bad method", func(t *testing.T) {
		testRequest(t, uri, "GET", "wrong method for uri\n", http.StatusBadRequest, nil)
		if len(msgs) != 0 {
			t.Fatalf("Expected to have no messages. Got - %d", len(msgs))
		}
	})

	t.Run("negative - bad json", func(t *testing.T) {
		body := strings.NewReader("not a json")
		testRequest(t, uri, "POST", "error parsing request\n", http.StatusBadRequest, body)
		if len(msgs) != 0 {
			t.Fatalf("Expected to have no messages. Got - %d", len(msgs))
		}
	})

	t.Run("negative - validation error - empty recipient", func(t *testing.T) {
		body := strings.NewReader(`{"originator": "some", "body": "some"}`)
		testRequest(t, uri, "POST", "recipient cannot be empty\n", http.StatusBadRequest, body)
		if len(msgs) != 0 {
			t.Fatalf("Expected to have no messages. Got - %d", len(msgs))
		}
	})

	t.Run("negative - validation error - empty originator", func(t *testing.T) {
		body := strings.NewReader(`{"recipient": "some", "body": "some"}`)
		testRequest(t, uri, "POST", "originator cannot be empty\n", http.StatusBadRequest, body)
		if len(msgs) != 0 {
			t.Fatalf("Expected to have no messages. Got - %d", len(msgs))
		}
	})

	t.Run("negative - validation error - empty body", func(t *testing.T) {
		body := strings.NewReader(`{"originator": "some", "recipient": "some"}`)
		testRequest(t, uri, "POST", "body cannot be empty\n", http.StatusBadRequest, body)
		if len(msgs) != 0 {
			t.Fatalf("Expected to have no messages. Got - %d", len(msgs))
		}
	})

	t.Run("positive", func(t *testing.T) {
		msg := Message{
			Recipient:  "some dude",
			Originator: "some other dude",
			Body:       "body",
		}
		body, err := json.Marshal(msg)
		failOnError(t, err)

		testRequest(t, uri, "POST", "", http.StatusOK, bytes.NewReader(body))
		if len(msgs) != 1 {
			t.Fatalf("Expected to have 1 message. Got - %d", len(msgs))
		}

		if msg != msgs[0] {
			t.Fatalf("Expected to have msg - %v, got - %v", msg, msgs[0])
		}
	})
}
