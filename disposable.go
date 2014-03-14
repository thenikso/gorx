package rx

type Disposable interface {
	Dispose()
}

type CompoundDisposable struct {
	disposeCallbacks []func()
}

func (d *CompoundDisposable) Dispose() {
	for _, callback := range d.disposeCallbacks {
		callback()
	}
}

func (d *CompoundDisposable) AddDisposeCallback(callback func()) {
	if d.disposeCallbacks == nil {
		d.disposeCallbacks = []func(){callback}
	} else {
		d.disposeCallbacks = append(d.disposeCallbacks, callback)
	}
}

func NewCompoundDisposable(callback func()) *CompoundDisposable {
	var disposable = new(CompoundDisposable)
	disposable.AddDisposeCallback(callback)
	return disposable
}
