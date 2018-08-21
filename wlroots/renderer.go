package wlroots

// #include <wlr/render/wlr_renderer.h>
import "C"

type Renderer struct {
	p *C.struct_wlr_renderer
}

func (r Renderer) InitDisplay(display Display) {
	C.wlr_renderer_init_wl_display(r.p, display.p)
}

func (r Renderer) Begin(output Output, width int, height int) {
	C.wlr_renderer_begin(r.p, C.int(width), C.int(height))
}

func (r Renderer) Clear(color Color) {
	c := []C.float{C.float(color.R), C.float(color.G), C.float(color.B), C.float(color.A)}
	C.wlr_renderer_clear(r.p, &c[0])
}

func (r Renderer) End() {
	C.wlr_renderer_end(r.p)
}

func (r Renderer) RenderTextureWithMatrix(texture Texture, matrix *Matrix, alpha float32) {
	C.wlr_render_texture_with_matrix(r.p, texture.p, &matrix.m[0], C.float(alpha))
}
