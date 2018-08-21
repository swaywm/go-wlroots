package wlroots

// #include <wlr/types/wlr_linux_dmabuf_v1.h>
import "C"

type DMABuf struct {
	p *C.struct_wlr_linux_dmabuf_v1
}

func NewDMABuf(display Display, renderer Renderer) DMABuf {
	p := C.wlr_linux_dmabuf_v1_create(display.p, renderer.p)
	return DMABuf{p: p}
}
