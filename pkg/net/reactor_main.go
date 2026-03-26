package net

import "btaskee-quiz/pkg/log"

type ReactorMain struct {
	acceptor *acceptor
	eg       *Engine
	log.Log
}

func NewReactorMain(eg *Engine) *ReactorMain {

	return &ReactorMain{
		acceptor: newAcceptor(eg),
		eg:       eg,
		Log:      log.NewBLog("ReactorMain"),
	}
}

func (m *ReactorMain) Start() error {
	return m.acceptor.Start()
}

func (m *ReactorMain) Stop() error {
	return m.acceptor.Stop()
}
