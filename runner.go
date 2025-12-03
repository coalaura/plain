package plain

import "sync"

type runner struct {
	mx sync.Mutex
	fn func()
}

func (r *runner) Store(fn func()) {
	r.mx.Lock()
	defer r.mx.Unlock()

	r.fn = fn
}

func (r *runner) RunAndClear() {
	r.mx.Lock()
	defer r.mx.Unlock()

	if r.fn == nil {
		return
	}

	r.fn()

	r.fn = nil
}
