package main

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

func TestFilterShouldFilterSignal(t *testing.T) {
	signal := NewValuesSignal([]interface{}{1, 2, 3, 4, 5, 6})
	result := make([]int, 0)
	expected := []int{2, 4, 6}

	signal.FilterAuto(func(v int) bool {
		return v%2 == 0
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

func TestScan(t *testing.T) {
	signal := NewValuesSignal([]interface{}{1, 2, 3, 4, 5, 6})
	result := make([]int, 0)
	expected := []int{101, 103, 106, 110, 115, 121}

	signal.ScanAuto(100, func(state int, current int) int {
		return state + current
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

func TestReduce(t *testing.T) {
	signal := NewValuesSignal([]interface{}{1, 2, 3, 4, 5, 6})
	result := make([]int, 0, 1)
	expected := []int{21}

	signal.ReduceAuto(0, func(acc int, curr int) int {
		return acc + curr
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

func TestTake(t *testing.T) {
	signal := NewValuesSignal([]interface{}{1, 2, 3, 4, 5, 6})
	result := make([]int, 0)
	expected := []int{1, 2, 3, 4}

	signal.Take(4).SubscribeAuto(func(v int) {
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

func TestTakeLast(t *testing.T) {
	signal := NewValuesSignal([]interface{}{1, 2, 3, 4, 5, 6})
	result := make([]int, 0)
	expected := []int{3, 4, 5, 6}

	signal.TakeLast(4).SubscribeAuto(func(v int) {
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

func TestTakeWhile(t *testing.T) {
	signal := NewValuesSignal([]interface{}{1, 2, 3, 4, 5, 6})
	result := make([]int, 0)
	expected := []int{1, 2, 3, 4}

	signal.TakeWhileAuto(func(v int) bool {
		return v < 5
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

func TestConcatWith(t *testing.T) {
	result := make([]int, 0)
	expected := []int{1, 2, 3, 4, 5, 6}

	NewValuesSignal([]interface{}{1, 2}).
		ConcatWith(NewValuesSignal([]interface{}{3, 4})).
		ConcatWith(NewValuesSignal([]interface{}{5, 6})).
		SubscribeAuto(func(v int) {
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
