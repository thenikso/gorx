package rx

import (
	"testing"
	"time"
)

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
