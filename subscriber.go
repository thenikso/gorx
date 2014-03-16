package rx

type Subscriber interface {
	SendNext(interface{})
	SendError(error)
	SendCompleted()
	Disposable() Disposable
}

type subscriber struct {
	Subscriber
	nextFunc   func(interface{})
	nextChan   chan interface{}
	errFunc    func(error)
	errChan    chan error
	compFunc   func()
	compChan   chan bool
	disposable Disposable
	dispChan   chan bool
}

func (s *subscriber) SendNext(value interface{}) {
	if s.nextChan == nil {
		return
	}
	s.nextChan <- value
}

func (s *subscriber) SendError(err error) {
	if s.errChan != nil {
		s.errChan <- err
	}
	s.dispChan <- false
}

func (s *subscriber) SendCompleted() {
	if s.compChan != nil {
		s.compChan <- true
	}
	s.dispChan <- true
}

func (s *subscriber) Disposable() Disposable {
	return s.disposable
}

func newSubstriber(next func(interface{}), err func(error), complete func()) *subscriber {
	var subscriber = &subscriber{
		nextFunc: next,
		errFunc:  err,
		compFunc: complete,
		dispChan: make(chan bool),
	}
	if next != nil {
		subscriber.nextChan = make(chan interface{})
	}
	if err != nil {
		subscriber.errChan = make(chan error)
	}
	if complete != nil {
		subscriber.compChan = make(chan bool)
	}
	subscriber.disposable = NewDisposable(func() {
		subscriber.nextFunc = nil
		subscriber.nextChan = nil
		subscriber.errFunc = nil
		subscriber.errChan = nil
		subscriber.compFunc = nil
		subscriber.compChan = nil
		subscriber.disposable = nil
		subscriber.dispChan = nil
	})
	subscriber.disposable.AddDispositionChan(subscriber.dispChan)
	go func() {
		for {
			select {
			case v := <-subscriber.nextChan:
				subscriber.nextFunc(v)
			case e := <-subscriber.errChan:
				subscriber.errFunc(e)
			case <-subscriber.compChan:
				subscriber.compFunc()
			case <-subscriber.dispChan:
				subscriber.disposable.Dispose()
				return
			}
		}
	}()
	return subscriber
}
