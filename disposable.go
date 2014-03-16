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
	canDisposeChan   chan bool
	dispositionChan  chan bool
	dispositionChans []chan<- bool
	disposed         bool
	callbacks        []func()
}

func (d *disposable) Dispose() {
	go func() {
		select {
		case <-d.canDisposeChan:
		default:
		}
	}()
}

func (d *disposable) disposeImpl() {
	d.canDisposeChan <- false
	if d.disposed == true {
		return
	}
	for _, c := range d.dispositionChans {
		go func(c chan<- bool) {
			select {
			case c <- true:
			case <-time.After(1 * time.Second):
			}
		}(c)
	}
	d.disposed = true
	for _, callback := range d.callbacks {
		callback()
	}
	d.callbacks = nil
	d.dispositionChan = nil
	d.dispositionChans = nil
	d.canDisposeChan = nil
}

func (d *disposable) DispositionChan() <-chan bool {
	return d.dispositionChan
}

func (d *disposable) AddDispositionChan(c chan<- bool) {
	d.dispositionChans = append(d.dispositionChans, c)
}

func (d *disposable) IsDisposed() bool {
	return d.disposed
}

func (d *disposable) AddCallback(callback func()) {
	d.callbacks = append(d.callbacks, callback)
}

func NewDisposable(callback func()) Disposable {
	dispositionChan := make(chan bool)
	d := &disposable{
		canDisposeChan:   make(chan bool, 1),
		dispositionChan:  dispositionChan,
		dispositionChans: []chan<- bool{dispositionChan},
		disposed:         false,
		callbacks:        make([]func(), 0, 1),
	}
	if callback != nil {
		d.AddCallback(callback)
	}
	d.canDisposeChan <- true
	go d.disposeImpl()
	return d
}
