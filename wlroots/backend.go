package wlroots

/*
 * This an unstable interface of wlroots. No guarantees are made regarding the
 * future consistency of this API.
 */

import (
	"errors"
	"unsafe"
)

// #cgo pkg-config: wlroots-0.18 wayland-server
// #cgo CFLAGS: -D_GNU_SOURCE -DWLR_USE_UNSTABLE
// #include <wlr/backend.h>
// #include <wlr/render/allocator.h>
// #include <wlr/render/wlr_renderer.h>
import "C"

type Backend struct {
	p *C.struct_wlr_backend
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
			// delete the wlr_input_device
			man.delete(unsafe.Pointer(dev.p))
		})
		cb(dev)
	})
}

func (b Backend) AllocatorAutocreate(r Renderer) (Allocator, error) {
	p := C.wlr_allocator_autocreate(b.p, r.p)
	if p == nil {
		return Allocator{}, errors.New("failed to wlr_allocator")
	}
	man.track(unsafe.Pointer(p), &p.events.destroy)
	return Allocator{p: p}, nil
}

func (b Backend) NewAllocator(r Renderer) (Allocator, error) {
	return b.AllocatorAutocreate(r)
}

func (b Backend) RendererAutoCreate() (Renderer, error) {
	p := C.wlr_renderer_autocreate(b.p)
	if p == nil {
		return Renderer{}, errors.New("failed to create wlr_renderer")
	}
	man.track(unsafe.Pointer(p), &p.events.destroy)
	return Renderer{p: p}, nil
}

func (b Backend) NewRenderer() (Renderer, error) {
	return b.RendererAutoCreate()
}

type Allocator struct {
	p *C.struct_wlr_allocator
}

func (s Allocator) Nil() bool {
	return s.p == nil
}
