package main

import (
	"fmt"
	"os"
	"time"

	"github.com/alexbakker/go-wlroots/wlroots"
	"github.com/alexbakker/go-wlroots/xkbcommon"
)

type Server struct {
	display    wlroots.Display
	backend    wlroots.Backend
	renderer   wlroots.Renderer
	layout     wlroots.OutputLayout
	cursor     wlroots.Cursor
	cursorMgr  wlroots.XCursorManager
	compositor wlroots.Compositor
	dataDevMgr wlroots.DataDeviceManager
	seat       wlroots.Seat
	xdgShell   wlroots.XDGShell
	dmaBuf     wlroots.DMABuf

	views       []*View
	hasKeyboard bool
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
	s.renderer.InitDisplay(s.display)

	// create compositor, dmabuf and data device manager interfaces
	s.compositor = wlroots.NewCompositor(s.display, s.renderer)
	s.dataDevMgr = wlroots.NewDataDeviceManager(s.display)
	s.dmaBuf = wlroots.NewDMABuf(s.display, s.renderer)

	// create output layout
	s.layout = wlroots.NewOutputLayout()

	// create xdg-shell
	s.xdgShell = wlroots.NewXDGShell(s.display)
	s.xdgShell.OnNewSurface(s.handleNewXDGSurface)

	// create cursor and load xcursor themes
	s.cursor = wlroots.NewCursor()
	s.cursor.OnMotion(s.handleCursorMotion)
	s.cursor.OnMotionAbsolute(s.handleCursorMotionAbsolute)
	s.cursor.AttachOutputLayout(s.layout)
	s.cursorMgr = wlroots.NewXCursorManager()
	s.cursorMgr.Load()

	// configure seat
	s.seat = wlroots.NewSeat(s.display, "seat0")
	s.seat.OnSetCursor(s.handleSetCursor)

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
	return nil
}

func (s *Server) Destroy() {
	panic("not implemented")
}

func (s *Server) handleNewFrame(output wlroots.Output) {
	output.MakeCurrent()

	width, height := output.EffectiveResolution()
	s.renderer.Begin(output, width, height)
	s.renderer.Clear(wlroots.Color{0.3, 0.3, 0.3, 1.0})

	// render all of the views
	for _, view := range s.views {
		if !view.Mapped() {
			continue
		}

		s.renderView(output, view)
	}

	s.renderer.End()
	output.SwapBuffers()
}

func (s *Server) handleDestroyOutput(output wlroots.Output) {
}

func (s *Server) handleNewOutput(output wlroots.Output) {
	modes := output.Modes()
	if len(modes) > 0 {
		panic("unsupported backend")
	}

	output.OnFrame(s.handleNewFrame)
	output.OnDestroy(s.handleDestroyOutput)

	s.layout.AddOutputAuto(output)
	output.CreateGlobal()
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
	s.cursorMgr.SetImage(s.cursor, "left_ptr")
}

func (s *Server) handleSetCursor() {
	fmt.Println("received set cursor request")
}

func (s *Server) handleNewInput(dev wlroots.InputDevice) {
	switch dev.Type() {
	case wlroots.InputDevicePointer:
		s.cursor.AttachInputDevice(dev)
	case wlroots.InputDeviceKeyboard:
		xkb := xkbcommon.NewContext()
		keymap := xkb.Map()
		keyboard := dev.Keyboard()
		keyboard.SetKeymap(keymap)
		keymap.Destroy()
		xkb.Destroy()
		keyboard.SetRepeatInfo(25, 600)

		s.seat.SetKeyboard(dev)
		s.hasKeyboard = true
	}

	caps := uint32(wlroots.SeatCapabilityPointer)
	if s.hasKeyboard {
		caps |= wlroots.SeatCapabilityKeyboard
	}
	s.seat.SetCapabilities(caps)
}

func (s *Server) handleNewXDGSurface(surface wlroots.XDGSurface) {
	if surface.Role() != wlroots.XDGSurfaceRoleTopLevel {
		return
	}

	view := NewView(surface)
	s.views = append(s.views, view)
}

func (s *Server) renderView(output wlroots.Output, view *View) {
	view.XDGSurface().Walk(func(surface wlroots.Surface, sx int, sy int) {
		texture := surface.Texture()
		ox, oy := s.layout.Coords(output)
		ox += 0 + float64(sx)
		oy += 0 + float64(sy)

		scale := output.Scale()
		state := surface.CurrentState()
		transform := wlroots.OutputTransformInvert(state.Transform())

		var box wlroots.Box
		box.Set(
			int(ox*float64(scale)),
			int(oy*float64(scale)),
			int(float32(state.Width())*scale),
			int(float32(state.Height())*scale),
		)

		var matrix wlroots.Matrix
		transformMatrix := output.TransformMatrix()
		matrix.ProjectBox(&box, transform, 0, &transformMatrix)

		s.renderer.RenderTextureWithMatrix(texture, &matrix, 1)

		surface.SendFrameDone(time.Now())
	})
}
