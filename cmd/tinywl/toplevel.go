package main

import "github.com/swaywm/go-wlroots/wlroots"

type TopLevel struct {
	surface     wlroots.XDGSurface
	Mapped      bool
	X           float64
	Y           float64
	xdgTopLevel wlroots.XDGTopLevel
	SceneTree   wlroots.SceneTree
}

func NewTopLevel(surface wlroots.XDGSurface) *TopLevel {
	return &TopLevel{surface: surface}
}

func (v *TopLevel) Surface() wlroots.Surface {
	return v.surface.Surface()
}

func (v *TopLevel) XDGSurface() wlroots.XDGSurface {
	return v.surface
}
