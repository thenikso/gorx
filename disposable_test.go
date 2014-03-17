package rx

import (
	"testing"
	"time"
)

func TestItShouldDisposeAsyncronously(t *testing.T) {
	d := NewDisposable(nil)
	if d.IsDisposed() == true {
		t.Error("Expect Disposable not to be disposed")
	}
	d.Dispose()
	if d.IsDisposed() == true {
		t.Error("Expect Disposable not to be disposed before DispositionChan triggers")
	}
	<-d.DispositionChan()
	if d.IsDisposed() == false {
		t.Error("Expect Disposable to be disposed")
	}
}

func TestItShouldCallCallback(t *testing.T) {
	callbackCalled := 0
	d := NewDisposable(func() {
		callbackCalled += 1
	})
	d.Dispose()
	<-d.DispositionChan()
	if callbackCalled != 1 {
		t.Error("Expect callback to be called on disposal")
	}
}

func TestItShouldCallMultipleCallbacks(t *testing.T) {
	calledCallbacks := []int{}
	d := NewDisposable(func() {
		calledCallbacks = append(calledCallbacks, 1)
	})
	d.AddCallback(func() {
		calledCallbacks = append(calledCallbacks, 2)
	})
	d.Dispose()
	<-d.DispositionChan()
	if len(calledCallbacks) != 2 || calledCallbacks[0] != 1 || calledCallbacks[1] != 2 {
		t.Errorf("Expect all callbacks to have be called on disposal, got: %v", calledCallbacks)
	}
}

func TestAddDispositionChan(t *testing.T) {
	d := NewDisposable(nil)
	chan1 := make(chan bool, 1)
	doneChan := make(chan bool, 1)
	receivedChanDefault := false
	receivedChan1 := false
	d.AddDispositionChan(chan1)
	d.Dispose()
	func() {
		for {
			select {
			case receivedChanDefault = <-d.DispositionChan():
				if receivedChan1 {
					doneChan <- true
					return
				}
			case receivedChan1 = <-chan1:
				if receivedChanDefault {
					doneChan <- true
					return
				}
			case <-time.After(1 * time.Second):
				t.Log("Timeout!")
				doneChan <- true
				return
			}
		}
	}()
	<-doneChan
	if !receivedChanDefault {
		t.Error("Expect default disposition channel to trigger")
	}
	if !receivedChan1 {
		t.Error("Expect chan1 disposition channel to trigger")
	}
}
