package wlroots

/*
 * This an unstable interface of wlroots. No guarantees are made regarding the
 * future consistency of this API.
 */

import (
	"errors"
	"time"
	"unsafe"

	"golang.org/x/sys/unix"
)

// #cgo pkg-config: wlroots wayland-server
// #cgo CFLAGS: -D_GNU_SOURCE -DWLR_USE_UNSTABLE
// #include <stdlib.h>
// #include <time.h>
// #include <wlr/types/wlr_compositor.h>
// #include <wlr/types/wlr_subcompositor.h>
// #include <wlr/types/wlr_xdg_shell.h>
// #include <wlr/xwayland.h>
import "C"

type Compositor struct {
	p *C.struct_wlr_compositor
}

func (c Compositor) OnDestroy(cb func(Compositor)) {
	man.add(unsafe.Pointer(c.p), &c.p.events.destroy, func(unsafe.Pointer) {
		cb(c)
	})
}

type SubCompositor struct {
	p *C.struct_wlr_subcompositor
}

func (c SubCompositor) OnDestroy(cb func(SubCompositor)) {
	man.add(unsafe.Pointer(c.p), &c.p.events.destroy, func(unsafe.Pointer) {
		cb(c)
	})
}

type SurfaceType uint32

const (
	SurfaceTypeNone SurfaceType = iota
	SurfaceTypeXDG
	SurfaceTypeXWayland
)

type Surface struct {
	p *C.struct_wlr_surface
}

/**
 * Map the surface. If the surface is already mapped, this is no-op.
 *
 * This function must only be used by surface role implementations.
 */
func (s Surface) Map() {
	C.wlr_surface_map(s.p)
}

/**
 * Unmap the surface. If the surface is already unmapped, this is no-op.
 *
 * This function must only be used by surface role implementations.
 */
func (s Surface) Unmap() {
	C.wlr_surface_unmap(s.p)
}

/**
 * Whether or not this surface currently has an attached buffer. A surface has
 * an attached buffer when it commits with a non-null buffer in its pending
 * state. A surface will not have a buffer if it has never committed one, has
 * committed a null buffer, or something went wrong with uploading the buffer.
 */
func (s Surface) HasBuffer() bool {
	return bool(C.wlr_surface_has_buffer(s.p))
}

/**
 * Get the texture of the buffer currently attached to this surface. Returns
 * NULL if no buffer is currently attached or if something went wrong with
 * uploading the buffer.
 */
func (s Surface) Texture() Texture {
	p := C.wlr_surface_get_texture(s.p)
	return Texture{p}
}

/**
 * Get the root of the subsurface tree for this surface. Can return NULL if
 * a surface in the tree has been destroyed.
 */
func (s Surface) RootSurface() Surface {
	p := C.wlr_surface_get_root_surface(s.p)
	return Surface{p}
}

/**
 * Notify the client that the surface has entered an output.
 *
 * This is a no-op if the surface has already entered the output.
 */
func (s Surface) SendEnter(o Output) {
	C.wlr_surface_send_enter(s.p, o.p)
}

/**
 * Notify the client that the surface has left an output.
 *
 * This is a no-op if the surface has already left the output.
 */
func (s Surface) SendLeave(o Output) {
	C.wlr_surface_send_leave(s.p, o.p)
}

func (s Surface) Nil() bool {
	return s.p == nil
}

func (s Surface) OnDestroy(cb func(Surface)) {
	man.add(unsafe.Pointer(s.p), &s.p.events.destroy, func(unsafe.Pointer) {
		cb(s)
	})
}

func (s Surface) Type() SurfaceType {
	if C.wlr_xdg_surface_try_from_wlr_surface(s.p) != nil {
		return SurfaceTypeXDG
	} else if C.wlr_xwayland_surface_try_from_wlr_surface(s.p) != nil {
		return SurfaceTypeXWayland
	}
	return SurfaceTypeNone
}

func (s Surface) SurfaceAt(sx float64, sy float64) (surface Surface, subX float64, subY float64) {
	var csubX, csubY C.double
	p := C.wlr_surface_surface_at(s.p, C.double(sx), C.double(sy), &csubX, &csubY)
	return Surface{p: p}, float64(csubX), float64(csubY)
}

func (s Surface) CurrentState() SurfaceState {
	return SurfaceState{p: s.p.current}
}

func (s Surface) Walk(visit func()) {
	panic("not implemented")
}

/**
 * Complete the queued frame callbacks for this surface.
 *
 * This will send an event to the client indicating that now is a good time to
 * draw its next frame.
 */
func (s Surface) SendFrameDone(when time.Time) {
	// we ignore the returned error; the only possible error is
	// ERANGE, when timespec on a platform has int32 precision, but
	// our time requires 64 bits. This should not occur.
	t, _ := unix.TimeToTimespec(when)
	C.wlr_surface_send_frame_done(s.p, (*C.struct_timespec)(unsafe.Pointer(&t)))
}

func (s Surface) XDGSurface() XDGSurface {
	p := C.wlr_xdg_surface_try_from_wlr_surface(s.p)
	return XDGSurface{p: p}
}

func (s Surface) XDGTopLevel() (XDGTopLevel, error) {
	p := C.wlr_xdg_toplevel_try_from_wlr_surface(s.p)
	if p == nil {
		return XDGTopLevel{}, errors.New("no xdg top level")
	}
	return XDGTopLevel{p: p}, nil
}

/**
 * Get a struct wlr_xwayland_surface from a struct wlr_surface.
 *
 * If the surface hasn't been created by Xwayland or has no X11 window
 * associated, NULL is returned.
 */
func (s Surface) XWaylandSurface() XWaylandSurface {
	p := C.wlr_xwayland_surface_try_from_wlr_surface(s.p)
	return XWaylandSurface{p: p}
}

type SurfaceStateField uint32

const (
	SurfaceStateBuffer            SurfaceStateField = C.WLR_SURFACE_STATE_BUFFER
	SurfaceStateSurfaceDamage     SurfaceStateField = C.WLR_SURFACE_STATE_SURFACE_DAMAGE
	SurfaceStateBufferDamage      SurfaceStateField = C.WLR_SURFACE_STATE_BUFFER_DAMAGE
	SurfaceStateOpaqueRegion      SurfaceStateField = C.WLR_SURFACE_STATE_OPAQUE_REGION
	SurfaceStateInputRegion       SurfaceStateField = C.WLR_SURFACE_STATE_INPUT_REGION
	SurfaceStateTransform         SurfaceStateField = C.WLR_SURFACE_STATE_TRANSFORM
	SurfaceStateScale             SurfaceStateField = C.WLR_SURFACE_STATE_SCALE
	SurfaceStateFrameCallbackList SurfaceStateField = C.WLR_SURFACE_STATE_FRAME_CALLBACK_LIST
	SurfaceStateViewport          SurfaceStateField = C.WLR_SURFACE_STATE_VIEWPORT
	SurfaceStateOffset            SurfaceStateField = C.WLR_SURFACE_STATE_OFFSET
)

type SurfaceState struct {
	p C.struct_wlr_surface_state
}

func (s SurfaceState) Commited() SurfaceStateField {
	return SurfaceStateField(s.p.committed)
}

func (s SurfaceState) Buffer() Buffer {
	return Buffer{p: s.p.buffer}
}

func (s SurfaceState) Scale() int {
	return int(s.p.scale)
}

// relative to previous position
func (s SurfaceState) DX() int {
	return int(s.p.dx)
}

// relative to previous position
func (s SurfaceState) DY() int {
	return int(s.p.dy)
}

// in surface-local coordinates
func (s SurfaceState) Width() int {
	return int(s.p.width)
}

// in surface-local coordinates
func (s SurfaceState) Height() int {
	return int(s.p.height)
}

func (s SurfaceState) BufferWidth() int {
	return int(s.p.buffer_width)
}

func (s SurfaceState) BufferHeight() int {
	return int(s.p.buffer_height)
}

func (s SurfaceState) Transform() uint32 {
	return uint32(s.p.transform)
}

type SurfaceRole struct {
	p *C.struct_wlr_surface_role
}

func (s SurfaceRole) Name() string {
	return C.GoString(s.p.name)
}

/**
 * If true, the role isn't represented by any object.
 * For example, this applies to cursor surfaces.
 */
func (s SurfaceRole) NoObject() bool {
	return bool(s.p.no_object)
}

type SurfaceOutput struct {
	p *C.struct_wlr_surface_output
}

func (s SurfaceOutput) Surface() Surface {
	return Surface{p: s.p.surface}
}

func (s SurfaceOutput) Output() Output {
	return Output{p: s.p.output}
}
