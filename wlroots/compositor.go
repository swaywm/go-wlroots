package wlroots

// #include <wlr/types/wlr_compositor.h>
import "C"

type Compositor struct {
	p *C.struct_wlr_compositor
}

func NewCompositor(display Display, renderer Renderer) Compositor {
	p := C.wlr_compositor_create(display.p, renderer.p)
	return Compositor{p: p}
}

func (c Compositor) Destroy() {
	C.wlr_compositor_destroy(c.p)
}
