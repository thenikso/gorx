package rx

import (
	"errors"
	"fmt"
	"reflect"
)

type Signal interface {
	Subscribe(...interface{}) Disposable
}

type signal struct {
	didSubscribe func(Subscriber)
}

// Starts producing events for the given subscriber.
//
// Returns a Disposable which will cancel the work associated with event
// production, and prevent any further events from being sent.
func (s *signal) Subscribe(params ...interface{}) Disposable {
	var nextFunc func(interface{})
	var errFunc func(error)
	var compFunc func()
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

	if subscriber == nil {
		subscriber = NewSubscriber(nextFunc, errFunc, compFunc)
	}

	s.didSubscribe(subscriber)

	return subscriber.Disposable()
}

// Creates a signal that will execute the given action upon subscription,
// then forward all events from the generated signal.
func NewSignal(didSubscribe func(subscriber Subscriber)) Signal {
	return &signal{didSubscribe: didSubscribe}
}

// Creates a signal that will immediately complete.
func NewEmptySignal() Signal {
	return &signal{func(subscriber Subscriber) {
		subscriber.OnCompleted()
	}}
}

// Creates a signal that will immediately yield a single value then
// complete.
func NewSingleSignal(value interface{}) Signal {
	return &signal{func(subscriber Subscriber) {
		subscriber.OnNext(value)
		subscriber.OnCompleted()
	}}
}

// Creates a signal that will immediately generate an error.
func NewErrorSignal(err error) Signal {
	return &signal{func(subscriber Subscriber) {
		subscriber.OnError(err)
	}}
}

// Creates a signal that will never send any events.
func NewNeverSignal() Signal {
	return &signal{func(_ Subscriber) {
	}}
}

// Creates a signal that will iterate over the given sequence whenever a
// Subscriber is attached.
func NewValuesSignal(values []interface{}) Signal {
	return &signal{func(subscriber Subscriber) {
		for _, v := range values {
			subscriber.OnNext(v)
		}
		subscriber.OnCompleted()
	}}
}
