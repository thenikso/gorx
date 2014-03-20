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

	disposed  chan bool
	callbacks []func()

	dispositionChan  chan bool
	dispositionChans []chan<- bool

	operationChan chan func(bool) bool
}

func pingDispositionChannel(c chan<- bool) {
	select {
	case c <- true:
	case <-time.After(1 * time.Second):
	}
}

func (d *disposable) doWithDisposedState(f func(bool)) {
	disp := <-d.disposed
	f(disp)
	d.disposed <- disp
}

func (d *disposable) Dispose() {
	d.doWithDisposedState(func(disposed bool) {
		d.operationChan <- func(bool) bool {
			for _, c := range d.dispositionChans {
				go pingDispositionChannel(c)
			}
			for _, callback := range d.callbacks {
				callback()
			}
			return true
		}
	})
}

func (d *disposable) DispositionChan() <-chan bool {
	return d.dispositionChan
}

func (d *disposable) AddDispositionChan(c chan<- bool) {
	d.doWithDisposedState(func(disposed bool) {
		if disposed {
			go pingDispositionChannel(c)
			return
		}
		d.operationChan <- func(disposed bool) bool {
			if disposed {
				pingDispositionChannel(c)
			} else {
				d.dispositionChans = append(d.dispositionChans, c)
			}
			return false
		}
	})
}

func (d *disposable) IsDisposed() bool {
	var disp bool
	d.doWithDisposedState(func(disposed bool) {
		disp = disposed
	})
	return disp
}

func (d *disposable) AddCallback(callback func()) {
	d.doWithDisposedState(func(disposed bool) {
		if disposed {
			callback()
			return
		}
		d.operationChan <- func(disposed bool) bool {
			if disposed {
				callback()
			} else {
				d.callbacks = append(d.callbacks, callback)
			}
			return false
		}
	})
}

func NewDisposable(callback func()) Disposable {
	dispositionChan := make(chan bool, 1)
	d := &disposable{
		disposed:  make(chan bool, 1),
		callbacks: make([]func(), 0, 1),

		dispositionChan:  dispositionChan,
		dispositionChans: []chan<- bool{dispositionChan},

		operationChan: make(chan func(bool) bool, 10),
	}
	if callback != nil {
		d.callbacks = append(d.callbacks, callback)
	}
	d.disposed <- false
	go func() {
		for {
			op := <-d.operationChan
			disp := <-d.disposed
			if op(disp) {
				d.disposed <- true
				for {
					select {
					case op = <-d.operationChan:
						<-d.disposed
						op(true)
						d.disposed <- true
					default:
						return
					}
				}
			}
			d.disposed <- false
		}
	}()
	return d
}
