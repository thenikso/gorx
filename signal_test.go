package rx

import (
	"testing"
)

func TestMapShouldMapInput(t *testing.T) {
	signal := NewValuesSignal([]interface{}{1, 2, 3})
	result := make([]int, 0)
	signal.Map(func(v int) int {
		return v + 1
	}).Subscribe(func(v int) {
		result = append(result, v)
	})
	if len(result) != 3 {
		t.Errorf("Expecting `len(result)` to equal 3 got %v", len(result))
	}
	expected := []int{2, 3, 4}
	for i, v := range result {
		if v != expected[i] {
			t.Fatalf("Expecting %v to equal %v", result, expected)
		}
	}
}
