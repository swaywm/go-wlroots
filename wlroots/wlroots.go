package wlroots

// #include <wayland-server.h>
// #include <wlr/backend.h>
// #cgo pkg-config: wlroots wayland-server
// #cgo CFLAGS: -DWLR_USE_UNSTABLE
import "C"

import (
	"errors"
	"unsafe"
)

type Color struct {
	R, G, B, A float32
}

type Backend struct {
	p *C.struct_wlr_backend
}

func NewBackend(display Display) Backend {
	p := C.wlr_backend_autocreate(display.p, nil)
	return Backend{p: p}
}

func (b Backend) Destroy() {
	C.wlr_backend_destroy(b.p)
}

func (b Backend) Start() error {
	if !C.wlr_backend_start(b.p) {
		return errors.New("can't start backend")
	}

	return nil
}

func (b Backend) OnNewOutput(cb func(Output)) {
	listener := NewListener(func(data unsafe.Pointer) {
		cb(wrapOutput(data))
	})

	C.wl_signal_add(&b.p.events.new_output, listener.p)
}

func (b Backend) OnNewInput(cb func(InputDevice)) {
	listener := NewListener(func(data unsafe.Pointer) {
		cb(wrapInputDevice(data))
	})

	C.wl_signal_add(&b.p.events.new_input, listener.p)
}

func (b Backend) Renderer() Renderer {
	p := C.wlr_backend_get_renderer(b.p)
	return Renderer{p: p}
}
