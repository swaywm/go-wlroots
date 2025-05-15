package wlroots

/*
 * This an unstable interface of wlroots. No guarantees are made regarding the
 * future consistency of this API.
 */

import "unsafe"

// #cgo pkg-config: wlroots wayland-server
// #cgo CFLAGS: -D_GNU_SOURCE -DWLR_USE_UNSTABLE
// #include <wlr/render/wlr_renderer.h>
import "C"

type Renderer struct {
	p *C.struct_wlr_renderer
}

func (r Renderer) Destroy() {
	C.wlr_renderer_destroy(r.p)
}

func (r Renderer) OnDestroy(cb func(Renderer)) {
	man.add(unsafe.Pointer(r.p), &r.p.events.destroy, func(unsafe.Pointer) {
		cb(r)
	})
}

func (r Renderer) InitDisplay(display Display) {
	C.wlr_renderer_init_wl_display(r.p, display.p)
}
