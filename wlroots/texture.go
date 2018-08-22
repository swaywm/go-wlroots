package wlroots

// #include <wlr/render/wlr_texture.h>
import "C"

type Texture struct {
	p *C.struct_wlr_texture
}

func (t Texture) Destroy() {
	C.wlr_texture_destroy(t.p)
}

func (t Texture) Nil() bool {
	return t.p == nil
}
