package main

import (
	"errors"
	"fmt"
	"reflect"
)

type T interface{}
type U interface{}

// A stream that will begin generating events when a Subscriber is attached,
// possibly performing some side effects in the process. Events are pushed to
// the subscriber as they are generated.
//
// A corollary to this is that different Subscribers may see a different timing
// of events, or even a different version of events altogether.
type Signal interface {
	Subscribe(Subscriber) Disposable
	SubscribeFunc(func(T), func(error), func()) Disposable
	SubscribeAuto(...interface{}) Disposable

	Map(func(T) U) Signal
	MapAuto(interface{}) Signal

	Filter(func(T) bool) Signal
	FilterAuto(interface{}) Signal

	Scan(U, func(U, T) U) Signal
	ScanAuto(U, interface{}) Signal

	Reduce(U, func(U, T) U) Signal
	ReduceAuto(U, interface{}) Signal

	Take(int) Signal
	TakeLast(int) Signal
	TakeWhile(func(T) bool) Signal
	TakeWhileAuto(interface{}) Signal

	Merge() Signal

	Concat() Signal
	ConcatWith(s Signal) Signal
}

type signal struct {
	didSubscribe func(Subscriber)
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

// Starts producing events for the given subscriber.
//
// Returns a Disposable which will cancel the work associated with event
// production, and prevent any further events from being sent.
func (signal *signal) Subscribe(subscriber Subscriber) Disposable {
	signal.didSubscribe(subscriber)
	return subscriber.Disposable()
}

func (signal *signal) SubscribeFunc(next func(T), err func(error), completed func()) Disposable {
	return signal.Subscribe(NewSubscriber(next, err, completed))
}

func (signal *signal) SubscribeAuto(params ...interface{}) Disposable {
	var nextFunc func(T)
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
			nextFunc = p.(func(T))
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
			nextFunc = func(v T) {
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

	return signal.Subscribe(subscriber)
}

// Maps each value in the stream to a new value.
func (signal *signal) Map(f func(T) U) Signal {
	return signal.mapAccumulate(struct{}{}, func(_ interface{}, value T) (interface{}, U) {
		return struct{}{}, f(value)
	})
}

func (signal *signal) MapAuto(p interface{}) Signal {
	mapFunc, err := castFunc(p, (func(T) U)(nil))
	if err != nil {
		panic(err)
	}
	return signal.Map(mapFunc.(func(T) U))
}

// Preserves only the values of the signal that pass the given predicate.
func (signal *signal) Filter(predicate func(T) bool) Signal {
	return signal.Map(func(value T) U {
		if predicate(value) {
			return NewSingleSignal(value)
		} else {
			return NewEmptySignal()
		}
	}).Merge()
}

func (signal *signal) FilterAuto(p interface{}) Signal {
	predicateFunc, err := castFunc(p, (func(T) bool)(nil))
	if err != nil {
		panic(err)
	}
	return signal.Filter(predicateFunc.(func(T) bool))
}

// Combines all the values in the stream, forwarding the result of each
// intermediate combination step.
func (signal *signal) Scan(initial U, f func(U, T) U) Signal {
	return signal.mapAccumulate(initial.(interface{}), func(previous interface{}, current T) (interface{}, U) {
		mapped := f(previous, current)
		return mapped, mapped
	})
}

func (signal *signal) ScanAuto(initial U, p interface{}) Signal {
	f, err := castFunc(p, (func(U, T) U)(nil))
	if err != nil {
		panic(err)
	}
	return signal.Scan(initial, f.(func(U, T) U))
}

// Combines all of the values in the stream.
//
// Returns a signal which will send the single, aggregated value when
// the receiver completes.
func (signal *signal) Reduce(initial U, f func(U, T) U) Signal {
	scanned := signal.Scan(initial, f)
	return NewSingleSignal(initial).ConcatWith(scanned).TakeLast(1)
}

func (signal *signal) ReduceAuto(initial U, p interface{}) Signal {
	f, err := castFunc(p, (func(U, T) U)(nil))
	if err != nil {
		panic(err)
	}
	return signal.Reduce(initial, f.(func(U, T) U))
}

/// Returns a signal that will yield the first `count` values from the
/// receiver.
func (signal *signal) Take(count int) Signal {
	if count < 0 {
		panic("Signal.Take: count parameter should be >= 0")
	}

	if count == 0 {
		return NewEmptySignal()
	}

	return signal.mapAccumulate(0, func(n interface{}, value T) (interface{}, U) {
		var newN interface{}
		if n.(int)+1 < count {
			newN = n.(int) + 1
		}
		return newN, value
	})
}

/// Waits for the receiver to complete successfully, then forwards only the
/// last `count` values.
func (signal *signal) TakeLast(count int) Signal {
	if count < 0 {
		panic("Signal.TakeLast: count parameter should be >= 0")
	}

	if count == 0 {
		return signal.Filter(func(_ T) bool {
			return false
		})
	}

	return NewSignal(func(subscriber Subscriber) {
		values := NewAtomic(make([]T, 0))
		disposable := signal.SubscribeFunc(
			func(value T) {
				values.Modify(func(a interface{}) interface{} {
					arr := a.([]T)
					arr = append(arr, value)

					if len(arr) > count {
						return arr[(len(arr) - count):]
					}

					return arr
				})
			},
			func(err error) {
				subscriber.OnError(err)
			},
			func() {
				for _, v := range values.Value().([]T) {
					subscriber.OnNext(v)
				}

				subscriber.OnCompleted()
			},
		)

		subscriber.Disposable().AddDisposable(disposable)
	})
}

// Returns a signal that will yield values from the receiver while
// `predicate` remains `true`.
func (signal *signal) TakeWhile(predicate func(T) bool) Signal {
	return signal.mapAccumulate(true, func(taking interface{}, value T) (interface{}, U) {
		if taking.(bool) && predicate(value) {
			return true, NewSingleSignal(value)
		} else {
			return nil, NewEmptySignal()
		}
	}).Merge()
}

func (signal *signal) TakeWhileAuto(p interface{}) Signal {
	f, err := castFunc(p, (func(T) bool)(nil))
	if err != nil {
		panic(err)
	}
	return signal.TakeWhile(f.(func(T) bool))
}

// Merges a signal of signals down into a single signal, biased toward the
// signals added earlier.
//
// Returns a signal that will forward events from the original signals
// as they arrive.
func (signal *signal) Merge() Signal {
	return NewSignal(func(subscriber Subscriber) {
		disposable := NewCompositeDisposable(nil)
		inFlight := NewAtomic(1)

		decrementInFlight := func() {
			orig := inFlight.Modify(func(v interface{}) interface{} {
				return v.(int) - 1
			})
			if orig == 1 {
				subscriber.OnCompleted()
			}
		}

		selfDisposable := signal.SubscribeFunc(
			func(stream T) {
				inFlight.Modify(func(v interface{}) interface{} {
					return v.(int) + 1
				})

				streamDisposable := NewSerialDisposable(nil)
				disposable.AddDisposable(streamDisposable)

				streamDisposable.SetInnerDisposable(stream.(Signal).SubscribeFunc(
					func(value T) {
						subscriber.OnNext(value.(interface{}))
					},
					func(err error) {
						streamDisposable.Dispose()
						disposable.PruneDisposed()
						subscriber.OnError(err)
					},
					func() {
						streamDisposable.Dispose()
						disposable.PruneDisposed()
						decrementInFlight()
					},
				))
			},
			func(err error) {
				subscriber.OnError(err)
			},
			func() {
				decrementInFlight()
			},
		)

		subscriber.Disposable().AddDisposable(selfDisposable)
	})
}

// Concatenates each inner signal with the previous and next inner signals.
//
// Returns a signal that will forward events from each of the original
// signals, in sequential order.
func (signal *signal) Concat() Signal {
	return NewSignal(func(subscriber Subscriber) {
		// TODO fix implementation to not rely on channel size
		// also check if this all makes sense
		concatChan := make(chan Signal, 10)
		var currentSignal Signal
		var subscribeToNextSignal func()
		subscribeToNextSignal = func() {
			if currentSignal != nil {
				return
			}
			select {
			case currentSignal, more := <-concatChan:
				if !more {
					subscriber.OnCompleted()
					return
				}
				signalDisposable := currentSignal.SubscribeFunc(
					func(v T) {
						subscriber.OnNext(v)
					},
					func(err error) {
						subscriber.OnError(err)
					},
					func() {
						subscribeToNextSignal()
					},
				)
				subscriber.Disposable().AddDisposable(signalDisposable)
			default:
			}
		}

		selfDisposable := signal.SubscribeFunc(
			func(signal T) {
				fmt.Println("1")
				concatChan <- signal.(Signal)
				subscribeToNextSignal()
			},
			func(err error) {
				subscriber.OnError(err)
			},
			func() {
				fmt.Println("e")
				close(concatChan)
			},
		)

		subscriber.Disposable().AddDisposable(selfDisposable)
	})
}

/// Concatenates the given signal after the receiver.
func (signal *signal) ConcatWith(s Signal) Signal {
	return NewValuesSignal([]interface{}{signal, s}).Concat()
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
func (s *signal) mapAccumulate(initialState interface{}, f func(state interface{}, current T) (newState interface{}, newValue U)) Signal {
	return NewSignal(func(subscriber Subscriber) {
		state := NewAtomic(initialState)
		disposable := s.SubscribeFunc(
			// Next
			func(value T) {
				newState, newValue := f(state.Value(), value)
				subscriber.OnNext(newValue)

				if newState != nil {
					state.SetValue(newState)
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
