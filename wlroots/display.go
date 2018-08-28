package wlroots

// #include <wayland-server.h>
import "C"
import (
	"errors"
	"unsafe"
)

type Display struct {
	p *C.struct_wl_display
}

func NewDisplay() Display {
	p := C.wl_display_create()
	d := Display{p: p}
	d.OnDestroy(func(Display) {
		man.delete(unsafe.Pointer(p))
	})
	return d
}

func (d Display) Destroy() {
	C.wl_display_destroy(d.p)
}

func (d Display) OnDestroy(cb func(Display)) {
	l := man.add(unsafe.Pointer(d.p), nil, func(data unsafe.Pointer) {
		cb(d)
	})
	C.wl_display_add_destroy_listener(d.p, l.p)
}

func (d Display) Run() {
	C.wl_display_run(d.p)
}

func (d Display) Terminate() {
	C.wl_display_terminate(d.p)
}

func (d Display) AddSocketAuto() (string, error) {
	socket := C.wl_display_add_socket_auto(d.p)
	if socket == nil {
		return "", errors.New("can't auto add wayland socket")
	}

	return C.GoString(socket), nil
}
