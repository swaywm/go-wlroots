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
	display    wlroots.Display
	backend    wlroots.Backend
	allocator  wlroots.Allocator
	renderer   wlroots.Renderer
	layout     wlroots.OutputLayout
	cursor     wlroots.Cursor
	cursorMgr  wlroots.XCursorManager
	compositor wlroots.Compositor
	dataDevMgr wlroots.DataDeviceManager
	seat       wlroots.Seat
	xdgShell   wlroots.XDGShell

	views     []*View
	keyboards []*Keyboard

	grabbedView *View
	cursorMode  CursorMode
	grabX       float64
	grabY       float64
	grabWidth   int
	grabHeight  int
	resizeEdges wlroots.Edges
}

type Keyboard struct {
	dev wlroots.InputDevice
}

func NewServer() (*Server, error) {
	s := new(Server)

	// create display
	s.display = wlroots.NewDisplay()

	// create backend
	s.backend = wlroots.NewBackend(s.display)
	s.backend.OnNewOutput(s.handleNewOutput)

	s.backend.OnNewInput(s.handleNewInput)
	s.renderer = s.backend.Renderer()
	s.allocator = s.backend.Allocator(s.renderer)
	s.renderer.InitDisplay(s.display)

	// create compositor and data device manager interfaces
	s.compositor = wlroots.NewCompositor(s.display, s.renderer)
	s.dataDevMgr = wlroots.NewDataDeviceManager(s.display)

	// create output layout
	s.layout = wlroots.NewOutputLayout()

	// create xdg-shell
	s.xdgShell = wlroots.NewXDGShell(s.display)
	s.xdgShell.OnNewSurface(s.handleNewXDGSurface)

	// create cursor and load xcursor themes
	s.cursor = wlroots.NewCursor()
	s.cursor.OnMotion(s.handleCursorMotion)
	s.cursor.OnMotionAbsolute(s.handleCursorMotionAbsolute)
	s.cursor.OnButton(s.handleCursorButton)
	s.cursor.OnAxis(s.handleCursorAxis)
	s.cursor.OnFrame(s.handleCursorFrame)
	s.cursor.AttachOutputLayout(s.layout)
	s.cursorMgr = wlroots.NewXCursorManager()
	s.cursorMgr.Load()

	// configure seat
	s.seat = wlroots.NewSeat(s.display, "seat0")
	s.seat.OnSetCursorRequest(s.handleSetCursorRequest)

	return s, nil
}

func (s *Server) Start() error {
	// start the backend
	if err := s.backend.Start(); err != nil {
		return err
	}

	// setup socket for wayland clients to connect to
	socket, err := s.display.AddSocketAuto()
	if err != nil {
		return err
	}
	if err = os.Setenv("WAYLAND_DISPLAY", socket); err != nil {
		return err
	}

	return nil
}

func (s *Server) Run() error {
	s.display.Run()

	s.display.Destroy()
	s.layout.Destroy()
	s.cursorMgr.Destroy()
	s.cursor.Destroy()
	return nil
}

func (s *Server) viewAt(lx float64, ly float64) (*View, wlroots.Surface, float64, float64) {
	for i := len(s.views) - 1; i >= 0; i-- {
		view := s.views[i]
		surface, sx, sy := view.XDGSurface().SurfaceAt(lx-view.X, ly-view.Y)
		if !surface.Nil() {
			return view, surface, sx, sy
		}
	}

	return nil, wlroots.Surface{}, 0, 0
}

func (s *Server) renderView(output wlroots.Output, view *View) {
	view.XDGSurface().Walk(func(surface wlroots.Surface, sx int, sy int) {
		texture := surface.Texture()
		if texture.Nil() {
			return
		}

		ox, oy := s.layout.Coords(output)
		ox += view.X + float64(sx)
		oy += view.Y + float64(sy)

		scale := output.Scale()
		state := surface.CurrentState()
		transform := wlroots.OutputTransformInvert(state.Transform())

		box := wlroots.Box{
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

func (s *Server) focusView(view *View, surface wlroots.Surface) {
	prevSurface := s.seat.KeyboardState().FocusedSurface()
	if prevSurface == surface {
		// don't re-focus an already focused surface
		return
	}

	if !prevSurface.Nil() {
		// deactivate the previously focused surface
		prev := prevSurface.XDGSurface()
		prev.TopLevelSetActivated(false)
	}

	// move the view to the front
	for i := len(s.views) - 1; i >= 0; i-- {
		if s.views[i] == view {
			s.views = append(s.views[:i], s.views[i+1:]...)
			s.views = append(s.views, view)
			break
		}
	}

	view.XDGSurface().TopLevelSetActivated(true)
	s.seat.NotifyKeyboardEnter(view.Surface(), s.seat.Keyboard())
}

func (s *Server) handleNewFrame(output wlroots.Output) {
	output.AttachRender()

	width, height := output.EffectiveResolution()
	s.renderer.Begin(output, width, height)
	s.renderer.Clear(&wlroots.Color{0.3, 0.3, 0.3, 1.0})

	// render all of the views
	for _, view := range s.views {
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
	// TODO: pick the preferred mode instead of the first one
	modes := output.Modes()
	if len(modes) > 0 {
		output.SetMode(modes[len(modes)-1])
	}

	output.InitRender(s.allocator, s.renderer)

	s.layout.AddOutputAuto(output)
	output.OnFrame(s.handleNewFrame)
	output.CreateGlobal()
	output.SetTitle(fmt.Sprintf("tinywl (wlroots) - %s", output.Name()))
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
	// check whether we're currently moving/resizing a view
	if s.cursorMode == CursorModeMove {
		s.processCursorMove(time)
		return
	} else if s.cursorMode == CursorModeResize {
		s.processCursorResize(time)
		return
	}

	// if not, find the view below the cursor and send the event to that
	view, surface, sx, sy := s.viewAt(s.cursor.X(), s.cursor.Y())
	if view == nil {
		// if there is no view, set the default cursor image
		s.cursorMgr.SetCursorImage(s.cursor, "left_ptr")
	}

	if !surface.Nil() {
		s.seat.NotifyPointerEnter(surface, sx, sy)
		if s.seat.PointerState().FocusedSurface() == surface {
			// we only need to notify on motion if the focus didn't change
			s.seat.NotifyPointerMotion(time, sx, sy)
		}
	} else {
		s.seat.ClearPointerFocus()
	}
}

func (s *Server) processCursorMove(time uint32) {
	s.grabbedView.X = s.cursor.X() - s.grabX
	s.grabbedView.Y = s.cursor.Y() - s.grabY
}

func (s *Server) processCursorResize(time uint32) {
	dx := s.cursor.X() - s.grabX
	dy := s.cursor.Y() - s.grabY
	x := s.grabbedView.X
	y := s.grabbedView.Y
	width := s.grabWidth
	height := s.grabHeight

	if s.resizeEdges&wlroots.EdgeTop != 0 {
		y = s.grabY + dy
		height -= int(dy)
		if height < 1 {
			y += float64(height)
		}
	} else if s.resizeEdges&wlroots.EdgeBottom != 0 {
		height += int(dy)
	}

	if s.resizeEdges&wlroots.EdgeLeft != 0 {
		x = s.grabX + dx
		width -= int(dx)
		if width < 1 {
			x += float64(width)
		}
	} else if s.resizeEdges&wlroots.EdgeRight != 0 {
		width += int(dx)
	}

	s.grabbedView.X = x
	s.grabbedView.Y = y
	s.grabbedView.XDGSurface().TopLevelSetSize(uint32(width), uint32(height))
}

func (s *Server) handleSetCursorRequest(client wlroots.SeatClient, surface wlroots.Surface, serial uint32, hotspotX int32, hotspotY int32) {
	focusedClient := s.seat.PointerState().FocusedClient()
	if focusedClient == client {
		s.cursor.SetSurface(surface, hotspotX, hotspotY)
	}
}

func (s *Server) handleNewInput(dev wlroots.InputDevice) {
	switch dev.Type() {
	case wlroots.InputDeviceTypePointer:
		s.cursor.AttachInputDevice(dev)
	case wlroots.InputDeviceTypeKeyboard:
		context := xkb.NewContext()
		keymap := context.Map()
		keyboard := dev.Keyboard()
		keyboard.SetKeymap(keymap)
		keymap.Destroy()
		context.Destroy()
		keyboard.SetRepeatInfo(25, 600)

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
		s.keyboards = append(s.keyboards, &Keyboard{dev: dev})
	}

	caps := wlroots.SeatCapabilityPointer
	if len(s.keyboards) > 0 {
		caps |= wlroots.SeatCapabilityKeyboard
	}
	s.seat.SetCapabilities(caps)
}

func (s *Server) handleNewXDGSurface(surface wlroots.XDGSurface) {
	if surface.Role() != wlroots.XDGSurfaceRoleTopLevel {
		return
	}

	view := NewView(surface)
	surface.OnMap(func(surface wlroots.XDGSurface) {
		view.Mapped = true
		s.focusView(view, surface.Surface())
	})
	surface.OnUnmap(func(surface wlroots.XDGSurface) {
		view.Mapped = false
	})
	surface.OnDestroy(func(surface wlroots.XDGSurface) {
		// TODO: keep track of views some other way
		for i := range s.views {
			if s.views[i] == view {
				s.views = append(s.views[:i], s.views[i+1:]...)
				break
			}
		}
	})

	toplevel := surface.TopLevel()
	toplevel.OnRequestMove(func(client wlroots.SeatClient, serial uint32) {
		s.beginInteractive(view, CursorModeMove, 0)
	})
	toplevel.OnRequestResize(func(client wlroots.SeatClient, serial uint32, edges wlroots.Edges) {
		s.beginInteractive(view, CursorModeResize, edges)
	})

	s.views = append(s.views, view)
}

func (s *Server) handleCursorButton(dev wlroots.InputDevice, time uint32, button uint32, state wlroots.ButtonState) {
	s.seat.NotifyPointerButton(time, button, state)

	if state == wlroots.ButtonStateReleased {
		s.cursorMode = CursorModePassThrough
	} else {
		view, surface, _, _ := s.viewAt(s.cursor.X(), s.cursor.Y())
		if view != nil {
			s.focusView(view, surface)
		}
	}
}

func (s *Server) handleCursorAxis(dev wlroots.InputDevice, time uint32, source wlroots.AxisSource, orientation wlroots.AxisOrientation, delta float64, deltaDiscrete int32) {
	s.seat.NotifyPointerAxis(time, orientation, delta, deltaDiscrete, source)
}

func (s *Server) handleCursorFrame() {
	s.seat.NotifyPointerFrame()
}

func (s *Server) handleKeyBinding(sym xkb.KeySym) bool {
	switch sym {
	case xkb.KeySymEscape:
		s.display.Terminate()
	case xkb.KeySymF1:
		if len(s.views) < 2 {
			break
		}

		i := len(s.views) - 1
		focusedView := s.views[i]
		nextView := s.views[i-1]

		// move the focused view to the back of the view list
		s.views = append(s.views[:i], s.views[i+1:]...)
		s.views = append([]*View{focusedView}, s.views...)

		// focus the next view
		s.focusView(nextView, nextView.Surface())
	default:
		return false
	}

	return true
}

func (s *Server) beginInteractive(view *View, mode CursorMode, edges wlroots.Edges) {
	// deny requests from unfocused clients
	if view.Surface() != s.seat.PointerState().FocusedSurface() {
		return
	}

	box := view.XDGSurface().Geometry()
	if mode == CursorModeMove {
		s.grabX = s.cursor.X() - view.X
		s.grabY = s.cursor.Y() - view.Y
	} else {
		s.grabX = s.cursor.X() + float64(box.X)
		s.grabY = s.cursor.Y() + float64(box.Y)
	}

	s.grabbedView = view
	s.cursorMode = mode
	s.grabWidth = box.Width
	s.grabHeight = box.Height
	s.resizeEdges = edges
}
