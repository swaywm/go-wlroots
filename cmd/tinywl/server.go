package main

import (
	"fmt"
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

	xdgShell  wlroots.XDGShell
	topLevels []*TopLevel

	cursor    wlroots.Cursor
	cursorMgr wlroots.XCursorManager

	seat            wlroots.Seat
	keyboards       []*Keyboard
	cursorMode      CursorMode
	grabbedTopLevel *TopLevel
	grabX, grabY    float64
	grabGeobox      wlroots.GeoBox
	resizeEdges     wlroots.Edges

	outputLayout wlroots.OutputLayout
	outputs      []wlroots.Output
}

type Keyboard struct {
	dev wlroots.InputDevice
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
	s.backend, err = wlroots.NewBackend(s.display)
	if err != nil {
		return nil, err
	}

	/* Autocreates a renderer, either Pixman, GLES2 or Vulkan for us. The user
	 * can also specify a renderer using the WLR_RENDERER env var.
	 * The renderer is responsible for defining the various pixel formats it
	 * supports for shared memory, this configures that for clients. */
	s.renderer, err = wlroots.NewRenderer(s.backend)
	if err != nil {
		return nil, err
	}
	s.renderer.InitDisplay(s.display)

	/* Autocreates an allocator for us.
	 * The allocator is the bridge between the renderer and the backend. It
	 * handles the buffer creation, allowing wlroots to render onto the
	 * screen */
	s.allocator, err = wlroots.NewAllocator(s.backend, s.renderer)
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
	wlroots.NewCompositor(s.display, 5, s.renderer)
	wlroots.NewSubCompositor(s.display)
	wlroots.NewDataDeviceManager(s.display)

	/* Creates an output layout, which a wlroots utility for working with an
	 * arrangement of screens in a physical layout. */
	s.outputLayout = wlroots.NewOutputLayout()

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
	s.xdgShell = wlroots.NewXDGShell(s.display, 3)
	s.xdgShell.OnNewSurface(s.handleNewXDGSurface)

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
	s.seat = wlroots.NewSeat(s.display, "seat0")
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

func (s *Server) topLevelAt(lx float64, ly float64) (*TopLevel, wlroots.Surface, float64, float64) {
	for i := len(s.topLevels) - 1; i >= 0; i-- {
		topLevel := s.topLevels[i]
		surface, sx, sy := topLevel.XDGSurface().SurfaceAt(lx-topLevel.X, ly-topLevel.Y)
		if !surface.Nil() {
			return topLevel, surface, sx, sy
		}
	}

	return nil, wlroots.Surface{}, 0, 0
}

func (s *Server) renderView(output wlroots.Output, topLevel *TopLevel) {
	topLevel.XDGSurface().Walk(func(surface wlroots.Surface, sx int, sy int) {
		texture := surface.Texture()
		if texture.Nil() {
			return
		}

		ox, oy := s.outputLayout.Coords(output)
		ox += topLevel.X + float64(sx)
		oy += topLevel.Y + float64(sy)

		scale := output.Scale()
		state := surface.CurrentState()
		transform := wlroots.OutputTransformInvert(state.Transform())

		box := wlroots.GeoBox{
			X:      int(ox * float64(scale)),
			Y:      int(oy * float64(scale)),
			Width:  int(float32(state.Width()) * scale),
			Height: int(float32(state.Height()) * scale),
		}

		var matrix wlroots.Matrix
		transformMatrix := output.TransformMatrix()
		matrix.ProjectBox(&box, transform, 0, &transformMatrix)

		s.renderer.RenderTextureWithMatrix(texture, &matrix, 1)

		surface.SendFrameDone(time.Now())
	})
}

func (s *Server) focusTopLevel(topLevel *TopLevel, surface wlroots.Surface) {
	/* Note: this function only deals with keyboard focus. */
	if topLevel == nil {
		return
	}
	prevSurface := s.seat.KeyboardState().FocusedSurface()
	if prevSurface == surface {
		/* Don't re-focus an already focused surface. */
		return
	}

	if !prevSurface.Nil() {
		/*
		 * Deactivate the previously focused surface. This lets the client know
		 * it no longer has focus and the client will repaint accordingly, e.g.
		 * stop displaying a caret.
		 */
		prev := prevSurface.XDGSurface()
		prev.TopLevelSetActivated(false)
	}

	/* Move the toplevel to the front */
	topLevel.SceneTree.Node().RaiseToTop()

	for i := len(s.topLevels) - 1; i >= 0; i-- {
		if s.topLevels[i] == topLevel {
			s.topLevels = append(s.topLevels[:i], s.topLevels[i+1:]...)
			s.topLevels = append(s.topLevels, topLevel)
			break
		}
	}
	/* Activate the new surface */
	topLevel.XDGSurface().TopLevelSetActivated(true)
	/*
	 * Tell the seat to have the keyboard enter this surface. wlroots will keep
	 * track of this and automatically send key events to the appropriate
	 * clients without additional work on your part.
	 */
	s.seat.NotifyKeyboardEnter(topLevel.Surface(), s.seat.Keyboard())
}

func (s *Server) handleNewFrame(output wlroots.Output) {
	output.AttachRender()

	width, height := output.EffectiveResolution()
	s.renderer.Begin(output, width, height)
	s.renderer.Clear(&wlroots.Color{R: 0.3, G: 0.3, B: 0.3, A: 1.0})

	// render all of the topLevels
	for _, view := range s.topLevels {
		if !view.Mapped {
			continue
		}

		s.renderView(output, view)
	}

	output.RenderSoftwareCursors()
	s.renderer.End()
	output.Commit()
}

func (s *Server) handleNewOutput(output wlroots.Output) {
	output.InitRender(s.allocator, s.renderer)

	/* The output may be disabled, switch it on. */
	var oState wlroots.OutputState
	oState.StateInit()
	oState.StateSetEnabled()

	/* Some backends don't have modes. DRM+KMS does, and we need to set a mode
	 * before we can use the output. The mode is a tuple of (width, height,
	 * refresh rate), and each monitor supports only a specific set of modes. We
	 * just pick the monitor's preferred mode, a more sophisticated compositor
	 * would let the user configure it. */
	mode := output.PrefferedMode()
	output.SetMode(mode)
	output.Enable(true)
	if !output.Commit() {
		return
	}

	output.OnFrame(s.handleNewFrame)
	s.outputLayout.AddOutputAuto(output)
	output.SetTitle(fmt.Sprintf("tinywl (go-wlroots) - %s", output.Name()))
}

func (s *Server) handleCursorMotion(dev wlroots.InputDevice, time uint32, dx float64, dy float64) {
	s.cursor.Move(dev, dx, dy)
	s.processCursorMotion(time)
}

func (s *Server) handleCursorMotionAbsolute(dev wlroots.InputDevice, time uint32, x float64, y float64) {
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

	// if not, find the view below the cursor and send the event to that
	view, surface, sx, sy := s.topLevelAt(s.cursor.X(), s.cursor.Y())
	if view == nil {
		// if there is no view, set the default cursor image
		// s.cursorMgr.SetCursorImage(s.cursor, "left_ptr")
	}

	if !surface.Nil() {
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
		s.seat.NotifyPointerEnter(surface, sx, sy)
		s.seat.NotifyPointerMotion(time, sx, sy)
	} else {
		/* Clear pointer focus so future button events and such are not sent to
		 * the last client to have the cursor over it. */
		s.seat.ClearPointerFocus()
	}
}

func (s *Server) processCursorMove(time uint32) {
	s.grabbedTopLevel.X = s.cursor.X() - s.grabX
	s.grabbedTopLevel.Y = s.cursor.Y() - s.grabY
	s.grabbedTopLevel.SceneTree.Node().SetPosition(s.grabbedTopLevel.X, s.grabbedTopLevel.Y)
}

func (s *Server) processCursorResize(time uint32) {
	borderX := s.cursor.X() - s.grabX
	borderY := s.cursor.Y() - s.grabY
	// x := s.grabbedTopLevel.X
	// y := s.grabbedTopLevel.Y
	nLeft := s.grabGeobox.X
	nRight := s.grabGeobox.X - s.grabGeobox.Width
	nTop := s.grabGeobox.Y
	nBottom := s.grabGeobox.Y - s.grabGeobox.Height

	if s.resizeEdges&wlroots.EdgeTop != 0 {
		nTop = int(borderY)
		if nTop >= nBottom {
			nTop = nBottom + 1
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
			nLeft = nRight + 1
		}
	} else if s.resizeEdges&wlroots.EdgeRight != 0 {
		nRight = int(borderX)
		if nRight <= nLeft {
			nRight = nLeft + 1
		}
	}

	nWidth := nRight - nLeft
	nHeight := nBottom - nTop
	s.grabbedTopLevel.XDGSurface().TopLevelSetSize(uint32(nWidth), uint32(nHeight))
}

func (s *Server) handleSetCursorRequest(client wlroots.SeatClient, surface wlroots.Surface, serial uint32, hotspotX int32, hotspotY int32) {
	focusedClient := s.seat.PointerState().FocusedClient()
	if focusedClient == client {
		s.cursor.SetSurface(surface, hotspotX, hotspotY)
	}
}

func (s *Server) handleNewPointer(dev wlroots.InputDevice) {
	s.cursor.AttachInputDevice(dev)
}

func (s *Server) handleNewKeyboard(dev wlroots.InputDevice) {
	keyboard := dev.Keyboard()

	/* We need to prepare an XKB keymap and assign it to the keyboard. This
	 * assumes the defaults (e.g. layout = "us"). */
	context := xkb.NewContext()
	keymap := context.KeyMap()
	keyboard.SetKeymap(keymap)
	keymap.Destroy()
	context.Destroy()
	keyboard.SetRepeatInfo(25, 600)

	/* Here we set up listeners for keyboard events. */
	keyboard.OnKey(func(keyboard wlroots.Keyboard, time uint32, keyCode uint32, updateState bool, state wlroots.KeyState) {
		// translate libinput keycode to xkbcommon and obtain keysyms
		syms := keyboard.XKBState().Syms(xkb.KeyCode(keyCode + 8))

		var handled bool
		modifiers := keyboard.Modifiers()
		if (modifiers&wlroots.KeyboardModifierAlt != 0) && state == wlroots.KeyStatePressed {
			for _, sym := range syms {
				handled = s.handleKeyBinding(sym)
			}
		}

		if !handled {
			s.seat.SetKeyboard(dev)
			s.seat.NotifyKeyboardKey(time, keyCode, state)
		}
	})

	keyboard.OnModifiers(func(keyboard wlroots.Keyboard) {
		s.seat.SetKeyboard(dev)
		s.seat.NotifyKeyboardModifiers(keyboard)
	})

	s.seat.SetKeyboard(dev)

	/* And add the keyboard to our list of keyboards */
	s.keyboards = append(s.keyboards, &Keyboard{dev: dev})
}

func (s *Server) handleNewInput(dev wlroots.InputDevice) {
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

func (s *Server) handleNewXDGSurface(xdgSurface wlroots.XDGSurface) {
	/* This event is raised when wlr_xdg_shell receives a new xdg xdgSurface from a
	 * client, either a toplevel (application window) or popup. */

	if xdgSurface.Role() == wlroots.XDGSurfaceRolePopup {
		// pTree := xdgSurface.Popup().Parent().SceneTree()
		// xdgSurface.SetData(pTree)
		return
	}

	topLevel := NewTopLevel(xdgSurface)
	topLevel.xdgTopLevel = xdgSurface.TopLevel()
	topLevel.SceneTree = s.scene.Tree().XDGSurfaceCreate(topLevel.xdgTopLevel.Base())
	xdgSurface.SetData(topLevel.SceneTree)
	xdgSurface.OnMap(func(surface wlroots.XDGSurface) {
		topLevel.Mapped = true
		s.focusTopLevel(topLevel, surface.Surface())
	})
	xdgSurface.OnUnmap(func(surface wlroots.XDGSurface) {
		topLevel.Mapped = false
	})
	xdgSurface.OnDestroy(func(surface wlroots.XDGSurface) {
		// TODO: keep track of topLevels some other way
		for i := range s.topLevels {
			if s.topLevels[i] == topLevel {
				s.topLevels = append(s.topLevels[:i], s.topLevels[i+1:]...)
				break
			}
		}
	})

	toplevel := xdgSurface.TopLevel()
	toplevel.OnRequestMove(func(client wlroots.SeatClient, serial uint32) {
		s.beginInteractive(topLevel, CursorModeMove, 0)
	})
	toplevel.OnRequestResize(func(client wlroots.SeatClient, serial uint32, edges wlroots.Edges) {
		s.beginInteractive(topLevel, CursorModeResize, edges)
	})

	s.topLevels = append(s.topLevels, topLevel)
}

func (s *Server) resetCursorMode() {
	/* Reset the cursor mode to passthrough. */
	s.cursorMode = CursorModePassThrough
	s.grabbedTopLevel = nil
}

func (s *Server) handleCursorButton(dev wlroots.InputDevice, time uint32, button uint32, state wlroots.ButtonState) {
	/* This event is forwarded by the cursor when a pointer emits a button
	 * event. */

	/* Notify the client with pointer focus that a button press has occurred */
	s.seat.NotifyPointerButton(time, button, state)

	if state == wlroots.ButtonStateReleased {
		/* If you released any buttons, we exit interactive move/resize mode. */
		s.resetCursorMode()
	} else {
		topLevel, surface, _, _ := s.topLevelAt(s.cursor.X(), s.cursor.Y())
		if topLevel != nil {
			/* Focus that client if the button was _pressed_ */
			s.focusTopLevel(topLevel, surface)
		}
	}
}

func (s *Server) handleCursorAxis(dev wlroots.InputDevice, time uint32, source wlroots.AxisSource, orientation wlroots.AxisOrientation, delta float64, deltaDiscrete int32) {
	/* This event is forwarded by the cursor when a pointer emits an axis event,
	 * for example when you move the scroll wheel. */

	/* Notify the client with pointer focus of the axis event. */
	s.seat.NotifyPointerAxis(time, orientation, delta, deltaDiscrete, source)
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
	switch sym {
	case xkb.KeySymEscape:
		s.display.Terminate()
	case xkb.KeySymF1:
		if len(s.topLevels) < 2 {
			break
		}

		i := len(s.topLevels) - 1
		focusedView := s.topLevels[i]
		nextView := s.topLevels[i-1]

		// move the focused view to the back of the view list
		s.topLevels = append(s.topLevels[:i], s.topLevels[i+1:]...)
		s.topLevels = append([]*TopLevel{focusedView}, s.topLevels...)

		// focus the next view
		s.focusTopLevel(nextView, nextView.Surface())
	default:
		return false
	}

	return true
}

func (s *Server) beginInteractive(topLevel *TopLevel, mode CursorMode, edges wlroots.Edges) {
	/* This function sets up an interactive move or resize operation, where the
	 * compositor stops propegating pointer events to clients and instead
	 * consumes them itself, to move or resize windows. */
	if topLevel.Surface() != s.seat.PointerState().FocusedSurface() {
		/* Deny move/resize requests from unfocused clients. */
		return
	}

	s.grabbedTopLevel = topLevel
	s.cursorMode = mode

	box := topLevel.XDGSurface().Geometry()
	if mode == CursorModeMove {
		s.grabX = s.cursor.X() - topLevel.X
		s.grabY = s.cursor.Y() - topLevel.Y
	} else {
		s.grabX = s.cursor.X() + float64(box.X)
		s.grabY = s.cursor.Y() + float64(box.Y)
	}

	s.resizeEdges = edges
	s.grabGeobox = box
}
