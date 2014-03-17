package rx

import "time"

type Disposable interface {
	Dispose()
	DispositionChan() <-chan bool
	AddDispositionChan(chan<- bool)
	IsDisposed() bool
	AddCallback(func())
}

type disposable struct {
	Disposable

	disposed  bool
	callbacks []func()

	dispositionChan  chan bool
	dispositionChans []chan<- bool

	operationChan chan func() bool
}

func (d *disposable) Dispose() {
	d.operationChan <- func() bool {
		d.disposed = true
		for _, c := range d.dispositionChans {
			go func(c chan<- bool) {
				select {
				case c <- true:
				case <-time.After(1 * time.Second):
				}
			}(c)
		}
		for _, callback := range d.callbacks {
			callback()
		}
		d.callbacks = nil
		d.dispositionChans = nil
		d.operationChan = nil
		return true
	}
}

func (d *disposable) DispositionChan() <-chan bool {
	return d.dispositionChan
}

func (d *disposable) AddDispositionChan(c chan<- bool) {
	if d.operationChan == nil {
		select {
		case c <- true:
		default:
		}
		return
	}
	d.operationChan <- func() bool {
		if d.disposed {
			select {
			case c <- true:
			default:
			}
		} else {
			d.dispositionChans = append(d.dispositionChans, c)
		}
		return false
	}
}

func (d *disposable) IsDisposed() bool {
	return d.disposed
}

func (d *disposable) AddCallback(callback func()) {
	if d.operationChan == nil {
		callback()
		return
	}
	d.operationChan <- func() bool {
		if d.disposed {
			callback()
		} else {
			d.callbacks = append(d.callbacks, callback)
		}
		return false
	}
}

func NewDisposable(callback func()) Disposable {
	dispositionChan := make(chan bool, 1)
	d := &disposable{
		disposed:  false,
		callbacks: make([]func(), 0, 1),

		dispositionChan:  dispositionChan,
		dispositionChans: []chan<- bool{dispositionChan},

		operationChan: make(chan func() bool, 10),
	}
	if callback != nil {
		d.callbacks = append(d.callbacks, callback)
	}
	go func() {
		for {
			op := <-d.operationChan
			if op() {
				return
			}
		}
	}()
	return d
}
