package rx

import (
	"testing"
)

func TestMapShouldMapInput(t *testing.T) {
	signal := NewValuesSignal([]interface{}{1, 2, 3})
	result := make([]int, 0)
	expected := []int{2, 3, 4}

	signal.MapAuto(func(v int) int {
		return v + 1
	}).SubscribeAuto(func(v int) {
		result = append(result, v)
	})

	if len(result) != len(expected) {
		t.Fatalf("Expecting `len(result)` to equal %v got %v", len(expected), len(result))
	}
	for i, v := range expected {
		if v != result[i] {
			t.Fatalf("Expecting %v to equal %v", result, expected)
		}
	}
}

func TestMergeShouldMergeSignals(t *testing.T) {
	signal := NewValuesSignal([]interface{}{
		NewValuesSignal([]interface{}{1, 2}),
		NewValuesSignal([]interface{}{3, 4}),
		NewValuesSignal([]interface{}{5, 6}),
	})
	result := make([]int, 0)
	expected := []int{1, 2, 3, 4, 5, 6}

	signal.Merge().SubscribeAuto(func(v int) {
		result = append(result, v)
	})

	if len(result) != len(expected) {
		t.Fatalf("Expecting `len(result)` to equal %v got %v", len(expected), len(result))
	}
	for i, v := range expected {
		if v != result[i] {
			t.Fatalf("Expecting %v to equal %v", result, expected)
		}
	}
}
