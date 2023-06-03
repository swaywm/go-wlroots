package wlroots

import (
	"sync"
	"time"
	"unsafe"
)

// This whole mess has to exist for a number of reasons:
//
// 1. We need to allocate all instances of wl_listener on the heap as storing Go
// pointers in C after a cgo call returns is not allowed.
//
// 2. The wlroots library implicitly destroys objects when wl_display is
// destroyed. So, we need to keep track of all objects (and their listeners)
// manually and listen for the destroy signal to be able to free everything.
//
// 3 (TODO). As we're keeping track of all objects anyway, we might as well
// store a Go pointer to the wrapper struct along with them in order to be able
// to pass the same Go pointer through callbacks every time. This will also
// allow calling runtime.SetFinalizer on some of them to clean them up early
// when the GC notices it has gone out of scope.
//
// Send help.

// #cgo pkg-config: wlroots wayland-server
// #cgo CFLAGS: -D_GNU_SOURCE -DWLR_USE_UNSTABLE
// #include <stdlib.h>
// #include <wayland-server-core.h>
//
// void _wl_listener_cb(struct wl_listener *listener, void *data);
//
// static inline void _wl_listener_set_cb(struct wl_listener *listener) {
//		listener->notify = &_wl_listener_cb;
// }
//
import "C"

type EventLoop struct {
	p *C.struct_wl_event_loop
}

func (evl EventLoop) OnDestroy(cb func(EventLoop)) {
	l := man.add(unsafe.Pointer(evl.p), nil, func(data unsafe.Pointer) {
		cb(evl)
	})
	C.wl_event_loop_add_destroy_listener(evl.p, l.p)
}

func (evl EventLoop) Fd() uintptr {
	return uintptr(C.wl_event_loop_get_fd(evl.p))
}

func (evl EventLoop) Dispatch(timeout time.Duration) {
	var d int
	if timeout >= 0 {
		d = int(timeout / time.Millisecond)
	} else {
		d = -1
	}
	C.wl_event_loop_dispatch(evl.p, C.int(d))
}

type (
	listenerCallback func(data unsafe.Pointer)
)

type manager struct {
	mutex     sync.RWMutex
	objects   map[unsafe.Pointer][]*listener
	listeners map[*C.struct_wl_listener]*listener
}

type listener struct {
	p   *C.struct_wl_listener
	s   *C.struct_wl_signal
	cbs []listenerCallback
}

var (
	man = &manager{
		objects:   map[unsafe.Pointer][]*listener{},
		listeners: map[*C.struct_wl_listener]*listener{},
	}
)

//export _wl_listener_cb
func _wl_listener_cb(listener *C.struct_wl_listener, data unsafe.Pointer) {
	man.mutex.RLock()
	l := man.listeners[listener]
	man.mutex.RUnlock()
	for _, cb := range l.cbs {
		cb(data)
	}
}

func (m *manager) add(p unsafe.Pointer, signal *C.struct_wl_signal, cb listenerCallback) *listener {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// if a listener for this object and signal already exists, add the callback
	// to the existing listener
	if signal != nil {
		for _, l := range m.objects[p] {
			if l.s != nil && l.s == signal {
				l.cbs = append(l.cbs, cb)
				return l
			}
		}
	}

	lp := (*C.struct_wl_listener)(C.calloc(C.sizeof_struct_wl_listener, 1))
	C._wl_listener_set_cb(lp)
	if signal != nil {
		C.wl_signal_add((*C.struct_wl_signal)(signal), lp)
	}

	l := &listener{
		p:   lp,
		s:   signal,
		cbs: []listenerCallback{cb},
	}
	m.listeners[lp] = l
	m.objects[p] = append(m.objects[p], l)

	return l
}

func (m *manager) has(p unsafe.Pointer) bool {
	m.mutex.RLock()
	_, found := m.objects[p]
	m.mutex.RUnlock()
	return found
}

func (m *manager) delete(p unsafe.Pointer) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	for _, l := range m.objects[p] {
		delete(m.listeners, l.p)

		// remove the listener from the signal
		C.wl_list_remove(&l.p.link)

		// free the listener
		C.free(unsafe.Pointer(l.p))
	}

	delete(m.objects, p)
}

func (m *manager) track(p unsafe.Pointer, destroySignal *C.struct_wl_signal) {
	m.add(p, destroySignal, func(data unsafe.Pointer) { m.delete(p) })
}
