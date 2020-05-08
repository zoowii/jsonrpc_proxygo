package registry

import "sync"

type Watcher struct {
	lock sync.RWMutex
	ch   chan *Event
}

func NewWatcher() *Watcher {
	return &Watcher{
		ch: make(chan *Event),
	}
}

func (w *Watcher) Close() {
	if w.ch != nil {
		w.lock.Lock()
		defer w.lock.Unlock()
		if w.ch != nil {
			close(w.ch)
			w.ch = nil
		}
	}
}

func (w *Watcher) Send(event *Event) {
	w.lock.RLock()
	defer w.lock.RUnlock()
	ch := w.ch
	if ch == nil {
		return
	}
	ch <- event
}

func (w *Watcher) C() chan *Event {
	return w.ch
}
