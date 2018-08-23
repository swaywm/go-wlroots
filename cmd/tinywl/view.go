package main

import "github.com/swaywm/go-wlroots/wlroots"

type View struct {
	surface wlroots.XDGSurface
	Mapped  bool
	X       float64
	Y       float64
}

func NewView(surface wlroots.XDGSurface) *View {
	return &View{surface: surface}
}

func (v *View) Surface() wlroots.Surface {
	return v.surface.Surface()
}

func (v *View) XDGSurface() wlroots.XDGSurface {
	return v.surface
}
