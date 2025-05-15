package wlroots

import (
	"errors"
	"unsafe"
)

// #cgo pkg-config: wlroots-0.18 wayland-server
// #cgo CFLAGS: -D_GNU_SOURCE -DWLR_USE_UNSTABLE
// #include <stdlib.h>
// #include <time.h>
// #include <wayland-server-core.h>
// #include <wlr/backend.h>
// #include <wlr/types/wlr_compositor.h>
// #include <wlr/types/wlr_subcompositor.h>
// #include <wlr/types/wlr_data_device.h>
// #include <wlr/types/wlr_output_layout.h>
// #include <wlr/types/wlr_xcursor_manager.h>
// #include <wlr/types/wlr_xdg_shell.h>
// #include <wlr/render/wlr_texture.h>
// #include <wlr/types/wlr_linux_dmabuf_v1.h>
// #include <wlr/types/wlr_matrix.h>
// #include <wlr/util/box.h>
// #include <wlr/util/edges.h>
import "C"

type XCursorManager struct {
	p *C.struct_wlr_xcursor_manager
}

func NewXCursorManager(name string, size int) XCursorManager {
	p := C.wlr_xcursor_manager_create(C.CString(name), C.uint(size))
	return XCursorManager{p: p}
}

func (m XCursorManager) Destroy() {
	C.wlr_xcursor_manager_destroy(m.p)
}

func (m XCursorManager) Load(scale float32) {
	C.wlr_xcursor_manager_load(m.p, C.float(scale))
}

type Edges uint32

const (
	EdgeNone   Edges = C.WLR_EDGE_NONE
	EdgeTop    Edges = C.WLR_EDGE_TOP
	EdgeBottom Edges = C.WLR_EDGE_BOTTOM
	EdgeLeft   Edges = C.WLR_EDGE_LEFT
	EdgeRight  Edges = C.WLR_EDGE_RIGHT
)

type Texture struct {
	p *C.struct_wlr_texture
}

func (t Texture) Destroy() {
	C.wlr_texture_destroy(t.p)
}

func (t Texture) Nil() bool {
	return t.p == nil
}

type OutputLayout struct {
	p *C.struct_wlr_output_layout
}
type OutputLayoutOutput struct {
	p *C.struct_wlr_output_layout_output
}

func NewOutputLayout(d Display) OutputLayout {
	return OutputLayoutCreate(d)
}

func OutputLayoutCreate(d Display) OutputLayout {
	p := C.wlr_output_layout_create(d.p)
	man.track(unsafe.Pointer(p), &p.events.destroy)
	return OutputLayout{p: p}
}

func (l OutputLayout) Destroy() {
	C.wlr_output_layout_destroy(l.p)
}

func (l OutputLayout) AddOutputAuto(output Output) OutputLayoutOutput {
	p := C.wlr_output_layout_add_auto(l.p, output.p)
	return OutputLayoutOutput{p: p}
}

func (l OutputLayout) Coords(output Output) (x float64, y float64) {
	var ox, oy C.double
	C.wlr_output_layout_output_coords(l.p, output.p, &ox, &oy)
	return float64(ox), float64(oy)
}

type Matrix [9]float32

func (m *Matrix) ProjectBox(box *GeoBox, transform uint32, rotation float32, projection *Matrix) {
	cm := m.toC()
	b := box.toC()
	pm := projection.toC()
	C.wlr_matrix_project_box(&cm[0], &b, C.enum_wl_output_transform(transform), C.float(rotation), &pm[0])
	m.fromC(&cm)
}

func (m *Matrix) toC() [9]C.float {
	var cm [9]C.float
	for i := range m {
		cm[i] = C.float(m[i])
	}
	return cm
}

func (m *Matrix) fromC(cm *[9]C.float) {
	for i := range cm {
		m[i] = float32(cm[i])
	}
}

type DMABuf struct {
	p *C.struct_wlr_linux_dmabuf_v1
}

func NewDMABuf(display Display, renderer Renderer) DMABuf {
	p := C.wlr_linux_dmabuf_v1_create_with_renderer(display.p, 4, renderer.p)
	man.track(unsafe.Pointer(p), &p.events.destroy)
	return DMABuf{p: p}
}

func (b DMABuf) OnDestroy(cb func(DMABuf)) {
	man.add(unsafe.Pointer(b.p), &b.p.events.destroy, func(unsafe.Pointer) {
		cb(b)
	})
}

type Display struct {
	p *C.struct_wl_display
}

func DisplayCreate() {
	NewDisplay()
}

func NewDisplay() Display {
	p := C.wl_display_create()
	d := Display{p: p}
	d.OnDestroy(func(Display) {
		man.delete(unsafe.Pointer(p))
	})
	return d
}

func (d Display) NewBackend() (Backend, error) {
	return d.BackendAutocreate()
}

func (d Display) BackendAutocreate() (Backend, error) {
	p := C.wlr_backend_autocreate(C.wl_display_get_event_loop(d.p), nil)
	if p == nil {
		return Backend{}, errors.New("failed to create wlr_backend")
	}
	man.track(unsafe.Pointer(p), &p.events.destroy)
	return Backend{p: p}, nil
}

func (d Display) NewSubCompositor() SubCompositor {
	return d.SubCompositorCreate()
}

func (d Display) SubCompositorCreate() SubCompositor {
	p := C.wlr_subcompositor_create(d.p)
	man.track(unsafe.Pointer(p), &p.events.destroy)
	return SubCompositor{p: p}
}

func (d Display) NewCompositor(version int, renderer Renderer) Compositor {
	return d.CompositorCreate(version, renderer)
}

func (d Display) CompositorCreate(version int, renderer Renderer) Compositor {
	p := C.wlr_compositor_create(d.p, C.uint(version), renderer.p)
	man.track(unsafe.Pointer(p), &p.events.destroy)
	return Compositor{p: p}
}

func (d Display) NewDataDeviceManager() DataDeviceManager {
	return d.DataDeviceManagerCreate()
}

func (d Display) DataDeviceManagerCreate() DataDeviceManager {
	p := C.wlr_data_device_manager_create(d.p)
	man.track(unsafe.Pointer(p), &p.events.destroy)
	return DataDeviceManager{p: p}
}

func (d Display) NewSeat(name string) Seat {
	return d.SeatCreate(name)
}

func (d Display) SeatCreate(name string) Seat {
	s := C.CString(name)
	p := C.wlr_seat_create(d.p, s)
	C.free(unsafe.Pointer(s))
	man.track(unsafe.Pointer(p), &p.events.destroy)
	return Seat{p: p}
}

func (d Display) NewXDGShell(version int) XDGShell {
	return d.XDGShellCreate(version)
}

func (d Display) XDGShellCreate(version int) XDGShell {
	p := C.wlr_xdg_shell_create(d.p, C.uint(version))
	man.track(unsafe.Pointer(p), &p.events.destroy)
	return XDGShell{p: p}
}

func (d Display) Destroy() {
	C.wl_display_destroy(d.p)
}

func (d Display) OnDestroy(cb func(Display)) {
	l := man.add(unsafe.Pointer(d.p), nil, func(data unsafe.Pointer) {
		cb(d)
	})
	C.wl_display_add_destroy_listener(d.p, l.p)
}

func (d Display) Run() {
	C.wl_display_run(d.p)
}

func (d Display) Terminate() {
	C.wl_display_terminate(d.p)
}

func (d Display) EventLoop() EventLoop {
	p := C.wl_display_get_event_loop(d.p)
	evl := EventLoop{p: p}
	evl.OnDestroy(func(EventLoop) {
		man.delete(unsafe.Pointer(p))
	})
	return evl
}

func (d Display) AddSocketAuto() (string, error) {
	socket := C.wl_display_add_socket_auto(d.p)
	if socket == nil {
		return "", errors.New("can't auto add wayland socket")
	}

	return C.GoString(socket), nil
}

func (d Display) FlushClients() {
	C.wl_display_flush_clients(d.p)
}

func (d Display) DestroyClients() {
	C.wl_display_destroy_clients(d.p)
}

type DataDeviceManager struct {
	p *C.struct_wlr_data_device_manager
}

func (m DataDeviceManager) OnDestroy(cb func(DataDeviceManager)) {
	man.add(unsafe.Pointer(m.p), &m.p.events.destroy, func(unsafe.Pointer) {
		cb(m)
	})
}

type Color struct {
	R, G, B, A float32
}

func (c *Color) Set(r, g, b, a float32) {
	c.R = r
	c.G = g
	c.B = b
	c.A = a
}

func (c *Color) toC() C.struct_wlr_render_color {
	return C.struct_wlr_render_color{
		r: C.float(c.R),
		g: C.float(c.G),
		b: C.float(c.B),
		a: C.float(c.A),
	}
}

type GeoBox struct {
	X, Y, Width, Height int
}

func (b *GeoBox) Set(x, y, width, height int) {
	b.X = x
	b.Y = y
	b.Width = width
	b.Height = height
}

func (b *GeoBox) toC() C.struct_wlr_box {
	return C.struct_wlr_box{
		x:      C.int(b.X),
		y:      C.int(b.Y),
		width:  C.int(b.Width),
		height: C.int(b.Height),
	}
}

func (b *GeoBox) fromC(cb *C.struct_wlr_box) {
	b.X = int(cb.x)
	b.Y = int(cb.y)
	b.Width = int(cb.width)
	b.Height = int(cb.height)
}

type FBox struct {
	X, Y, Width, Height float64
}

func (b *FBox) Set(x, y, width, height float64) {
	b.X = x
	b.Y = y
	b.Width = width
	b.Height = height
}

func (b *FBox) toC() C.struct_wlr_fbox {
	return C.struct_wlr_fbox{
		x:      C.double(b.X),
		y:      C.double(b.Y),
		width:  C.double(b.Width),
		height: C.double(b.Height),
	}
}

func (b *FBox) fromC(cb *C.struct_wlr_box) {
	b.X = float64(cb.x)
	b.Y = float64(cb.y)
	b.Width = float64(cb.width)
	b.Height = float64(cb.height)
}
