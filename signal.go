package rx

type nextFunc func(interface{})
type errorFunc func(error)
type completeFunc func()

type Signal struct {
}

func (s *Signal) Subscribe(next nextFunc, err errorFunc, complete completeFunc) Disposable {
	var subscriber = newLiveSubstriber(next, err, complete)
	s.attachSubscriber(subscriber)
	return subscriber.Disposable
}

func (s *Signal) attachSubscriber(subscriber Subscriber) {
	// Default to empty signal
	subscriber.SendCompleted()
}
