package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

type Server struct {
	s              *http.Server
	senderCallback func(msg Message)
}

func NewServer(port int, senderCallback func(msg Message)) *Server {
	s := &http.Server{
		Addr: fmt.Sprintf(":%d", port),
	}

	return &Server{
		s:              s,
		senderCallback: senderCallback,
	}
}

func (s *Server) Start() {
	http.HandleFunc("/send", func(w http.ResponseWriter, r *http.Request) {
		s.sendMessage(w, r)
	})

	go func() {
		if err := s.s.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalln(err)
		}
	}()

	log.Printf("Listening to requests on %s", s.s.Addr)
}

func (s *Server) Close(ctx context.Context) {
	log.Println("Closing server...")

	if err := s.s.Shutdown(ctx); err != nil {
		log.Fatalln(err)
	}
}

func (s *Server) sendMessage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		log.Printf("Wrong method - %s", r.Method)
		http.Error(w, "wrong method for uri", http.StatusBadRequest)
		return
	}

	body, readErr := ioutil.ReadAll(r.Body)
	if readErr != nil {
		log.Printf("Error reading request - %v", readErr)
		http.Error(w, "error reading request", http.StatusBadRequest)
		return
	}

	msg := Message{}
	if marshalErr := json.Unmarshal(body, &msg); marshalErr != nil {
		log.Printf("Error parsing request - %v", marshalErr)
		http.Error(w, "error parsing request", http.StatusBadRequest)
		return
	}

	if validationErr := msg.Validate(); validationErr != nil {
		log.Printf("Error validating request - %v", validationErr)
		http.Error(w, validationErr.Error(), http.StatusBadRequest)
		return
	}

	s.senderCallback(msg)
}
