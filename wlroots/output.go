package wlroots

/*
 * This an unstable interface of wlroots. No guarantees are made regarding the
 * future consistency of this API.
 */

import (
	"errors"
	"unsafe"
)

// #cgo pkg-config: wlroots wayland-server
// #cgo CFLAGS: -D_GNU_SOURCE -DWLR_USE_UNSTABLE
// #include <wlr/backend/wayland.h>
// #include <wlr/types/wlr_output.h>
// #include <wlr/backend/x11.h>
import "C"

/**
 * A compositor output region. This typically corresponds to a monitor that
 * displays part of the compositor space.
 *
 * The `frame` event will be emitted when it is a good time for the compositor
 * to submit a new frame.
 *
 * To render a new frame, compositors should call wlr_output_attach_render(),
 * render and call wlr_output_commit(). No rendering should happen outside a
 * `frame` event handler or before wlr_output_attach_render().
 */
type Output struct {
	p *C.struct_wlr_output
}

func wrapOutput(p unsafe.Pointer) Output {
	return Output{p: (*C.struct_wlr_output)(p)}
}

func (o Output) Name() string {
	return C.GoString(o.p.name)
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

func (o Output) OnRequestState(cb func(Output, OutputState)) {
	man.add(unsafe.Pointer(o.p), &o.p.events.request_state, func(data unsafe.Pointer) {
		cb(o, OutputState{p: (*C.struct_wlr_output_state)(data)})
	})
}

func (o Output) OnDestroy(cb func(Output)) {
	man.add(unsafe.Pointer(o.p), &o.p.events.destroy, func(data unsafe.Pointer) {
		cb(o)
	})
}

func (o Output) RenderSoftwareCursors() {
	C.wlr_output_render_software_cursors(o.p, nil)
}

/**
 * Computes the transformed output resolution.
 */
func (o Output) TransformedResolution() (int, int) {
	var width, height C.int
	C.wlr_output_transformed_resolution(o.p, &width, &height)
	return int(width), int(height)
}

/**
 * Computes the transformed and scaled output resolution.
 */
func (o Output) EffectiveResolution() (int, int) {
	var width, height C.int
	C.wlr_output_effective_resolution(o.p, &width, &height)
	return int(width), int(height)
}

/**
 * Attach the renderer's buffer to the output. Compositors must call this
 * function before rendering. After they are done rendering, they should call
 * wlr_output_commit() to submit the new frame. The output needs to be
 * enabled.
 *
 * If non-NULL, `buffer_age` is set to the drawing buffer age in number of
 * frames or -1 if unknown. This is useful for damage tracking.
 *
 * If the compositor decides not to render after calling this function, it
 * must call wlr_output_rollback().
 */
func (o Output) AttachRender() (int, error) {
	var bufferAge C.int
	if !C.wlr_output_attach_render(o.p, &bufferAge) {
		return 0, errors.New("can't make output context current")
	}

	return int(bufferAge), nil
}

/**
 * Schedule a done event.
 *
 * This is intended to be used by wl_output add-on interfaces.
 */
func (o Output) ScheduleDone() {
	C.wlr_output_schedule_done(o.p)
}

func (o Output) Destroy() {
	C.wlr_output_destroy(o.p)
}

/**
 * Test whether the pending output state would be accepted by the backend. If
 * this function returns true, wlr_output_commit() can only fail due to a
 * runtime error.
 *
 * This function doesn't mutate the pending state.
 */
func (o Output) Test() bool {
	return bool(C.wlr_output_test(o.p))
}

/**
 * Commit the pending output state. If wlr_output_attach_render() has been
 * called, the pending frame will be submitted for display and a `frame` event
 * will be scheduled.
 *
 * On failure, the pending changes are rolled back.
 */
func (o Output) Commit() bool {
	return bool(C.wlr_output_commit(o.p))
}

/**
 * Discard the pending output state.
 */
func (o Output) Rollback() {
	C.wlr_output_rollback(o.p)
}

/**
 * Test whether this output state would be accepted by the backend. If this
 * function returns true, wlr_output_commit_state() will only fail due to a
 * runtime error. This function does not change the current state of the
 * output.
 */
func (o Output) TestState(s OutputState) bool {
	return bool(C.wlr_output_test_state(o.p, s.p))
}

/**
 * Attempts to apply the state to this output. This function may fail for any
 * reason and return false. If failed, none of the state would have been applied,
 * this function is atomic. If the commit succeeded, true is returned.
 *
 * Note: wlr_output_state_finish() would typically be called after the state
 * has been committed.
 */
func (o Output) CommitState(s OutputState) bool {
	return bool(C.wlr_output_commit_state(o.p, s.p))
}

/**
 * Manually schedules a `frame` event. If a `frame` event is already pending,
 * it is a no-op.
 */
func (o Output) ScheduleFrame() {
	C.wlr_output_schedule_frame(o.p)
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

/**
 * Sets the output mode. The output needs to be enabled.
 *
 * Mode is double-buffered state, see wlr_output_commit().
 */
func (o Output) SetMode(mode OutputMode) {
	C.wlr_output_set_mode(o.p, mode.p)
}

/**
 * Returns the preferred mode for this output. If the output doesn't support
 * modes, returns NULL.
 */
func (o Output) PreferredMode() (OutputMode, error) {
	mode := C.wlr_output_preferred_mode(o.p)
	if mode == nil {
		return OutputMode{}, errors.New("no preferred mode")
	}
	return OutputMode{p: mode}, nil
}

/**
 * Sets a custom mode on the output.
 *
 * When the output advertises fixed modes, custom modes are not guaranteed to
 * work correctly, they may result in visual artifacts. If a suitable fixed mode
 * is available, compositors should prefer it and use wlr_output_set_mode()
 * instead of custom modes.
 *
 * Setting `refresh` to zero lets the backend pick a preferred value. The
 * output needs to be enabled.
 *
 * Custom mode is double-buffered state, see wlr_output_commit().
 */
func (o Output) SetCustomMode(width int, height int, refresh int) {
	C.wlr_output_set_custom_mode(o.p, C.int(width), C.int(height), C.int(refresh))
}

/**
 * Enables or disables adaptive sync (ie. variable refresh rate) on this
 * output. On some backends, this is just a hint and may be ignored.
 * Compositors can inspect `wlr_output.adaptive_sync_status` to query the
 * effective status. Backends that don't support adaptive sync will reject
 * the output commit.
 *
 * When enabled, compositors can submit frames a little bit later than the
 * deadline without dropping a frame.
 *
 * Adaptive sync is double-buffered state, see wlr_output_commit().
 */
func (o Output) EnableAdaptiveSync(enable bool) {
	C.wlr_output_enable_adaptive_sync(o.p, C.bool(enable))
}

/**
 * Sets a scale for the output.
 *
 * Scale is double-buffered state, see wlr_output_commit().
 */
func (o Output) SetScale(scale float32) {
	C.wlr_output_set_scale(o.p, C.float(scale))
}

/**
 * Set the output name.
 *
 * Output names are subject to the following rules:
 *
 * - Each output name must be unique.
 * - The name cannot change after the output has been advertised to clients.
 *
 * For more details, see the protocol documentation for wl_output.name.
 */
func (o Output) SetName(name string) {
	C.wlr_output_set_name(o.p, C.CString(name))
}

func (o Output) SetDescription(desc string) {
	C.wlr_output_set_description(o.p, C.CString(desc))
}

/**
 * Enables or disables the output. A disabled output is turned off and doesn't
 * emit `frame` events.
 *
 * Whether an output is enabled is double-buffered state, see
 * wlr_output_commit().
 */
func (o Output) Enable(enable bool) {
	C.wlr_output_enable(o.p, C.bool(enable))
}

func (o Output) Enabled() bool {
	return bool(o.p.enabled)
}

func (o Output) Refresh() int {
	return int(o.p.refresh)
}

func (o Output) CreateGlobal() {
	C.wlr_output_create_global(o.p)
}

func (o Output) DestroyGlobal() {
	C.wlr_output_destroy_global(o.p)
}

func (o Output) SetTitle(title string) error {
	if C.wlr_output_is_wl(o.p) {
		C.wlr_wl_output_set_title(o.p, C.CString(title))
	} else if C.wlr_output_is_x11(o.p) {
		C.wlr_x11_output_set_title(o.p, C.CString(title))
	} else {
		return errors.New("this output type cannot have a title")
	}

	return nil
}

/**
 * Initialize the output's rendering subsystem with the provided allocator and
 * renderer. After initialization, this function may invoked again to reinitialize
 * the allocator and renderer to different values.
 *
 * Call this function prior to any call to wlr_output_attach_render(),
 * wlr_output_commit() or wlr_output_cursor_create().
 *
 * The buffer capabilities of the provided must match the capabilities of the
 * output's backend. Returns false otherwise.
 */
func (o Output) InitRender(a Allocator, r Renderer) bool {
	return bool(C.wlr_output_init_render(o.p, a.p, r.p))
}

/**
 * Holds the double-buffered output state.
 */
type OutputState struct {
	p *C.struct_wlr_output_state
}

func NewOutputState() OutputState {
	return OutputState{p: &C.struct_wlr_output_state{}}
}

func (os OutputState) StateInit() {
	C.wlr_output_state_init(os.p)
}

func (os OutputState) StateSetEnabled(enabled bool) {
	C.wlr_output_state_set_enabled(os.p, C.bool(enabled))
}

func (os OutputState) SetMode(mode OutputMode) {
	C.wlr_output_state_set_mode(os.p, mode.p)
}

func (os OutputState) Finish() {
	C.wlr_output_state_finish(os.p)
}

func OutputTransformInvert(transform uint32) uint32 {
	return uint32(C.wlr_output_transform_invert(C.enum_wl_output_transform(transform)))
}

type OutputMode struct {
	p *C.struct_wlr_output_mode
}

func (m OutputMode) Width() int {
	return int(m.p.width)
}

func (m OutputMode) Height() int {
	return int(m.p.height)
}

// mHz
func (m OutputMode) Refresh() int {
	return int(m.p.refresh)
}

func (m OutputMode) Preferred() bool {
	return bool(m.p.preferred)
}

func (m OutputMode) PictureAspectRatio() OutputModeAspectRatio {
	return OutputModeAspectRatio(m.p.picture_aspect_ratio)
}

type OutputModeAspectRatio uint32

const (
	OutputModeAspectRatio_None    OutputModeAspectRatio = C.WLR_OUTPUT_MODE_ASPECT_RATIO_NONE
	OutputModeAspectRatio_4_3     OutputModeAspectRatio = C.WLR_OUTPUT_MODE_ASPECT_RATIO_4_3
	OutputModeAspectRatio_16_9    OutputModeAspectRatio = C.WLR_OUTPUT_MODE_ASPECT_RATIO_16_9
	OutputModeAspectRatio_64_27   OutputModeAspectRatio = C.WLR_OUTPUT_MODE_ASPECT_RATIO_64_27
	OutputModeAspectRatio_256_135 OutputModeAspectRatio = C.WLR_OUTPUT_MODE_ASPECT_RATIO_256_135
)

type OutputStateField uint32

const (
	OutputState_BUFFER                OutputStateField = C.WLR_OUTPUT_STATE_BUFFER
	OutputState_DAMAGE                OutputStateField = C.WLR_OUTPUT_STATE_DAMAGE
	OutputState_MODE                  OutputStateField = C.WLR_OUTPUT_STATE_MODE
	OutputState_ENABLED               OutputStateField = C.WLR_OUTPUT_STATE_ENABLED
	OutputState_SCALE                 OutputStateField = C.WLR_OUTPUT_STATE_SCALE
	OutputState_TRANSFORM             OutputStateField = C.WLR_OUTPUT_STATE_TRANSFORM
	OutputState_ADAPTIVE_SYNC_ENABLED OutputStateField = C.WLR_OUTPUT_STATE_ADAPTIVE_SYNC_ENABLED
	OutputState_GAMMA_LUT             OutputStateField = C.WLR_OUTPUT_STATE_GAMMA_LUT
	OutputState_RENDER_FORMAT         OutputStateField = C.WLR_OUTPUT_STATE_RENDER_FORMAT
	OutputState_SUBPIXEL              OutputStateField = C.WLR_OUTPUT_STATE_SUBPIXEL
	OutputState_LAYERS                OutputStateField = C.WLR_OUTPUT_STATE_LAYERS
)

type OutputSdaptiveSyncStatus uint32

const (
	OutputSdaptiveSync_Disabled OutputSdaptiveSyncStatus = C.WLR_OUTPUT_ADAPTIVE_SYNC_DISABLED
	OutputSdaptiveSync_Enabled  OutputSdaptiveSyncStatus = C.WLR_OUTPUT_ADAPTIVE_SYNC_ENABLED
)
