package rx

import (
	"errors"
	"fmt"
	"reflect"
	"sync"
)

type Signal interface {
	Subscribe(...interface{}) Disposable

	MapI(func(interface{}) interface{}) Signal
	Map(interface{}) Signal
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

// Maps over the elements of the signal, accumulating a state along the
// way.
//
// This is meant as a primitive operator from which more complex operators
// can be built.
//
// Yielding a `nil` state at any point will stop evaluation of the original
// signal, and dispose of it.
//
// Returns a signal of the mapped values.
func (s *signal) mapAccumulate(initialState interface{}, f func(state interface{}, current interface{}) (newState interface{}, newValue interface{})) Signal {
	return NewSignal(func(subscriber Subscriber) {
		var mutex sync.Mutex
		state := initialState
		disposable := s.Subscribe(
			// Next
			func(value interface{}) {
				mutex.Lock()
				st := state
				mutex.Unlock()
				newState, newValue := f(st, value)
				subscriber.OnNext(newValue)

				if newState != nil {
					mutex.Lock()
					state = newState
					mutex.Unlock()
				} else {
					subscriber.OnCompleted()
				}
			},
			// Error
			func(err error) {
				subscriber.OnError(err)
			},
			// Completed
			func() {
				subscriber.OnCompleted()
			},
		)
		subscriber.Disposable().AddDisposable(disposable)
	})
}

// Maps each value in the stream to a new value.
func (signal *signal) MapI(f func(interface{}) interface{}) Signal {
	return signal.mapAccumulate(struct{}{}, func(_, value interface{}) (interface{}, interface{}) {
		return struct{}{}, f(value)
	})
}
func (signal *signal) Map(p interface{}) Signal {
	mapFunc, err := castFunc(p, (func(interface{}) interface{})(nil))
	if err != nil {
		panic(err)
	}
	return signal.MapI(mapFunc.(func(interface{}) interface{}))
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

// Utility function to convert a function signature.
// The number of inputs and outputs between the input function and the desired
// signature must be the same.
// If the usual Go conversion rules do not allow conversion of inputs and outputs,
// castFunc panics.
// Typical usage (without error management):
//     f, _ := castFunc(aFunc, (func(interface{}) interface{})(nil))
//     f.(func(interface{}) interface{})(param)
func castFunc(p interface{}, to interface{}) (interface{}, error) {
	toT := reflect.TypeOf(to)
	pT := reflect.TypeOf(p)
	if pT.Kind() != reflect.Func {
		return nil, errors.New(fmt.Sprintf("Invalid parameter kind (%v) expecting function", pT.Kind()))
	}
	if pT.Kind() != toT.Kind() {
		return nil, errors.New(fmt.Sprintf("Invalid parameter kind (%v) expecting %v", pT.Kind(), toT.Kind()))
	}
	if pT.NumIn() != toT.NumIn() {
		return nil, errors.New(fmt.Sprintf("Invalid parameter inputs number (%v) expecting %v", pT.NumIn(), toT.NumIn()))
	}
	if pT.NumOut() != toT.NumOut() {
		return nil, errors.New(fmt.Sprintf("Invalid parameter outputs number (%v) expecting %v", pT.NumOut(), toT.NumOut()))
	}
	for i := 0; i < pT.NumIn(); i++ {
		if pT.In(i).AssignableTo(toT.In(i)) == false {
			return nil, errors.New(fmt.Sprintf("Invalid function %v input (%v) not assignable as %v", i, pT.In(i), toT.In(i)))
		}
	}
	funcValue := reflect.MakeFunc(toT, func(args []reflect.Value) []reflect.Value {
		properArgs := make([]reflect.Value, 0, len(args))
		for _, a := range args {
			properArgs = append(properArgs, reflect.ValueOf(a.Interface()))
		}
		results := reflect.ValueOf(p).Call(properArgs)
		properResults := make([]reflect.Value, 0, len(results))
		for i, a := range results {
			properResults = append(properResults, a.Convert(toT.Out(i)))
		}
		return properResults
	})
	return funcValue.Interface(), nil
}
