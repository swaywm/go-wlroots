package wlroots

// #include <wayland-server.h>
//
// void _wl_listener_cb(struct wl_listener *listener, void *data);
//
// static inline void _wl_listener_set_cb(struct wl_listener *listener) {
//		listener->notify = &_wl_listener_cb;
// }
import "C"
import "unsafe"

type (
	ListenerCallback func(data unsafe.Pointer)
)

type Listener struct {
	p  *C.struct_wl_listener
	cb ListenerCallback
}

var (
	// TODO: guard this with a mutex
	listeners = map[*C.struct_wl_listener]Listener{}
)

//export _wl_listener_cb
func _wl_listener_cb(listener *C.struct_wl_listener, data unsafe.Pointer) {
	l := listeners[listener]
	if l.cb != nil {
		l.cb(data)
	}
}

func NewListener(cb ListenerCallback) Listener {
	p := new(C.struct_wl_listener)
	C._wl_listener_set_cb(p)

	listener := Listener{p: p, cb: cb}
	listeners[p] = listener
	return listener
}
