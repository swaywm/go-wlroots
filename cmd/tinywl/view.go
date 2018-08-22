package main

import (
	"github.com/alexbakker/go-wlroots/wlroots"
)

type View struct {
	surface wlroots.XDGSurface
	mapped  bool
	X       float64
	Y       float64
}

func NewView(surface wlroots.XDGSurface) *View {
	view := &View{surface: surface}
	return view
}

func (v *View) Mapped() bool {
	return v.mapped
}

func (v *View) SetMapped(mapped bool) {
	v.mapped = mapped
}

func (v *View) Surface() wlroots.Surface {
	return v.surface.Surface()
}

func (v *View) XDGSurface() wlroots.XDGSurface {
	return v.surface
}
