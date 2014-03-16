package rx

import (
	"errors"
	"fmt"
	"reflect"
)

type Signal interface {
	// questo può avere tipo ...interface{} e con reflect vedere i tipi dei parametri
	// un func(error) sarà per error, un func() per complete e un altro func(qualcosa) per next
	Subscribe(...interface{}) <-chan bool

	Concat(Signal) Signal
}

type signal struct {
	Signal
	didSubscribe func(Subscriber)
}

func (s *signal) Subscribe(params ...interface{}) <-chan bool {
	var nextFunc func(interface{})
	var errFunc func(error)
	var compFunc func()
	var extDisposable Disposable
	var subscriber Subscriber
	for _, p := range params {
		switch p.(type) {
		case func(error):
			if errFunc != nil {
				panic("Error function already defined")
			}
			errFunc = p.(func(error))
		case func():
			if compFunc != nil {
				panic("Completion function already defined")
			}
			compFunc = p.(func())
		case func(interface{}):
			if nextFunc != nil {
				panic("'Next' function already defined")
			}
			nextFunc = p.(func(interface{}))
		case Disposable:
			if extDisposable != nil {
				panic("Disposable already defined")
			}
			extDisposable = p.(Disposable)
		case Subscriber:
			if subscriber != nil {
				panic("Subscriber already defined")
			}
			subscriber = p.(Subscriber)
		default:
			if nextFunc != nil {
				panic("'Next' function already defined")
			}
			nextFuncT := reflect.TypeOf(p)
			if nextFuncT.Kind() != reflect.Func || nextFuncT.NumIn() != 1 {
				panic("Invalid 'next' function")
			}
			nextFuncV := reflect.ValueOf(p)
			nextArgT := nextFuncT.In(0)
			nextFunc = func(v interface{}) {
				vV := reflect.ValueOf(v)
				if vV.Type().AssignableTo(nextArgT) {
					nextFuncV.Call([]reflect.Value{vV})
				} else if errFunc != nil {
					errFunc(errors.New(fmt.Sprintf("Expect type %v got %v", nextArgT.Name(), vV.Type().Name())))
				}
			}
		}
	}

	if extDisposable != nil && extDisposable.IsDisposed() {
		im := make(chan bool, 1)
		im <- true
		return im
	}

	if subscriber == nil {
		subscriber = newSubstriber(nextFunc, errFunc, compFunc)
	}

	if extDisposable != nil {
		extDisposable.AddCallback(subscriber.Disposable().Dispose)
	}

	go s.didSubscribe(subscriber)

	return subscriber.Disposable().DispositionChan()
}

func NewSignal(didSubscribe func(subscriber Subscriber)) Signal {
	return &signal{didSubscribe: didSubscribe}
}

func (s *signal) Concat(signal Signal) Signal {
	return NewSignal(func(sub Subscriber) {
		s.Subscribe(func(v interface{}) {
			sub.SendNext(v)
		}, func(err error) {
			sub.SendError(err)
		}, func() {
			signal.Subscribe(sub)
		}, sub.Disposable())
	})
}
