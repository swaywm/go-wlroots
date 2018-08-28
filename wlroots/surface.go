package wlroots

// #include <time.h>
// #include <wlr/types/wlr_surface.h>
// #include <wlr/types/wlr_xdg_shell.h>
import "C"
import (
	"time"
	"unsafe"
)

type Surface struct {
	p *C.struct_wlr_surface
}

type SurfaceState struct {
	s C.struct_wlr_surface_state
}

func (s Surface) Nil() bool {
	return s.p == nil
}

func (s Surface) OnDestroy(cb func(Surface)) {
	man.add(unsafe.Pointer(s.p), &s.p.events.destroy, func(unsafe.Pointer) {
		cb(s)
	})
}

func (s Surface) Texture() Texture {
	p := C.wlr_surface_get_texture(s.p)
	return Texture{p: p}
}

func (s Surface) CurrentState() SurfaceState {
	return SurfaceState{s: s.p.current}
}

func (s Surface) Walk(visit func()) {
	panic("not implemented")
}

func (s Surface) SendFrameDone(when time.Time) {
	t := C.struct_timespec{}
	C.wlr_surface_send_frame_done(s.p, &t)
}

func (s Surface) XDGSurface() XDGSurface {
	p := C.wlr_xdg_surface_from_wlr_surface(s.p)
	return XDGSurface{p: p}
}

func (s SurfaceState) Width() int {
	return int(s.s.width)
}

func (s SurfaceState) Height() int {
	return int(s.s.height)
}

func (s SurfaceState) Transform() uint32 {
	return uint32(s.s.transform)
}
