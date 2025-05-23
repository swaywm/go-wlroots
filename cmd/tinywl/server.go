package main

import (
	"container/list"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/swaywm/go-wlroots/wlroots"
	"github.com/swaywm/go-wlroots/xkb"
)

type CursorMode int

const (
	CursorModePassThrough CursorMode = iota
	CursorModeMove
	CursorModeResize
)

type Server struct {
	display     wlroots.Display
	backend     wlroots.Backend
	renderer    wlroots.Renderer
	allocator   wlroots.Allocator
	scene       wlroots.Scene
	sceneLayout wlroots.SceneOutputLayout

	xdgShell     wlroots.XDGShell
	topLevelList list.List

	cursor    wlroots.Cursor
	cursorMgr wlroots.XCursorManager

	seat            wlroots.Seat
	keyboards       []*Keyboard
	cursorMode      CursorMode
	grabbedTopLevel *wlroots.XDGTopLevel
	grabX, grabY    float64
	grabGeobox      wlroots.GeoBox
	resizeEdges     wlroots.Edges

	outputLayout wlroots.OutputLayout
}

type Keyboard struct {
	dev wlroots.InputDevice
}

func (s *Server) inTopLevel(topLevel *wlroots.XDGTopLevel) *list.Element {
	for e := s.topLevelList.Front(); e != nil; e = e.Next() {
		if *e.Value.(*wlroots.XDGTopLevel) == *topLevel {
			return e
		}
	}
	return nil
}

func (s *Server) moveFrontTopLevel(topLevel *wlroots.XDGTopLevel) {
	slog.Debug("moveFrontTopLevel", "s.topLevelList.Len()", s.topLevelList.Len())
	e := s.inTopLevel(topLevel)
	if e != nil {
		slog.Debug("moveFrontTopLevel", "topLevel", topLevel)
		s.topLevelList.MoveToFront(e)
	}
	slog.Debug("moveFrontTopLevel", "s.topLevelList.Len()", s.topLevelList.Len())
}

func (s *Server) removeTopLevel(topLevel *wlroots.XDGTopLevel) {
	slog.Debug("removeTopLevel", "s.topLevelList.Len()", s.topLevelList.Len())
	e := s.inTopLevel(topLevel)
	if e != nil {
		slog.Debug("removeTopLevel", "topLevel", topLevel)
		s.topLevelList.Remove(e)
	}
	slog.Debug("removeTopLevel", "s.topLevelList.Len()", s.topLevelList.Len())
}

func (s *Server) focusTopLevel(topLevel *wlroots.XDGTopLevel, surface *wlroots.Surface) {
	/* Note: this function only deals with keyboard focus. */
	if topLevel == nil {
		return
	}
	prevSurface := s.seat.KeyboardState().FocusedSurface()
	slog.Debug("focusTopLevel", "prev surface:", prevSurface)
	slog.Debug("focusTopLevel", "current surface:", *surface)
	if prevSurface == *surface {
		/* Don't re-focus an already focused surface. */
		return
	}

	if !prevSurface.Nil() {
		/*
		 * Deactivate the previously focused surface. This lets the client know
		 * it no longer has focus and the client will repaint accordingly, e.g.
		 * stop displaying a caret.
		 */
		prevTopLevel, err := prevSurface.XDGTopLevel()
		if err == nil {
			prevTopLevel.SetActivated(false)
		}
	}

	/* Move the toplevel to the front */
	topLevel.Base().SceneTree().Node().RaiseToTop()
	slog.Debug("focusTopLevel", "s.topLevelList.Len()", s.topLevelList.Len())
	slog.Debug("focusTopLevel", "topLevel", topLevel)
	s.moveFrontTopLevel(topLevel)
	slog.Debug("focusTopLevel", "s.topLevelList.Len()", s.topLevelList.Len())
	/* Activate the new surface */
	topLevel.SetActivated(true)
	/*
	 * Tell the seat to have the keyboard enter this surface. wlroots will keep
	 * track of this and automatically send key events to the appropriate
	 * clients without additional work on your part.
	 */
	s.seat.NotifyKeyboardEnter(topLevel.Base().Surface(), s.seat.Keyboard())
}

func (s *Server) handleNewPointer(dev wlroots.InputDevice) {
	/* We don't do anything special with pointers. All of our pointer handling
	 * is proxied through wlr_cursor. On another compositor, you might take this
	 * opportunity to do libinput configuration on the device to set
	 * acceleration, etc. */
	s.cursor.AttachInputDevice(dev)
}

func (s *Server) handleKey(keyboard wlroots.Keyboard, time uint32, keyCode uint32, updateState bool, state wlroots.KeyState) {
	/* This event is raised when a key is pressed or released. */

	// translate libinput keycode to xkbcommon and obtain keysyms
	syms := keyboard.XKBState().Syms(xkb.KeyCode(keyCode + 8))

	handled := false
	modifiers := keyboard.Modifiers()
	if (modifiers&wlroots.KeyboardModifierAlt != 0) && state == wlroots.KeyStatePressed {
		/* If alt is held down and this button was _pressed_, we attempt to
		 * process it as a compositor keybinding. */
		for _, sym := range syms {
			handled = s.handleKeyBinding(sym)
		}
	}

	if !handled {
		/* Otherwise, we pass it along to the client. */
		s.seat.SetKeyboard(keyboard.Base())
		s.seat.NotifyKeyboardKey(time, keyCode, state)
	}
}

func (s *Server) handleNewKeyboard(dev wlroots.InputDevice) {
	keyboard := dev.Keyboard()

	/* We need to prepare an XKB keymap and assign it to the keyboard. This
	 * assumes the defaults (e.g. layout = "us"). */
	context := xkb.NewContext(xkb.KeySymFlagNoFlags)
	keymap := context.KeyMap()
	keyboard.SetKeymap(keymap)
	keymap.Destroy()
	context.Destroy()
	keyboard.SetRepeatInfo(25, 600)

	/* Here we set up listeners for keyboard events. */
	keyboard.OnModifiers(func(keyboard wlroots.Keyboard) {
		/* This event is raised when a modifier key, such as shift or alt, is
		* pressed. We simply communicate this to the client. */
		s.seat.SetKeyboard(dev)
		s.seat.NotifyKeyboardModifiers(keyboard)
	})
	keyboard.OnKey(s.handleKey)

	s.seat.SetKeyboard(dev)

	/* And add the keyboard to our list of keyboards */
	s.keyboards = append(s.keyboards, &Keyboard{dev: dev})
}

func (s *Server) handleNewInput(dev wlroots.InputDevice) {
	/* This event is raised by the backend when a new input device becomes
	 * available. */
	switch dev.Type() {
	case wlroots.InputDeviceTypePointer:
		s.handleNewPointer(dev)
	case wlroots.InputDeviceTypeKeyboard:
		s.handleNewKeyboard(dev)
	}

	/* We need to let the wlr_seat know what our capabilities are, which is
	 * communicated to the client. In TinyWL we always have a cursor, even if
	 * there are no pointer devices, so we always include that capability. */
	caps := wlroots.SeatCapabilityPointer
	if len(s.keyboards) > 0 {
		caps |= wlroots.SeatCapabilityKeyboard
	}
	s.seat.SetCapabilities(caps)
}

func (s *Server) topLevelAt(lx float64, ly float64) (*wlroots.XDGTopLevel, *wlroots.Surface, float64, float64) {
	/* This returns the topmost node in the scene at the given layout coords.
	 * We only care about surface nodes as we are specifically looking for a
	 * surface in the surface tree of a tinywl_toplevel. */

	node, sx, sy := s.scene.Tree().Node().At(lx, ly)

	if node.Nil() || node.Type() != wlroots.SceneNodeBuffer {
		return nil, nil, 0, 0
	}
	sceneSurface := node.SceneBuffer().SceneSurface()
	slog.Debug("topLevelAt", "sceneSurface:", sceneSurface)
	if sceneSurface.Nil() {
		return nil, nil, 0, 0
	}
	surface := sceneSurface.Surface()
	slog.Debug("topLevelAt", "surface:", surface)

	/* Find the node corresponding to the tinywl_toplevel at the root of this
	 * surface tree, it is the only one for which we set the data field. */

	xdgSurface := surface.XDGSurface()
	if xdgSurface.Nil() {
		return nil, nil, 0, 0
	}
	topLevel := xdgSurface.TopLevel()
	slog.Debug("topLevelAt", "topLevel", topLevel)
	slog.Debug("topLevelAt", "s.topLevelList.Len()", s.topLevelList.Len())

	if s.inTopLevel(&topLevel) != nil {
		return &topLevel, &surface, sx, sy
	} else {
		return nil, &surface, sx, sy
	}
}

func (s *Server) handleNewFrame(output wlroots.Output) {
	/* This function is called every time an output is ready to display a frame,
	 * generally at the output's refresh rate (e.g. 60Hz). */

	sOut, err := s.scene.SceneOutput(output)
	if err != nil {
		return
	}

	/* Render the scene if needed and commit the output */
	sOut.Commit()
	sOut.SendFrameDone(time.Now())
}

func (s *Server) handleOutputRequestState(output wlroots.Output, state wlroots.OutputState) {
	/* This function is called when the backend requests a new state for
	 * the output. For example, Wayland and X11 backends request a new mode
	 * when the output window is resized. */
	slog.Debug("handleRequestState", "output", output, "state", state)
	output.CommitState(state)
}

func (s *Server) handleOutputDestroy(output wlroots.Output) {
	slog.Debug("handleDestroy", "output", output)
}

func (s *Server) handleNewOutput(output wlroots.Output) {
	/* This event is raised by the backend when a new output (aka a display or
	 * monitor) becomes available. */

	/* Configures the output created by the backend to use our allocator
	 * and our renderer. Must be done once, before commiting the output */
	output.InitRender(s.allocator, s.renderer)

	/* The output may be disabled, switch it on. */
	oState := wlroots.NewOutputState()
	oState.StateInit()
	oState.StateSetEnabled(true)

	/* Some backends don't have modes. DRM+KMS does, and we need to set a mode
	 * before we can use the output. The mode is a tuple of (width, height,
	 * refresh rate), and each monitor supports only a specific set of modes. We
	 * just pick the monitor's preferred mode, a more sophisticated compositor
	 * would let the user configure it. */
	mode, err := output.PreferredMode()
	if err == nil {
		oState.SetMode(mode)
	}

	/* Atomically applies the new output state. */
	output.CommitState(oState)
	oState.Finish()

	/* Sets up a listener for the frame event. */
	output.OnFrame(s.handleNewFrame)

	/* Sets up a listener for the state request event. */
	output.OnRequestState(s.handleOutputRequestState)

	/* Sets up a listener for the destroy event. */
	output.OnDestroy(s.handleOutputDestroy)

	/* Adds this to the output layout. The add_auto function arranges outputs
	 * from left-to-right in the order they appear. A more sophisticated
	 * compositor would let the user configure the arrangement of outputs in the
	 * layout.
	 *
	 * The output layout utility automatically adds a wl_output global to the
	 * display, which Wayland clients can see to find out information about the
	 * output (such as DPI, scale factor, manufacturer, etc).
	 */
	lOutput := s.outputLayout.AddOutputAuto(output)
	sceneOutput := s.scene.NewOutput(output)
	s.sceneLayout.AddOutput(lOutput, sceneOutput)

	err = output.SetTitle(fmt.Sprintf("tinywl (go-wlroots) - %s", output.Name()))
	if err != nil {
		return
	}
}

func (s *Server) handleCursorMotion(dev wlroots.InputDevice, time uint32, dx float64, dy float64) {
	/* This event is forwarded by the cursor when a pointer emits a _relative_
	 * pointer motion event (i.e. a delta) */

	/* The cursor doesn't move unless we tell it to. The cursor automatically
	 * handles constraining the motion to the output layout, as well as any
	 * special configuration applied for the specific input device which
	 * generated the event. You can pass NULL for the device if you want to move
	 * the cursor around without any input. */
	s.cursor.Move(dev, dx, dy)
	s.processCursorMotion(time)
}

func (s *Server) handleCursorMotionAbsolute(dev wlroots.InputDevice, time uint32, x float64, y float64) {
	/* This event is forwarded by the cursor when a pointer emits an _absolute_
	 * motion event, from 0..1 on each axis. This happens, for example, when
	 * wlroots is running under a Wayland window rather than KMS+DRM, and you
	 * move the mouse over the window. You could enter the window from any edge,
	 * so we have to warp the mouse there. There is also some hardware which
	 * emits these events. */
	s.cursor.WarpAbsolute(dev, x, y)
	s.processCursorMotion(time)
}

func (s *Server) processCursorMotion(time uint32) {
	/* If the mode is non-passthrough, delegate to those functions. */
	if s.cursorMode == CursorModeMove {
		s.processCursorMove(time)
		return
	} else if s.cursorMode == CursorModeResize {
		s.processCursorResize(time)
		return
	}

	/* Otherwise, find the toplevel under the pointer and send the event along. */
	topLevel, surface, sx, sy := s.topLevelAt(s.cursor.X(), s.cursor.Y())
	if topLevel == nil {
		/* If there's no toplevel under the cursor, set the cursor image to a
		 * default. This is what makes the cursor image appear when you move it
		 * around the screen, not over any toplevels. */
		s.cursor.SetXCursor(s.cursorMgr, "default")
	}
	if surface != nil {
		/*
		 * Send pointer enter and motion events.
		 *
		 * The enter event gives the surface "pointer focus", which is distinct
		 * from keyboard focus. You get pointer focus by moving the pointer over
		 * a window.
		 *
		 * Note that wlroots will avoid sending duplicate enter/motion events if
		 * the surface has already has pointer focus or if the client is already
		 * aware of the coordinates passed.
		 */
		s.seat.NotifyPointerEnter(*surface, sx, sy)
		s.seat.NotifyPointerMotion(time, sx, sy)
	} else {
		/* Clear pointer focus so future button events and such are not sent to
		 * the last client to have the cursor over it. */
		s.seat.ClearPointerFocus()
	}
}

func (s *Server) processCursorMove(_ uint32) {
	/* Move the grabbed toplevel to the new position. */
	s.grabbedTopLevel.Base().SceneTree().Node().SetPosition(s.cursor.X()-s.grabX, s.cursor.Y()-s.grabY)
}

func (s *Server) processCursorResize(_ uint32) {
	/*
	 * Resizing the grabbed toplevel can be a little bit complicated, because we
	 * could be resizing from any corner or edge. This not only resizes the
	 * toplevel on one or two axes, but can also move the toplevel if you resize
	 * from the top or left edges (or top-left corner).
	 *
	 * Note that some shortcuts are taken here. In a more fleshed-out
	 * compositor, you'd wait for the client to prepare a buffer at the new
	 * size, then commit any movement that was prepared.
	 */

	// borderX := s.cursor.X() - s.grabX
	// borderY := s.cursor.Y() - s.grabY
	borderX := s.cursor.X()
	borderY := s.cursor.Y()
	nLeft := s.grabGeobox.X
	nRight := s.grabGeobox.X + s.grabGeobox.Width
	nTop := s.grabGeobox.Y
	nBottom := s.grabGeobox.Y + s.grabGeobox.Height

	if s.resizeEdges&wlroots.EdgeTop != 0 {
		nTop = int(borderY)
		if nTop >= nBottom {
			nTop = nBottom - 1
		}
	} else if s.resizeEdges&wlroots.EdgeBottom != 0 {
		nBottom = int(borderY)
		if nBottom <= nTop {
			nBottom = nTop + 1
		}
	}

	if s.resizeEdges&wlroots.EdgeLeft != 0 {
		nLeft = int(borderX)
		if nLeft >= nRight {
			nLeft = nRight - 1
		}
	} else if s.resizeEdges&wlroots.EdgeRight != 0 {
		nRight = int(borderX)
		if nRight <= nLeft {
			nRight = nLeft + 1
		}
	}

	nWidth := nRight - nLeft
	nHeight := nBottom - nTop
	s.grabbedTopLevel.Base().TopLevelSetSize(uint32(nWidth), uint32(nHeight))
}

func (s *Server) handleSetCursorRequest(client wlroots.SeatClient, surface wlroots.Surface, _ uint32, hotspotX int32, hotspotY int32) {
	/* This event is raised by the seat when a client provides a cursor image */

	focusedClient := s.seat.PointerState().FocusedClient()

	/* This can be sent by any client, so we check to make sure this one is
	 * actually has pointer focus first. */
	if focusedClient == client {
		/* Once we've vetted the client, we can tell the cursor to use the
		 * provided surface as the cursor image. It will set the hardware cursor
		 * on the output that it's currently on and continue to do so as the
		 * cursor moves between outputs. */
		s.cursor.SetSurface(surface, hotspotX, hotspotY)
	}
}

func (s *Server) resetCursorMode() {
	/* Reset the cursor mode to passthrough. */
	s.cursorMode = CursorModePassThrough
	s.grabbedTopLevel = nil
}

func (s *Server) handleCursorButton(_ wlroots.InputDevice, time uint32, button uint32, state wlroots.ButtonState) {
	/* This event is forwarded by the cursor when a pointer emits a button
	 * event. */

	/* Notify the client with pointer focus that a button press has occurred */
	s.seat.NotifyPointerButton(time, button, state)

	if state == wlroots.ButtonStateReleased {
		/* If you released any buttons, we exit interactive move/resize mode. */
		s.resetCursorMode()
	} else {
		topLevel, surface, _, _ := s.topLevelAt(s.cursor.X(), s.cursor.Y())
		slog.Debug("handleCursorButton", "surface", surface)
		slog.Debug("handleCursorButton", "topLevel", topLevel)
		/* Focus that client if the button was _pressed_ */
		s.focusTopLevel(topLevel, surface)
	}
}

func (s *Server) handleCursorAxis(_ wlroots.InputDevice, time uint32, source wlroots.AxisSource, orientation wlroots.AxisOrientation, delta float64, deltaDiscrete int32) {
	/* This event is forwarded by the cursor when a pointer emits an axis event,
	 * for example when you move the scroll wheel. */

	/* Notify the client with pointer focus of the axis event. */
	s.seat.NotifyPointerAxis(time, orientation, delta, deltaDiscrete, source, wlroots.RelativeDirectionIdentical)
}

func (s *Server) handleCursorFrame() {
	/* This event is forwarded by the cursor when a pointer emits an frame
	 * event. Frame events are sent after regular pointer events to group
	 * multiple events together. For instance, two axis events may happen at the
	 * same time, in which case a frame event won't be sent in between. */

	/* Notify the client with pointer focus of the frame event. */
	s.seat.NotifyPointerFrame()
}

func (s *Server) handleKeyBinding(sym xkb.KeySym) bool {
	/*
	 * Here we handle compositor keybindings. This is when the compositor is
	 * processing keys, rather than passing them on to the client for its own
	 * processing.
	 *
	 * This function assumes Alt is held down.
	 */
	switch sym {
	case xkb.KeySymEscape:
		s.display.Terminate()
	case xkb.KeySymF1:
		/* Cycle to the next toplevel */
		if s.topLevelList.Len() < 2 {
			break
		}
		// focus the next view
		nextView := s.topLevelList.Front().Next().Value.(*wlroots.XDGTopLevel)
		nextSurface := nextView.Base().Surface()
		s.focusTopLevel(nextView, &nextSurface)
	default:
		return false
	}
	return true
}

func (s *Server) handleMapXDGToplevel(xdgSurface wlroots.XDGSurface) {
	/* Called when the surface is mapped, or ready to display on-screen. */

	topLevel := xdgSurface.TopLevel()
	surface := xdgSurface.Surface()
	slog.Debug("handleMapXDGToplevel", "topLevel", topLevel)
	slog.Debug("handleMapXDGToplevel", "s.topLevelList.Len()", s.topLevelList.Len())
	s.topLevelList.PushFront(&topLevel)
	slog.Debug("handleMapXDGToplevel", "s.topLevelList.Len()", s.topLevelList.Len())
	s.focusTopLevel(&topLevel, &surface)
	slog.Debug("handleMapXDGToplevel", "s.topLevelList.Len()", s.topLevelList.Len())
}

func (s *Server) handleUnMapXDGToplevel(xdgSurface wlroots.XDGSurface) {
	/* Called when the surface is unmapped, and should no longer be shown. */

	/* Reset the cursor mode if the grabbed toplevel was unmapped. */
	topLevel := xdgSurface.TopLevel()
	if s.grabbedTopLevel != nil && topLevel == *s.grabbedTopLevel {
		s.resetCursorMode()
	}
	s.removeTopLevel(&topLevel)
}
func (s *Server) handleNewXDGPopup(popup wlroots.XDGPopup) {
	xdgSurface := popup.Base()
	parent := popup.Parent()
	if parent.Nil() {
		panic("xdgSurface popup parent is nil")
	}
	xdgSurface.SetData(parent.XDGSurface().SceneTree().NewXDGSurface(xdgSurface))

	xdgSurface.OnCommit(func(surface wlroots.XDGSurface) {
		if surface.InitialCommit() {
			surface.ScheduleConfigure()
		}
	})
}

func (s *Server) handleNewXDGTopLevel(toplevel wlroots.XDGTopLevel) {
	xdgSurface := toplevel.Base()
	xdgSurface.SetData(s.scene.Tree().NewXDGSurface(xdgSurface.TopLevel().Base()))
	xdgSurface.OnMap(s.handleMapXDGToplevel)
	xdgSurface.OnUnmap(s.handleUnMapXDGToplevel)
	xdgSurface.OnDestroy(func(surface wlroots.XDGSurface) {})
	xdgSurface.OnCommit(func(surface wlroots.XDGSurface) {
		if surface.InitialCommit() {
			surface.ScheduleConfigure()
		}
	})

	toplevel.OnRequestMove(func(client wlroots.SeatClient, serial uint32) {
		s.beginInteractive(&toplevel, CursorModeMove, 0)
	})
	toplevel.OnRequestResize(func(client wlroots.SeatClient, serial uint32, edges wlroots.Edges) {
		s.beginInteractive(&toplevel, CursorModeResize, edges)
	})
}

func (s *Server) beginInteractive(topLevel *wlroots.XDGTopLevel, mode CursorMode, edges wlroots.Edges) {
	/* This function sets up an interactive move or resize operation, where the
	 * compositor stops propegating pointer events to clients and instead
	 * consumes them itself, to move or resize windows. */
	if topLevel.Base().Surface() != s.seat.PointerState().FocusedSurface() {
		/* Deny move/resize requests from unfocused clients. */
		return
	}
	s.grabbedTopLevel = topLevel
	s.cursorMode = mode

	if mode == CursorModeMove {
		s.grabX = s.cursor.X() - float64(topLevel.Base().SceneTree().Node().X())
		s.grabY = s.cursor.Y() - float64(topLevel.Base().SceneTree().Node().Y())
	} else {
		box := topLevel.Base().Geometry()
		r := func() int {
			if edges&wlroots.EdgeRight != 0 {
				return box.Width
			} else {
				return 0
			}
		}()
		b := func() int {
			if edges&wlroots.EdgeBottom != 0 {
				return box.Height
			} else {
				return 0
			}
		}()
		borderX := (topLevel.Base().SceneTree().Node().X() + box.X) + r
		borderY := (topLevel.Base().SceneTree().Node().Y() + box.Y) + b
		s.grabX = s.cursor.X() + float64(borderX)
		s.grabY = s.cursor.Y() + float64(borderY)
		s.grabGeobox = box
		s.grabGeobox.X += topLevel.Base().SceneTree().Node().X()
		s.grabGeobox.Y += topLevel.Base().SceneTree().Node().Y()

		s.resizeEdges = edges
	}
}

func NewServer() (s *Server, err error) {
	s = new(Server)

	/* The Wayland display is managed by libwayland. It handles accepting
	 * clients from the Unix socket, manging Wayland globals, and so on. */
	s.display = wlroots.NewDisplay()

	/* The backend is a wlroots feature which abstracts the underlying input and
	 * output hardware. The autocreate option will choose the most suitable
	 * backend based on the current environment, such as opening an X11 window
	 * if an X11 server is running. */
	s.backend, err = s.display.BackendAutocreate()
	if err != nil {
		return nil, err
	}

	/* Autocreates a renderer, either Pixman, GLES2 or Vulkan for us. The user
	 * can also specify a renderer using the WLR_RENDERER env var.
	 * The renderer is responsible for defining the various pixel formats it
	 * supports for shared memory, this configures that for clients. */
	s.renderer, err = s.backend.RendererAutoCreate()
	if err != nil {
		return nil, err
	}
	s.renderer.InitDisplay(s.display)

	/* Autocreates an allocator for us.
	 * The allocator is the bridge between the renderer and the backend. It
	 * handles the buffer creation, allowing wlroots to render onto the
	 * screen */
	s.allocator, err = s.backend.AllocatorAutocreate(s.renderer)
	if err != nil {
		return nil, err
	}

	/* This creates some hands-off wlroots interfaces. The compositor is
	 * necessary for clients to allocate surfaces, the subcompositor allows to
	 * assign the role of subsurfaces to surfaces and the data device manager
	 * handles the clipboard. Each of these wlroots interfaces has room for you
	 * to dig your fingers in and play with their behavior if you want. Note that
	 * the clients cannot set the selection directly without compositor approval,
	 * see the handling of the request_set_selection event below.*/
	s.display.CompositorCreate(5, s.renderer)
	s.display.SubCompositorCreate()
	s.display.DataDeviceManagerCreate()

	/* Creates an output layout, which a wlroots utility for working with an
	 * arrangement of screens in a physical layout. */
	s.outputLayout = wlroots.NewOutputLayout(s.display)

	/* Configure a listener to be notified when new outputs are available on the
	 * backend. */
	s.backend.OnNewOutput(s.handleNewOutput)

	/* Create a scene graph. This is a wlroots abstraction that handles all
	 * rendering and damage tracking. All the compositor author needs to do
	 * is add things that should be rendered to the scene graph at the proper
	 * positions and then call wlr_scene_output_commit() to render a frame if
	 * necessary.
	 */
	s.scene = wlroots.NewScene()
	s.sceneLayout = s.scene.AttachOutputLayout(s.outputLayout)

	/* Set up xdg-shell version 3. The xdg-shell is a Wayland protocol which is
	 * used for application windows. For more detail on shells, refer to
	 * https://drewdevault.com/2018/07/29/Wayland-shells.html.
	 */
	s.topLevelList.Init()
	s.xdgShell = s.display.XDGShellCreate(3)
	s.xdgShell.OnNewSurface(func(xdgSurface wlroots.XDGSurface) {})
	s.xdgShell.OnNewTopLevel(s.handleNewXDGTopLevel)
	s.xdgShell.OnNewPopup(s.handleNewXDGPopup)

	/*
	 * Creates a cursor, which is a wlroots utility for tracking the cursor
	 * image shown on screen.
	 */
	s.cursor = wlroots.NewCursor()
	s.cursor.AttachOutputLayout(s.outputLayout)

	/* Creates an xcursor manager, another wlroots utility which loads up
	 * Xcursor themes to source cursor images from and makes sure that cursor
	 * images are available at all scale factors on the screen (necessary for
	 * HiDPI support). */
	s.cursorMgr = wlroots.NewXCursorManager("", 24)

	/*
	 * wlr_cursor *only* displays an image on screen. It does not move around
	 * when the pointer moves. However, we can attach input devices to it, and
	 * it will generate aggregate events for all of them. In these events, we
	 * can choose how we want to process them, forwarding them to clients and
	 * moving the cursor around. More detail on this process is described in
	 * https://drewdevault.com/2018/07/17/Input-handling-in-wlroots.html.
	 *
	 * And more comments are sprinkled throughout the notify functions above.
	 */
	s.cursorMode = CursorModePassThrough
	s.cursor.OnMotion(s.handleCursorMotion)
	s.cursor.OnMotionAbsolute(s.handleCursorMotionAbsolute)
	s.cursor.OnButton(s.handleCursorButton)
	s.cursor.OnAxis(s.handleCursorAxis)
	s.cursor.OnFrame(s.handleCursorFrame)
	s.cursorMgr.Load(1)

	/*
	 * Configures a seat, which is a single "seat" at which a user sits and
	 * operates the computer. This conceptually includes up to one keyboard,
	 * pointer, touch, and drawing tablet device. We also rig up a listener to
	 * let us know when new input devices are available on the backend.
	 */
	s.backend.OnNewInput(s.handleNewInput)
	s.seat = s.display.SeatCreate("seat0")
	s.seat.OnSetCursorRequest(s.handleSetCursorRequest)

	return
}

func (s *Server) Start() (err error) {

	var socket string
	/* Add a Unix socket to the Wayland display. */
	if socket, err = s.display.AddSocketAuto(); err != nil {
		s.backend.Destroy()
		return
	}

	/* Start the backend. This will enumerate outputs and inputs, become the DRM
	 * master, etc */
	if err = s.backend.Start(); err != nil {
		s.backend.Destroy()
		s.display.Destroy()
		return
	}

	/* Set the WAYLAND_DISPLAY environment variable to our socket and run the
	 * startup command if requested. */
	if err = os.Setenv("WAYLAND_DISPLAY", socket); err != nil {
		return
	}

	slog.Info(fmt.Sprintf("Running Wayland compositor on WAYLAND_DISPLAY=%s", socket))
	return
}

func (s *Server) Run() error {

	/* Run the Wayland event loop. This does not return until you exit the
	 * compositor. Starting the backend rigged up all of the necessary event
	 * loop configuration to listen to libinput events, DRM events, generate
	 * frame events at the refresh rate, and so on. */
	s.display.Run()

	/* Once s.display.Run() returns, we destroy all clients then shut down the
	 * server. */
	s.display.DestroyClients()
	s.scene.Tree().Node().Destroy()
	s.cursorMgr.Destroy()
	s.outputLayout.Destroy()
	s.display.Destroy()
	return nil
}
