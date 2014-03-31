package rx

// import (
// 	"testing"
// )

// func TestSignal(t *testing.T) {
// 	var expectedValue = "1"
// 	var signal = NewSignal(func(subscriber Subscriber) {
// 		subscriber.SendNext(expectedValue)
// 		subscriber.SendCompleted()
// 	})

// 	var expectedValue2 = "2"
// 	var signal2 = NewSignal(func(subscriber Subscriber) {
// 		subscriber.SendNext(expectedValue2)
// 		subscriber.SendCompleted()
// 	})

// 	var receivedValue string
// 	<-signal.Concat(signal2).Subscribe(func(value string) {
// 		receivedValue = value
// 	})

// 	if expectedValue2 != receivedValue {
// 		t.Errorf("Expect '%v' got '%v'", expectedValue, receivedValue)
// 	}
// }
