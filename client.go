package main

import (
	messagebird "github.com/messagebird/go-rest-api"
	"github.com/messagebird/go-rest-api/sms"
)

type messageBirder interface {
	Create(originator string, recipients []string, body string, msgParams *sms.Params) (*sms.Message, error)
}

type messageBirdClientWrapper struct {
	client *messagebird.Client
}

func (m *messageBirdClientWrapper) Create(
	originator string,
	recipients []string,
	body string,
	msgParams *sms.Params,
) (*sms.Message, error) {
	return sms.Create(m.client, originator, recipients, body, msgParams)
}

type messageBirdMock struct {
	CreateFunc func(originator string, recipients []string, body string, msgParams *sms.Params) (*sms.Message, error)
}

func (m *messageBirdMock) Create(
	originator string,
	recipients []string,
	body string,
	msgParams *sms.Params,
) (*sms.Message, error) {
	return m.CreateFunc(originator, recipients, body, msgParams)
}
