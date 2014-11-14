package rx

import "testing"

func TestShouldReadAndWriteDirectly(t *testing.T) {
	atomic := NewAtomic(1)
	if atomic.Value() != 1 {
		t.Errorf("Expect `atomic.Value()` to equal 1, got %v", atomic.Value())
	}

	atomic.SetValue(2)
	if atomic.Value() != 2 {
		t.Errorf("Expect `atomic.Value()` to equal 2, got %v", atomic.Value())
	}
}

func TestShouldSwapAtomically(t *testing.T) {
	atomic := NewAtomic(1)
	if atomic.Swap(2) != 1 {
		t.Errorf("Expect `atomic.Swap(2)` to equal 1, got %v", atomic.Swap(2))
	}
	if atomic.Value() != 2 {
		t.Errorf("Expect `atomic.Value()` to equal 2, got %v", atomic.Value())
	}
}

func TestShouldModifyAtomically(t *testing.T) {
	atomic := NewAtomic(1)
	if atomic.Modify(func(v interface{}) interface{} { return v.(int) + 1 }) != 1 {
		t.Errorf("Expect `atomic.Modify(func(v interface{}) interface{} { return v.(int) + 1 })` to equal 1, got %v", atomic.Modify(func(v interface{}) interface{} { return v.(int) + 1 }))
	}
	if atomic.Value() != 2 {
		t.Errorf("Expect `atomic.Value()` to equal 2, got %v", atomic.Value())
	}
}

func TestShouldModifyAndReturnSomeData(t *testing.T) {
	atomic := NewAtomic(1)
	orig, data := atomic.ModifyData(func(oldValue interface{}) (interface{}, interface{}) {
		return oldValue.(int) + 1, "foobar"
	})
	if orig != 1 {
		t.Errorf("Expect `orig` to equal 1, got %v", orig)
	}
	if data != "foobar" {
		t.Errorf("Expect `data` to equal \"foobar\", got %v", data)
	}
	if atomic.Value() != 2 {
		t.Errorf("Expect `atomic.Value()` to equal 2, got %v", atomic.Value())
	}
}

func TestShouldPerformAnAction(t *testing.T) {
	atomic := NewAtomic(1)
	result := atomic.WithValue(func(value interface{}) interface{} {
		return value.(int) == 1
	})
	if result != true {
		t.Errorf("Expect `result` to equal true, got %v", result)
	}
	if atomic.Value() != 1 {
		t.Errorf("Expect `atomic.Value()` to equal 1, got %v", atomic.Value())
	}
}
