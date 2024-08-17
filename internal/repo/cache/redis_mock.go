package cache

import (
	"context"

	"github.com/gomodule/redigo/redis"
)

var (
	GetFunc        func() redis.Conn
	GetContextFunc func(ctx context.Context) (redis.Conn, error)
	CloseFunc      func() error
	ErrFunc        func() error
	DoFunc         func(commandName string, args ...interface{}) (reply interface{}, err error)
	SendFunc       func(commandName string, args ...interface{}) error
	FlushFunc      func() error
	ReceiveFunc    func() (reply interface{}, err error)
)

type ConnMock struct{}

func (ConnMock) Close() error {
	return CloseFunc()
}

func (ConnMock) Err() error {
	return ErrFunc()
}

func (ConnMock) Do(commandName string, args ...interface{}) (reply interface{}, err error) {
	return DoFunc(commandName, args...)
}

func (ConnMock) Send(commandName string, args ...interface{}) error {
	return SendFunc(commandName, args...)
}

func (ConnMock) Flush() error {
	return FlushFunc()
}

func (ConnMock) Receive() (reply interface{}, err error) {
	return ReceiveFunc()
}

type HandlerMock struct{}

func (h HandlerMock) Get() redis.Conn {
	return GetFunc()
}

func (h HandlerMock) GetContext(ctx context.Context) (redis.Conn, error) {
	return GetContextFunc(ctx)
}
