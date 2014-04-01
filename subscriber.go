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
	errFunc    func(error)
	compFunc   func()
	disposable Disposable

	operationChan chan func()
}

func (s *subscriber) SendNext(value interface{}) {
	s.operationChan <- func() {
		if s.nextFunc != nil {
			s.nextFunc(value)
		}
	}
}

func (s *subscriber) SendError(err error) {
	s.operationChan <- func() {
		if s.errFunc != nil {
			s.errFunc(err)
		}
		s.Disposable().Dispose()
	}
}

func (s *subscriber) SendCompleted() {
	s.operationChan <- func() {
		if s.compFunc != nil {
			s.compFunc()
		}
		s.Disposable().Dispose()
	}
}

func (s *subscriber) Disposable() Disposable {
	return s.disposable
}

func NewSubscriber(next func(interface{}), err func(error), complete func()) Subscriber {
	var subscriber = &subscriber{
		nextFunc:      next,
		errFunc:       err,
		compFunc:      complete,
		operationChan: make(chan func()),
	}
	subscriber.disposable = NewDisposable(func() {
		subscriber.nextFunc = nil
		subscriber.errFunc = nil
		subscriber.compFunc = nil
	})
	go func() {
		for {
			select {
			case op := <-subscriber.operationChan:
				op()
			case <-subscriber.disposable.DispositionChan():
				return
			}
		}
	}()
	return subscriber
}
