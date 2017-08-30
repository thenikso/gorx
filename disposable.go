package main

// Represents something that can be “disposed”, usually associated with freeing
// resources or canceling work.
type Disposable interface {
	Dispose() error
	IsDisposed() bool
}

// A disposable that only flips `disposed` upon disposal, and performs no other
// work.
type simpleDisposable struct {
	disposed Atomic
}

func (disposable *simpleDisposable) IsDisposed() bool {
	return disposable.disposed.Value().(bool)
}

func (disposable *simpleDisposable) Dispose() error {
	disposable.disposed.SetValue(true)
	return nil
}

func NewSimpleDisposable() Disposable {
	disposable := &simpleDisposable{disposed: NewAtomic(false)}
	return disposable
}

// A disposable that will run an action upon disposal.
type actionDisposable struct {
	action Atomic
}

func (disposable *actionDisposable) IsDisposed() bool {
	return disposable.action.Value() == nil
}

func (disposable *actionDisposable) Dispose() error {
	oldAction := disposable.action.Swap(nil)
	var err error
	if oldAction != nil {
		err = oldAction.(func() error)()
	}
	return err
}

func NewActionDisposable(action func() error) Disposable {
	disposable := &actionDisposable{action: NewAtomic(action)}
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
	disposables Atomic
}

func (disposable *compositeDisposable) IsDisposed() bool {
	return disposable.disposables.Value() == nil
}

func (disposable *compositeDisposable) Dispose() error {
	ds := disposable.disposables.Swap(nil)
	var err error
	if ds != nil {
		for _, d := range ds.([]Disposable) {
			err = d.Dispose()
		}
	}
	return err
}

func (disposable *compositeDisposable) AddDisposable(d Disposable) error {
	if d == nil {
		return nil
	}

	_, shouldDispose := disposable.disposables.ModifyData(func(ds interface{}) (interface{}, interface{}) {
		if ds != nil {
			return append(ds.([]Disposable), d), false
		} else {
			return nil, true
		}
	})

	if shouldDispose == true {
		return d.Dispose()
	}

	return nil
}

func (disposable *compositeDisposable) AddDisposableFunc(action func() error) error {
	if action == nil {
		return nil
	}
	return disposable.AddDisposable(NewActionDisposable(action))
}

func (disposable *compositeDisposable) PruneDisposed() {
	disposable.disposables.Modify(func(ds interface{}) interface{} {
		filteredDisposables := make([]Disposable, 0, len(ds.([]Disposable)))
		for _, d := range ds.([]Disposable) {
			if d.IsDisposed() == false {
				filteredDisposables = append(filteredDisposables, d)
			}
		}
		return filteredDisposables
	})
}

func NewCompositeDisposable(action func() error) CompositeDisposable {
	disposable := &compositeDisposable{disposables: NewAtomic(make([]Disposable, 0, 1))}
	disposable.AddDisposableFunc(action)
	return disposable
}

// A disposable that will optionally dispose of another disposable.
type SerialDisposable interface {
	Disposable
	InnerDisposable() Disposable
	SetInnerDisposable(Disposable)
}

type serialDisposableState struct {
	innerDisposable Disposable
	disposed        bool
}

type serialDisposable struct {
	state Atomic
}

func (disposable *serialDisposable) IsDisposed() bool {
	return disposable.state.Value().(serialDisposableState).disposed
}

func (disposable *serialDisposable) Dispose() error {
	orig := disposable.state.Swap(serialDisposableState{
		innerDisposable: nil,
		disposed:        true,
	})
	if d := orig.(serialDisposableState).innerDisposable; d != nil {
		return d.Dispose()
	}
	return nil
}

func (disposable *serialDisposable) InnerDisposable() Disposable {
	return disposable.state.Value().(serialDisposableState).innerDisposable
}

func (disposable *serialDisposable) SetInnerDisposable(inner Disposable) {
	oldState := disposable.state.Modify(func(state interface{}) interface{} {
		return serialDisposableState{
			innerDisposable: inner,
			disposed:        state.(serialDisposableState).disposed,
		}
	})

	if d := oldState.(serialDisposableState).innerDisposable; d != nil {
		d.Dispose()
	}
	if oldState.(serialDisposableState).disposed {
		inner.Dispose()
	}
}

func NewSerialDisposable(innerDisposable Disposable) SerialDisposable {
	disposable := &serialDisposable{
		NewAtomic(serialDisposableState{
			innerDisposable: innerDisposable,
			disposed:        false,
		})}
	return disposable
}
