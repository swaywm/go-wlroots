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

func (o Output) OnDestroy(cb func(Output)) {
	man.add(unsafe.Pointer(o.p), &o.p.events.destroy, func(unsafe.Pointer) {
		cb(o)
	})
}

func (o Output) Name() string {
	return C.GoString(&o.p.name[0])
}

func (o Output) Scale() float32 {
	return float32(o.p.scale)
}

func (o Output) TransformMatrix() Matrix {
	var matrix Matrix
	matrix.fromC(&o.p.transform_matrix)
	return matrix
}

func (o Output) OnFrame(cb func(Output)) {
	man.add(unsafe.Pointer(o.p), &o.p.events.frame, func(data unsafe.Pointer) {
		cb(o)
	})
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
		return 0, errors.New("can't make output context current")
	}

	return int(bufferAge), nil
}

func (o Output) CreateGlobal() {
	C.wlr_output_create_global(o.p)
}

func (o Output) DestroyGlobal() {
	C.wlr_output_destroy_global(o.p)
}

func (o Output) SwapBuffers() {
	C.wlr_output_swap_buffers(o.p, nil, nil)
}

func (o Output) Modes() []OutputMode {
	// TODO: figure out what to do with this ridiculous for loop
	// perhaps this can be refactored into a less ugly hack that uses reflection
	var modes []OutputMode
	var mode *C.struct_wlr_output_mode
	for mode := (*C.struct_wlr_output_mode)(unsafe.Pointer(uintptr(unsafe.Pointer(o.p.modes.next)) - unsafe.Offsetof(mode.link))); &mode.link != &o.p.modes; mode = (*C.struct_wlr_output_mode)(unsafe.Pointer(uintptr(unsafe.Pointer(mode.link.next)) - unsafe.Offsetof(mode.link))) {
		modes = append(modes, OutputMode{p: mode})
	}

	return modes
}

func (o Output) SetMode(mode OutputMode) {
	C.wlr_output_set_mode(o.p, mode.p)
}

type OutputLayout struct {
	p *C.struct_wlr_output_layout
}

func NewOutputLayout() OutputLayout {
	p := C.wlr_output_layout_create()
	man.track(unsafe.Pointer(p), &p.events.destroy)
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
