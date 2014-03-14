package rx

import (
	"testing"
)

func slicesEquals(a []interface{}, b []interface{}) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if b[i] != v {
			return false
		}
	}
	return true
}

func verifyValues(t *testing.T, signal *Signal, expectedValues []interface{}) {
	if signal == nil {
		t.Error("Expect signal to be non nil")
	}

	var collectedValues = make([]interface{}, 0, len(expectedValues))
	var success = false
	var err error = nil

	signal.Subscribe(func(value interface{}) {
		collectedValues = append(collectedValues, value)
	}, func(e error) {
		err = e
	}, func() {
		success = true
	})

	if success == false {
		t.Error("Expect signal to complete")
	}
	if err != nil {
		t.Error("Expect signal not to return an error")
	}
	if !slicesEquals(collectedValues, expectedValues) {
		t.Errorf("Expect signal to return %#v got %#v", expectedValues, collectedValues)
	}
}

func TestEmptyStream(t *testing.T) {
	var stream = new(Signal)
	verifyValues(t, stream, make([]interface{}, 0))
}
