package main

import "errors"

type Verifier struct {
	err error
}

func (v *Verifier) That(condition bool, msg string) *Verifier {
	if v.err != nil {
		return v
	}

	if !condition {
		v.err = errors.New(msg)
	}

	return v
}
