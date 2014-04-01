package rx

import (
	"testing"
	"time"
)

type testSubscriber struct {
	subscriber Subscriber

	receivedValuesChan chan []interface{}
	receivedErrorChan  chan error
	successChan        chan bool

	testing *testing.T
}

func newTestSubscriber(t *testing.T) testSubscriber {
	test := testSubscriber{
		receivedValuesChan: make(chan []interface{}, 1),
		receivedErrorChan:  make(chan error, 1),
		successChan:        make(chan bool, 1),
		testing:            t,
	}
	test.receivedValuesChan <- []interface{}{}
	test.receivedErrorChan <- nil
	test.successChan <- true
	test.subscriber = NewSubscriber(func(v interface{}) {
		values := <-test.receivedValuesChan
		values = append(values, v)
		test.receivedValuesChan <- values
		test.testing.Logf("Received next: %v", v)
	}, func(err error) {
		<-test.receivedErrorChan
		<-test.successChan
		test.receivedErrorChan <- err
		test.successChan <- false
		test.testing.Logf("Received error: %v", err)
	}, func() {
		<-test.successChan
		test.successChan <- true
		test.testing.Log("Received completed")
	})
	return test
}

func (t testSubscriber) expectReceiving(values []interface{}) <-chan []interface{} {
	resultChan := make(chan []interface{}, 1)
	go func() {
		rec := []interface{}{}
		for i := 0; i < 10; i++ {
			rec = <-t.receivedValuesChan
			t.receivedValuesChan <- rec
			if len(values) == len(rec) {
				for i, v := range rec {
					if values[i] != v {
						goto tryagain
					}
				}
				resultChan <- nil
				return
			}
		tryagain:
			time.Sleep(10 * time.Millisecond)
		}
		resultChan <- rec
	}()
	return resultChan
}

func TestItShouldCallNext(t *testing.T) {
	testChan := make(chan int, 1)
	s := NewSubscriber(func(val interface{}) {
		testChan <- val.(int)
	}, nil, nil)
	expectedValue := 2
	s.SendNext(expectedValue)
	select {
	case receivedValue := <-testChan:
		if receivedValue != expectedValue {
			t.Error("Expect 'next' callback to receive sent value")
		}
	case <-time.After(1 * time.Second):
		t.Error("Timout!")
	}
}

func TestItShouldCallError(t *testing.T) {
	testChan := make(chan int, 1)
	expectedValue := 2
	s := NewSubscriber(nil, func(err error) {
		testChan <- expectedValue
	}, nil)
	s.SendError(nil)
	<-s.Disposable().DispositionChan()
	select {
	case receivedValue := <-testChan:
		if receivedValue != expectedValue {
			t.Error("Expect 'complete' callback to be called")
		}
	case <-time.After(1 * time.Second):
		t.Error("Timout!")
	}
}

func TestItShouldCallCompleted(t *testing.T) {
	testChan := make(chan int, 1)
	expectedValue := 2
	s := NewSubscriber(nil, nil, func() {
		testChan <- expectedValue
	})
	s.SendCompleted()
	<-s.Disposable().DispositionChan()
	select {
	case receivedValue := <-testChan:
		if receivedValue != expectedValue {
			t.Error("Expect 'complete' callback to be called")
		}
	case <-time.After(1 * time.Second):
		t.Error("Timout!")
	}
}

func TestItShouldSendNextSerially(t *testing.T) {
	test := newTestSubscriber(t)
	go func() {
		for i := 0; i < 5; i++ {
			test.subscriber.SendNext(interface{}(i))
		}
	}()

	expectedValues := []interface{}{0, 1, 2, 3, 4}
	if receivedValues := <-test.expectReceiving(expectedValues); receivedValues != nil {
		t.Errorf("Expecting %v, received %v", expectedValues, receivedValues)
	}
}
