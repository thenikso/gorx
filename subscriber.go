package rx

// An Subscriber is a receiver of events from an Signal.
type Subscriber interface {
	OnNext(T)
	OnError(error)
	OnCompleted()
	Disposable() CompositeDisposable
}

type subscriber struct {
	nextFunc   func(T)
	errFunc    func(error)
	compFunc   func()
	disposable CompositeDisposable
}

func (o *subscriber) OnNext(value T) {
	if o.nextFunc != nil {
		o.nextFunc(value)
	}
}

func (o *subscriber) OnError(err error) {
	if o.errFunc != nil {
		o.errFunc(err)
	}
	o.disposable.Dispose()
}

func (o *subscriber) OnCompleted() {
	if o.compFunc != nil {
		o.compFunc()
	}
	o.disposable.Dispose()
}

func (o *subscriber) Disposable() CompositeDisposable {
	return o.disposable
}

func NewSubscriber(next func(T), err func(error), completed func()) Subscriber {
	var subscriber = &subscriber{
		nextFunc: next,
		errFunc:  err,
		compFunc: completed,
	}
	subscriber.disposable = NewCompositeDisposable(func() error {
		subscriber.nextFunc = nil
		subscriber.errFunc = nil
		subscriber.compFunc = nil
		return nil
	})
	return subscriber
}
