package rx

import (
	"sync"
	"sync/atomic"
)

// Represents something that can be “disposed”, usually associated with freeing
// resources or canceling work.
type Disposable interface {
	Dispose() error
	IsDisposed() bool
}

// A disposable that only flips `disposed` upon disposal, and performs no other
// work.
type simpleDisposable struct {
	disposed uint32
}

func (disposable *simpleDisposable) IsDisposed() bool {
	val := atomic.LoadUint32(&disposable.disposed)
	return val == 1
}

func (disposable *simpleDisposable) Dispose() error {
	atomic.StoreUint32(&disposable.disposed, 1)
	return nil
}

func NewSimpleDisposable() Disposable {
	disposable := &simpleDisposable{disposed: 0}
	return disposable
}

// A disposable that will run an action upon disposal.
type actionDisposable struct {
	action func() error
	mutex  sync.Mutex
}

func (disposable *actionDisposable) IsDisposed() bool {
	disposable.mutex.Lock()
	disposed := disposable.action == nil
	disposable.mutex.Unlock()
	return disposed
}

func (disposable *actionDisposable) Dispose() error {
	disposable.mutex.Lock()
	var err error
	if disposable.action != nil {
		err = disposable.action()
	} else {
		err = nil
	}
	disposable.action = nil
	disposable.mutex.Unlock()
	return err
}

func NewActionDisposable(action func() error) Disposable {
	disposable := &actionDisposable{action: action}
	return disposable
}

// A disposable that will dispose of any number of other disposables.
type CompositeDisposable interface {
	Disposable
	AddDisposable(Disposable) error
	AddDisposableFunc(func() error) error
	PruneDisposed()
}

type compositeDisposable struct {
	disposables []Disposable
	mutex       sync.Mutex
}

func (disposable *compositeDisposable) IsDisposed() bool {
	disposable.mutex.Lock()
	disposed := disposable.disposables == nil
	disposable.mutex.Unlock()
	return disposed
}

func (disposable *compositeDisposable) Dispose() error {
	disposable.mutex.Lock()
	var err error
	for _, d := range disposable.disposables {
		err = d.Dispose()
	}
	disposable.disposables = nil
	disposable.mutex.Unlock()
	return err
}

func (disposable *compositeDisposable) AddDisposable(d Disposable) error {
	if d == nil {
		return nil
	}

	shouldDispose := false
	disposable.mutex.Lock()
	if disposable.disposables != nil {
		disposable.disposables = append(disposable.disposables, d)
	} else {
		shouldDispose = true
	}
	disposable.mutex.Unlock()

	if shouldDispose == true {
		return d.Dispose()
	}

	return nil
}

func (disposable *compositeDisposable) AddDisposableFunc(action func() error) error {
	return disposable.AddDisposable(NewActionDisposable(action))
}

func (disposable *compositeDisposable) PruneDisposed() {
	disposable.mutex.Lock()
	filteredDisposables := make([]Disposable, 0, len(disposable.disposables))
	for _, d := range disposable.disposables {
		if d.IsDisposed() == false {
			filteredDisposables = append(filteredDisposables, d)
		}
	}
	disposable.disposables = filteredDisposables
	disposable.mutex.Unlock()
}

func NewCompositeDisposable() CompositeDisposable {
	disposable := &compositeDisposable{disposables: make([]Disposable, 0, 1)}
	return disposable
}
