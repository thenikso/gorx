package rx

type Subscriber interface {
	SendNext(interface{})
	SendError(error)
	SendCompleted()
}

type liveSubscriber struct {
	Next       func(interface{})
	Err        func(error)
	Complete   func()
	Disposable Disposable
}

func (l *liveSubscriber) SendNext(value interface{}) {
	l.Next(value)
}

func (l *liveSubscriber) SendError(err error) {
	l.Err(err)
}

func (l *liveSubscriber) SendCompleted() {
	var completeFunc = l.Complete
	l.Disposable.Dispose()
	completeFunc()
}

func newLiveSubstriber(next func(interface{}), err func(error), complete func()) *liveSubscriber {
	var subscriber = &liveSubscriber{
		Next:     next,
		Err:      err,
		Complete: complete,
	}
	subscriber.Disposable = NewCompoundDisposable(func() {
		subscriber.Next = nil
		subscriber.Err = nil
		subscriber.Complete = nil
	})
	return subscriber
}
