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

// #cgo pkg-config: wlroots-0.18 wayland-server
// #cgo CFLAGS: -D_GNU_SOURCE -DWLR_USE_UNSTABLE
// #include <wlr/types/wlr_scene.h>
import "C"

/**
 * The scene-graph API provides a declarative way to display surfaces. The
 * compositor creates a scene, adds surfaces, then renders the scene on
 * outputs.
 *
 * The scene-graph API only supports basic 2D composition operations (like the
 * KMS API or the Wayland protocol does). For anything more complicated,
 * compositors need to implement custom rendering logic.
 */

/** The root scene-graph node. */
type Scene struct {
	p *C.struct_wlr_scene
}

func (s Scene) OutputCreate(o Output) SceneOutput {
	p := C.wlr_scene_output_create(s.p, o.p)
	return SceneOutput{p: p}
}

func (s Scene) NewOutput(o Output) SceneOutput {
	return s.OutputCreate(o)
}

/**
 * Create a new scene-graph.
 */

func NewScene() Scene {
	p := C.wlr_scene_create()
	return Scene{p: p}
}

func (s Scene) AttachOutputLayout(layout OutputLayout) SceneOutputLayout {
	p := C.wlr_scene_attach_output_layout(s.p, layout.p)
	return SceneOutputLayout{p: p}
}

func (s Scene) SceneOutput(o Output) (SceneOutput, error) {
	p := C.wlr_scene_get_scene_output(s.p, o.p)
	if p != nil {
		return SceneOutput{p: p}, nil
	}
	return SceneOutput{}, errors.New(" output hasn't been added to the scene-graph")
}

func (s Scene) Tree() SceneTree {
	return SceneTree{p: &s.p.tree}
}

/** A scene-graph node displaying a single surface. */
type SceneSurface struct {
	p *C.struct_wlr_scene_surface
}

func (s SceneSurface) Nil() bool {
	return s.p == nil
}

func (s SceneSurface) Surface() Surface {
	return Surface{p: s.p.surface}
}

/** A sub-tree in the scene-graph. */
type SceneTree struct {
	p *C.struct_wlr_scene_tree
}

func (st SceneTree) Nil() bool {
	return st.p == nil
}

/**
 * Add a node displaying nothing but its children.
 */
func (parent SceneTree) NewSceneTree() SceneTree {
	p := C.wlr_scene_tree_create(parent.p)
	return SceneTree{p: p}
}

/**
 * Add a node displaying a single surface to the scene-graph.
 *
 * The child sub-surfaces are ignored.
 *
 * wlr_surface_send_enter() and wlr_surface_send_leave() will be called
 * automatically based on the position of the surface and outputs in
 * the scene.
 */

func (parent SceneTree) NewSurface(surface Surface) SceneSurface {
	p := C.wlr_scene_surface_create(parent.p, surface.p)
	return SceneSurface{p: p}
}

func (st SceneTree) Node() SceneNode {
	return SceneNode{p: (*C.struct_wlr_scene_node)(&st.p.node)}
}

/**
 * Add a node displaying an xdg_surface and all of its sub-surfaces to the
 * scene-graph.
 *
 * The origin of the returned scene-graph node will match the top-left corner
 * of the xdg_surface window geometry.
 */
func (st SceneTree) XDGSurfaceCreate(s XDGSurface) SceneTree {
	p := C.wlr_scene_xdg_surface_create(st.p, s.p)
	return SceneTree{p: p}
}
func (st SceneTree) NewXDGSurface(s XDGSurface) SceneTree {
	return st.XDGSurfaceCreate(s)
}

func (parent SceneTree) BufferCreate(b Buffer) SceneBuffer {
	p := C.wlr_scene_buffer_create(parent.p, b.p)
	return SceneBuffer{p: p}
}
func (parent SceneTree) NewBuffer(b Buffer) SceneBuffer {
	return parent.BufferCreate(b)
}

/** A scene-graph node displaying a solid-colored rectangle */
type SceneRect struct {
	p *C.struct_wlr_scene_rect
}

/** A scene-graph node displaying a buffer */
type SceneBuffer struct {
	p *C.struct_wlr_scene_buffer
}

/**
 * If this buffer is backed by a surface, then the struct wlr_scene_surface is
 * returned. If not, NULL will be returned.
 */
func (sb SceneBuffer) SceneSurface() SceneSurface {
	p := C.wlr_scene_surface_try_from_buffer(sb.p)
	return SceneSurface{p: p}
}

/**
 * Calls the buffer's frame_done signal.
 */
func (sb SceneBuffer) SendFrameDone(when time.Time) {
	t, _ := unix.TimeToTimespec(when)
	C.wlr_scene_buffer_send_frame_done(sb.p, (*C.struct_timespec)(unsafe.Pointer(&t)))
}

func (sb SceneBuffer) SetBuffer(b Buffer) {
	C.wlr_scene_buffer_set_buffer(sb.p, b.p)
}

/** A viewport for an output in the scene-graph */
type SceneOutput struct {
	p *C.struct_wlr_scene_output
}

func (s SceneOutput) Commit() {
	C.wlr_scene_output_commit(s.p, nil)
}

func (s SceneOutput) Destroy() {
	C.wlr_scene_output_destroy(s.p)
}

func (s SceneOutput) SendFrameDone(when time.Time) {
	t, _ := unix.TimeToTimespec(when)
	C.wlr_scene_output_send_frame_done(s.p, (*C.struct_timespec)(unsafe.Pointer(&t)))
}

type SceneTimer struct {
	p *C.struct_wlr_scene_timer
}

type SceneOutputLayout struct {
	p *C.struct_wlr_scene_output_layout
}

func (sol SceneOutputLayout) AddOutput(l OutputLayoutOutput, o SceneOutput) {
	C.wlr_scene_output_layout_add_output(sol.p, l.p, o.p)
}

type SceneNodeType uint32

const (
	SceneNodeTree   SceneNodeType = C.WLR_SCENE_NODE_TREE
	SceneNodeRect   SceneNodeType = C.WLR_SCENE_NODE_RECT
	SceneNodeBuffer SceneNodeType = C.WLR_SCENE_NODE_BUFFER
)

/** A node is an object in the scene. */
type SceneNode struct {
	p *C.struct_wlr_scene_node
}

/**
 * If this node represents a wlr_scene_tree, that tree will be returned. It
 * is not legal to feed a node that does not represent a wlr_scene_tree.
 */
func (sn SceneNode) SceneTree() SceneTree {
	p := C.wlr_scene_tree_from_node(sn.p)
	return SceneTree{p: p}
}

/**
 * If this node represents a wlr_scene_rect, that rect will be returned. It
 * is not legal to feed a node that does not represent a wlr_scene_rect.
 */
func (sn SceneNode) SceneRect() SceneRect {
	p := C.wlr_scene_rect_from_node(sn.p)
	return SceneRect{p: p}
}

/**
 * If this node represents a wlr_scene_buffer, that buffer will be returned. It
 * is not legal to feed a node that does not represent a wlr_scene_buffer.
 */
func (sn SceneNode) SceneBuffer() SceneBuffer {
	p := C.wlr_scene_buffer_from_node(sn.p)
	return SceneBuffer{p: p}
}

/**
 * Immediately destroy the scene-graph node.
 */
func (sn SceneNode) Destroy() {
	C.wlr_scene_node_destroy(sn.p)
}

/**
 * Move the node below all of its sibling nodes.
 */
func (sn SceneNode) LowerToBottom() {
	C.wlr_scene_node_lower_to_bottom(sn.p)
}

/**
 * Move the node right above the specified sibling.
 * Asserts that node and sibling are distinct and share the same parent.
 */
func (sn SceneNode) PlaceAbove(sib SceneNode) {
	C.wlr_scene_node_place_above(sn.p, sib.p)
}

/**
 * Move the node right below the specified sibling.
 * Asserts that node and sibling are distinct and share the same parent.
 */
func (sn SceneNode) PlaceBellow(sib SceneNode) {
	C.wlr_scene_node_place_below(sn.p, sib.p)
}

/**
 * Move the node above all of its sibling nodes.
 */
func (sn SceneNode) RaiseToTop() {
	C.wlr_scene_node_raise_to_top(sn.p)
}

/**
 * Move the node to another location in the tree.
 */
func (sn SceneNode) Reparent(tree SceneTree) {
	C.wlr_scene_node_reparent(sn.p, tree.p)
}

/**
 * Enable or disable this node. If a node is disabled, all of its children are
 * implicitly disabled as well.
 */
func (sn SceneNode) SetEnabled(enabled bool) {
	C.wlr_scene_node_set_enabled(sn.p, C.bool(enabled))
}

/**
 * Set the position of the node relative to its parent.
 */
func (sn SceneNode) SetPosition(x float64, y float64) {
	C.wlr_scene_node_set_position(sn.p, C.int(x), C.int(y))
}

/**
 * Find the topmost node in this scene-graph that contains the point at the
 * given layout-local coordinates. (For surface nodes, this means accepting
 * input events at that point.) Returns the node and coordinates relative to the
 * returned node, or NULL if no node is found at that location.
 */
func (sn SceneNode) At(x float64, y float64) (SceneNode, float64, float64) {
	var lx *C.double = new(C.double)
	var ly *C.double = new(C.double)
	p := C.wlr_scene_node_at(sn.p, C.double(x), C.double(y), lx, ly)
	return SceneNode{p: p}, float64(*lx), float64(*ly)
}

func (sn SceneNode) Nil() bool {
	return sn.p == nil
}

func (sn SceneNode) Type() SceneNodeType {
	return SceneNodeType(sn.p._type)
}

func (sn SceneNode) Parent() SceneTree {
	return SceneTree{p: sn.p.parent}
}

// relative to parent
func (sn SceneNode) X() int {
	return int(sn.p.x)
}

// relative to parent
func (sn SceneNode) Y() int {
	return int(sn.p.y)
}

func (sn SceneNode) SetData(tree SceneTree) {
	sn.p.data = unsafe.Pointer(tree.p)
}

func (x SceneNode) SceneTreeFromData() SceneTree {
	// slog.Debug("XDGSurface SceneTree(): x.p", x.p)
	// slog.Debug("XDGSurface SceneTree(): x.p.data", x.p.data)
	return SceneTree{p: (*C.struct_wlr_scene_tree)(x.p.data)}
}
