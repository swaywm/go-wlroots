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
	man.track(unsafe.Pointer(p), &p.events.destroy)
	return Backend{p: p}
}

func (b Backend) Destroy() {
	C.wlr_backend_destroy(b.p)
}

func (b Backend) OnDestroy(cb func(Backend)) {
	man.add(unsafe.Pointer(b.p), &b.p.events.destroy, func(unsafe.Pointer) {
		cb(b)
	})
}

func (b Backend) Start() error {
	if !C.wlr_backend_start(b.p) {
		return errors.New("can't start backend")
	}

	return nil
}

func (b Backend) OnNewOutput(cb func(Output)) {
	man.add(unsafe.Pointer(b.p), &b.p.events.new_output, func(data unsafe.Pointer) {
		output := wrapOutput(data)
		man.track(unsafe.Pointer(output.p), &output.p.events.destroy)
		cb(output)
	})
}

func (b Backend) OnNewInput(cb func(InputDevice)) {
	man.add(unsafe.Pointer(b.p), &b.p.events.new_input, func(data unsafe.Pointer) {
		dev := wrapInputDevice(data)
		man.add(unsafe.Pointer(dev.p), &dev.p.events.destroy, func(data unsafe.Pointer) {
			// delete the underlying device type first
			man.delete(*(*unsafe.Pointer)(unsafe.Pointer(&dev.p.anon0[0])))
			// then delete the wlr_input_device itself
			man.delete(unsafe.Pointer(dev.p))
		})
		cb(dev)
	})
}

func (b Backend) Renderer() Renderer {
	p := C.wlr_backend_get_renderer(b.p)
	// TODO: find out why events.destroy is not found
	// (error: inconsistent definitions for C.struct_wlr_renderer)
	//man.track(unsafe.Pointer(p), &p.events.destroy)
	return Renderer{p: p}
}
