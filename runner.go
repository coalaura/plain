package plain

import "sync"

type runner struct {
	mx sync.Mutex
	fn func()
}

// Store sets the function to be run by RunAndClear, replacing any previously stored function
func (r *runner) Store(fn func()) {
	r.mx.Lock()
	defer r.mx.Unlock()

	r.fn = fn
}

// RunAndClear runs the stored function if present and then clears it
func (r *runner) RunAndClear() {
	r.mx.Lock()
	defer r.mx.Unlock()

	if r.fn == nil {
		return
	}

	r.fn()

	r.fn = nil
}
