package wlroots

/*
 * This an unstable interface of wlroots. No guarantees are made regarding the
 * future consistency of this API.
 */

// #cgo pkg-config: wlroots-0.18 wayland-server
// #cgo CFLAGS: -D_GNU_SOURCE -DWLR_USE_UNSTABLE
// #include <wlr/render/wlr_renderer.h>
import "C"

type (
	BlendMode  uint32
	FilterMode uint32
)

const (
	BlendModePremultiplied BlendMode = C.WLR_RENDER_BLEND_MODE_PREMULTIPLIED
	BlendModeNone          BlendMode = C.WLR_RENDER_BLEND_MODE_NONE

	FilterBilinear FilterMode = C.WLR_SCALE_FILTER_BILINEAR
	FilterNearest  FilterMode = C.WLR_SCALE_FILTER_NEAREST
)

type RenderPass struct {
	p *C.struct_wlr_render_pass
}

func (r RenderPass) Submit() {
	C.wlr_render_pass_submit(r.p)
}

func (r RenderPass) AddTexture(texture Texture, srcBox FBox, dstBox GeoBox, alpha float32, transform uint32, filterMode FilterMode, blendMode BlendMode) {
	var alphaC C.float
	alphaC = C.float(alpha)
	var options C.struct_wlr_render_texture_options
	options.texture = texture.p
	options.src_box = srcBox.toC()
	options.dst_box = dstBox.toC()
	options.alpha = &alphaC
	options.transform = C.enum_wl_output_transform(transform)
	options.filter_mode = uint32(filterMode)
	options.blend_mode = uint32(blendMode)
	C.wlr_render_pass_add_texture(r.p, &options)
}

func (r RenderPass) AddRect(box *GeoBox, color *Color, blendMode BlendMode) {
	var options C.struct_wlr_render_rect_options
	options.box = box.toC()
	options.color = color.toC()
	options.blend_mode = uint32(blendMode)
	C.wlr_render_pass_add_rect(r.p, &options)
}
