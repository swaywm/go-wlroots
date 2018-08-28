package wlroots

// #include <wlr/types/wlr_compositor.h>
import "C"
import "unsafe"

type Compositor struct {
	p *C.struct_wlr_compositor
}

func NewCompositor(display Display, renderer Renderer) Compositor {
	p := C.wlr_compositor_create(display.p, renderer.p)
	man.track(unsafe.Pointer(p), &p.events.destroy)
	return Compositor{p: p}
}

func (c Compositor) Destroy() {
	C.wlr_compositor_destroy(c.p)
}

func (c Compositor) OnDestroy(cb func(Compositor)) {
	man.add(unsafe.Pointer(c.p), &c.p.events.destroy, func(unsafe.Pointer) {
		cb(c)
	})
}
