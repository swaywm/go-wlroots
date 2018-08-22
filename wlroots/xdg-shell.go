package wlroots

// #include <wlr/types/wlr_xdg_shell.h>
//
// void _wlr_xdg_surface_for_each_cb(struct wlr_surface *surface, int sx, int sy, void *data);
//
// static inline void _wlr_xdg_surface_for_each_surface(struct wlr_xdg_surface *surface, void *user_data) {
//		wlr_xdg_surface_for_each_surface(surface, &_wlr_xdg_surface_for_each_cb, user_data);
// }
import "C"
import (
	"unsafe"
)

type XDGSurfaceRole uint32

const (
	XDGSurfaceRoleNone     XDGSurfaceRole = C.WLR_XDG_SURFACE_ROLE_NONE
	XDGSurfaceRoleTopLevel XDGSurfaceRole = C.WLR_XDG_SURFACE_ROLE_TOPLEVEL
	XDGSurfaceRolePopup    XDGSurfaceRole = C.WLR_XDG_SURFACE_ROLE_POPUP
)

var (
	xdgSurfaceWalkers = map[*C.struct_wlr_xdg_surface]XDGSurfaceWalkFunc{}
)

type XDGShell struct {
	p *C.struct_wlr_xdg_shell
}

type XDGSurface struct {
	p *C.struct_wlr_xdg_surface
}

type XDGSurfaceWalkFunc func(surface Surface, sx int, sy int)

type XDGTopLevel struct {
	p *C.struct_wlr_xdg_toplevel
}

func NewXDGShell(display Display) XDGShell {
	p := C.wlr_xdg_shell_create(display.p)
	return XDGShell{p: p}
}

func (s XDGShell) OnNewSurface(cb func(surface XDGSurface)) {
	listener := NewListener(func(data unsafe.Pointer) {
		cb(XDGSurface{p: (*C.struct_wlr_xdg_surface)(data)})
	})

	C.wl_signal_add(&s.p.events.new_surface, listener.p)
}

//export _wlr_xdg_surface_for_each_cb
func _wlr_xdg_surface_for_each_cb(surface *C.struct_wlr_surface, sx C.int, sy C.int, data unsafe.Pointer) {
	cb := xdgSurfaceWalkers[(*C.struct_wlr_xdg_surface)(data)]
	if cb != nil {
		cb(Surface{p: surface}, int(sx), int(sy))
	}
}

func (s XDGSurface) Nil() bool {
	return s.p == nil
}

func (s XDGSurface) Walk(visit XDGSurfaceWalkFunc) {
	xdgSurfaceWalkers[s.p] = visit
	C._wlr_xdg_surface_for_each_surface(s.p, unsafe.Pointer(s.p))
	delete(xdgSurfaceWalkers, s.p)
}

func (s XDGSurface) Role() XDGSurfaceRole {
	return XDGSurfaceRole(s.p.role)
}

func (s XDGSurface) TopLevel() XDGTopLevel {
	p := *(*unsafe.Pointer)(unsafe.Pointer(&s.p.anon0[0]))
	return XDGTopLevel{p: (*C.struct_wlr_xdg_toplevel)(p)}
}

func (s XDGSurface) TopLevelSetActivated(activated bool) {
	C.wlr_xdg_toplevel_set_activated(s.p, C.bool(activated))
}

func (s XDGSurface) TopLevelSetSize(width uint32, height uint32) {
	C.wlr_xdg_toplevel_set_size(s.p, C.uint32_t(width), C.uint32_t(height))
}

func (s XDGSurface) Surface() Surface {
	return Surface{p: s.p.surface}
}

func (s XDGSurface) SurfaceAt(sx float64, sy float64) (surface Surface, subX float64, subY float64) {
	var csubX, csubY C.double
	p := C.wlr_xdg_surface_surface_at(s.p, C.double(sx), C.double(sy), &csubX, &csubY)
	return Surface{p: p}, float64(csubX), float64(csubY)
}

func (s XDGSurface) OnMap(cb func(XDGSurface)) {
	listener := NewListener(func(data unsafe.Pointer) {
		cb(s)
	})

	C.wl_signal_add(&s.p.events._map, listener.p)
}

func (s XDGSurface) OnUnmap(cb func(XDGSurface)) {
	listener := NewListener(func(data unsafe.Pointer) {
		cb(s)
	})

	C.wl_signal_add(&s.p.events.unmap, listener.p)
}

func (s XDGSurface) OnDestroy(cb func(XDGSurface)) {
	listener := NewListener(func(data unsafe.Pointer) {
		cb(s)
	})

	C.wl_signal_add(&s.p.events.destroy, listener.p)
}

func (s XDGSurface) Geometry() Box {
	var box Box
	C.wlr_xdg_surface_get_geometry(s.p, &box.b)
	return box
}

func (t XDGTopLevel) OnRequestMove(cb func(client SeatClient, serial uint32)) {
	listener := NewListener(func(data unsafe.Pointer) {
		event := (*C.struct_wlr_xdg_toplevel_move_event)(data)
		client := SeatClient{p: event.seat}
		cb(client, uint32(event.serial))
	})

	C.wl_signal_add(&t.p.events.request_move, listener.p)
}

func (t XDGTopLevel) OnRequestResize(cb func(client SeatClient, serial uint32, edges Edges)) {
	listener := NewListener(func(data unsafe.Pointer) {
		event := (*C.struct_wlr_xdg_toplevel_resize_event)(data)
		client := SeatClient{p: event.seat}
		cb(client, uint32(event.serial), Edges(event.edges))
	})

	C.wl_signal_add(&t.p.events.request_resize, listener.p)
}

func (s XDGTopLevel) Nil() bool {
	return s.p == nil
}

func (t XDGTopLevel) Title() string {
	return C.GoString(t.p.title)
}
