package rx

import (
	"testing"
	"time"
)

func TestItShouldCallCallback(t *testing.T) {
	testChan := make(chan int, 2)
	d := NewDisposable(func() {
		testChan <- 1
	})
	<-d.Dispose()
	if countChan(testChan) != 1 {
		t.Error("Expect callback to be called on disposal")
	}
}

func TestItShouldCallMultipleCallbacks(t *testing.T) {
	testChan := make(chan int, 2)
	d := NewDisposable(func() {
		testChan <- 1
	})
	d.AddCallback(func() {
		testChan <- 2
	})
	<-d.Dispose()
	calledCallbacks := chanToSlice(testChan)
	if len(calledCallbacks) != 2 || calledCallbacks[0] != 1 || calledCallbacks[1] != 2 {
		t.Errorf("Expect all callbacks to have be called on disposal, got: %v", calledCallbacks)
	}
}

func TestItShouldCallCallbacksOnlyOnce(t *testing.T) {
	testChan := make(chan int, 2)
	d := NewDisposable(func() {
		testChan <- 1
	})
	d.Dispose()
	<-d.Dispose()
	if countChan(testChan) != 1 {
		t.Error("Expect callback to be called only once")
	}
}

func TestItShouldAllowWaitingMultipleTimesOnDispositionChan(t *testing.T) {
	testChan := make(chan int, 3)
	d := NewDisposable(nil)
	go func() {
		<-d.Dispose()
		testChan <- 1
	}()
	go func() {
		<-d.DispositionChan()
		testChan <- 1
	}()
	<-d.Dispose()
	go func() {
		<-d.DispositionChan()
		testChan <- 1
	}()
	returnedFunc := 0
	for {
		select {
		case <-testChan:
			returnedFunc += 1
			if returnedFunc == 3 {
				goto end
			}
		case <-time.After(1 * time.Second):
			goto end
		}
	}
end:
	if returnedFunc != 3 {
		t.Error("Expect disposable to signal all go routines")
	}
}
