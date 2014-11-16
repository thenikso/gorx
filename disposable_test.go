package rx

import (
	"testing"
)

func TestSimpleDisposableShouldDispose(t *testing.T) {
	disposable := NewSimpleDisposable()
	if disposable.IsDisposed() != false {
		t.Error("Expect SimpleDisposable to have IsDisposed to false")
	}
	err := disposable.Dispose()
	if err != nil || disposable.IsDisposed() != true {
		t.Error("Expect SimpleDisposable dispose")
	}
}

func TestActionDisposableShouldRunActionUponDisposal(t *testing.T) {
	didDispose := false
	disposable := NewActionDisposable(func() error {
		didDispose = true
		return nil
	})

	if didDispose != false {
		t.Error("Expect `didDispose` to be false")
	}
	if disposable.IsDisposed() != false {
		t.Error("Expect `disposable.IsDisposed()` to be false")
	}

	disposable.Dispose()

	if didDispose != true {
		t.Error("Expect `didDispose` to be true")
	}
	if disposable.IsDisposed() != true {
		t.Error("Expect `disposable.IsDisposed()` to be true")
	}
}

func TestCompositeDisposableShouldDisposeAddedDisposables(t *testing.T) {
	disposable := NewCompositeDisposable(nil)

	simpleDisposable := NewSimpleDisposable()
	disposable.AddDisposable(simpleDisposable)

	didDispose := false
	disposable.AddDisposableFunc(func() error {
		didDispose = true
		return nil
	})

	if simpleDisposable.IsDisposed() != false {
		t.Error("Expect `simpleDisposable.IsDisposed()` to be false")
	}
	if didDispose != false {
		t.Error("Expect `didDispose` to be false")
	}
	if disposable.IsDisposed() != false {
		t.Error("Expect `disposable.IsDisposed()` to be false")
	}

	disposable.Dispose()

	if simpleDisposable.IsDisposed() != true {
		t.Error("Expect `simpleDisposable.IsDisposed()` to be true")
	}
	if didDispose != true {
		t.Error("Expect `didDispose` to be true")
	}
	if disposable.IsDisposed() != true {
		t.Error("Expect `disposable.IsDisposed()` to be true")
	}
}

func TestCompositeDisposableShouldNotPruneActiveDisposables(t *testing.T) {
	disposable := NewCompositeDisposable(nil)

	simpleDisposable := NewSimpleDisposable()
	disposable.AddDisposable(simpleDisposable)

	didDispose := false
	disposable.AddDisposableFunc(func() error {
		didDispose = true
		return nil
	})

	simpleDisposable.Dispose()

	disposable.PruneDisposed()
	if didDispose != false {
		t.Error("Expect `didDispose` to be false")
	}

	disposable.Dispose()
	if didDispose != true {
		t.Error("Expect `didDispose` to be true")
	}
}

func TestSerialDisposableShouldDisposeOfInnerDisposable(t *testing.T) {
	disposable := NewSerialDisposable(nil)

	simpleDisposable := NewSimpleDisposable()
	disposable.SetInnerDisposable(simpleDisposable)

	if disposable.InnerDisposable == nil {
		t.Error("Expect `disposable.InnerDisposable` not to be nil")
	}
	if simpleDisposable.IsDisposed() != false {
		t.Error("Expect `simpleDisposable.IsDisposed()` to be false")
	}
	if disposable.IsDisposed() != false {
		t.Error("Expect `disposable.IsDisposed()` to be false")
	}

	disposable.Dispose()
	if disposable.InnerDisposable() != nil {
		t.Errorf("Expect `disposable.InnerDisposable` to be nil, got %v", disposable.InnerDisposable())
	}
	if simpleDisposable.IsDisposed() != true {
		t.Error("Expect `simpleDisposable.IsDisposed()` to be true")
	}
	if disposable.IsDisposed() != true {
		t.Error("Expect `disposable.IsDisposed()` to be true")
	}
}

func TestSerialDisposableShouldDisposeOfPreviousDisposableWhenSwappingInnerDisposable(t *testing.T) {
	disposable := NewSerialDisposable(nil)

	oldDisposable := NewSimpleDisposable()
	newDisposable := NewSimpleDisposable()
	disposable.SetInnerDisposable(oldDisposable)

	if oldDisposable.IsDisposed() != false {
		t.Error("Expect `oldDisposable.IsDisposed()` to be false")
	}
	if newDisposable.IsDisposed() != false {
		t.Error("Expect `newDisposable.IsDisposed()` to be false")
	}

	disposable.SetInnerDisposable(newDisposable)
	if oldDisposable.IsDisposed() != true {
		t.Error("Expect `oldDisposable.IsDisposed()` to be true")
	}
	if newDisposable.IsDisposed() != false {
		t.Error("Expect `newDisposable.IsDisposed()` to be false")
	}
	if disposable.IsDisposed() != false {
		t.Error("Expect `disposable.IsDisposed()` to be false")
	}

	disposable.SetInnerDisposable(nil)
	if newDisposable.IsDisposed() != true {
		t.Error("Expect `newDisposable.IsDisposed()` to be true")
	}
	if disposable.IsDisposed() != false {
		t.Error("Expect `disposable.IsDisposed()` to be false")
	}
}
