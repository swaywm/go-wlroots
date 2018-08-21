package wlroots

// #include <wlr/types/wlr_output.h>
// #include <wlr/types/wlr_output_layout.h>
import "C"

import (
	"errors"
	"unsafe"
)

type Output struct {
	p *C.struct_wlr_output
}

type OutputMode struct {
	p *C.struct_wlr_output_mode
}

func wrapOutput(p unsafe.Pointer) Output {
	return Output{p: (*C.struct_wlr_output)(p)}
}

func (o Output) Name() string {
	return C.GoString(&o.p.name[0])
}

func (o Output) Scale() float32 {
	return float32(o.p.scale)
}

func (o Output) TransformMatrix() Matrix {
	return Matrix{m: o.p.transform_matrix}
}

func (o Output) OnFrame(cb func(Output)) {
	listener := NewListener(func(data unsafe.Pointer) {
		cb(o)
	})

	C.wl_signal_add(&o.p.events.frame, listener.p)
}

func (o Output) OnDestroy(cb func(Output)) {
	listener := NewListener(func(data unsafe.Pointer) {
		cb(o)
	})

	C.wl_signal_add(&o.p.events.destroy, listener.p)
}

func (o Output) TransformedResolution() (int, int) {
	var width, height C.int
	C.wlr_output_transformed_resolution(o.p, &width, &height)
	return int(width), int(height)
}

func (o Output) EffectiveResolution() (int, int) {
	var width, height C.int
	C.wlr_output_effective_resolution(o.p, &width, &height)
	return int(width), int(height)
}

func (o Output) MakeCurrent() (int, error) {
	var bufferAge C.int
	if !C.wlr_output_make_current(o.p, &bufferAge) {
		return 0, errors.New("error making output context current")
	}

	return int(bufferAge), nil
}

func (o Output) CreateGlobal() {
	C.wlr_output_create_global(o.p)
}

func (o Output) SwapBuffers() {
	C.wlr_output_swap_buffers(o.p, nil, nil)
}

func (o Output) Modes() []OutputMode {
	return nil
}

type OutputLayout struct {
	p *C.struct_wlr_output_layout
}

func NewOutputLayout() OutputLayout {
	p := C.wlr_output_layout_create()
	return OutputLayout{p: p}
}

func (l OutputLayout) Destroy() {
	C.wlr_output_layout_destroy(l.p)
}

func (l OutputLayout) AddOutputAuto(output Output) {
	C.wlr_output_layout_add_auto(l.p, output.p)
}

func (l OutputLayout) Coords(output Output) (x float64, y float64) {
	var ox, oy C.double
	C.wlr_output_layout_output_coords(l.p, output.p, &ox, &oy)
	return float64(ox), float64(oy)
}

func OutputTransformInvert(transform uint32) uint32 {
	return uint32(C.wlr_output_transform_invert(C.enum_wl_output_transform(transform)))
}

func (m OutputMode) Flags() uint32 {
	return uint32(m.p.flags)
}

func (m OutputMode) Width() int32 {
	return int32(m.p.width)
}

func (m OutputMode) Height() int32 {
	return int32(m.p.height)
}

func (m OutputMode) RefreshRate() int32 {
	return int32(m.p.refresh)
}
