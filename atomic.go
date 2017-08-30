package main

import "sync"

type Atomic interface {
	Value() interface{}
	SetValue(interface{})
	Modify(func(interface{}) interface{}) interface{}
	ModifyData(func(interface{}) (interface{}, interface{})) (interface{}, interface{})
	Swap(interface{}) interface{}
	WithValue(func(interface{}) interface{}) interface{}
}

type _atomic struct {
	value interface{}
	mutex sync.Mutex
}

func NewAtomic(v interface{}) Atomic {
	return &_atomic{v, sync.Mutex{}}
}

// Atomically gets the value of the variable.
func (atomic *_atomic) Value() interface{} {
	atomic.mutex.Lock()
	v := atomic.value
	atomic.mutex.Unlock()
	return v
}

// Atomically sets the value of the variable.
func (atomic *_atomic) SetValue(v interface{}) {
	atomic.mutex.Lock()
	atomic.value = v
	atomic.mutex.Unlock()
}

// Atomically modifies the variable.
//
// Returns the old value, plus arbitrary user-defined data.
func (atomic *_atomic) ModifyData(action func(interface{}) (interface{}, interface{})) (interface{}, interface{}) {
	atomic.mutex.Lock()
	oldValue := atomic.value
	newValue, data := action(atomic.value)
	atomic.value = newValue
	atomic.mutex.Unlock()
	return oldValue, data
}

// Atomically modifies the variable.
//
// Returns the old value.
func (atomic *_atomic) Modify(action func(interface{}) interface{}) interface{} {
	oldValue, _ := atomic.ModifyData(func(oldValue interface{}) (interface{}, interface{}) {
		return action(oldValue), 0
	})
	return oldValue
}

// Atomically replaces the contents of the variable.
//
// Returns the old value.
func (atomic *_atomic) Swap(newValue interface{}) interface{} {
	return atomic.Modify(func(_ interface{}) interface{} {
		return newValue
	})
}

// Atomically performs an arbitrary action using the current value of the
// variable.
//
// Returns the result of the action.
func (atomic *_atomic) WithValue(action func(interface{}) interface{}) interface{} {
	atomic.mutex.Lock()
	result := action(atomic.value)
	atomic.mutex.Unlock()
	return result
}
