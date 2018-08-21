package main

import (
	"github.com/alexbakker/go-wlroots/wlroots"
)

type View struct {
	surface wlroots.XDGSurface
	mapped  bool
}

func NewView(surface wlroots.XDGSurface) *View {
	view := &View{surface: surface}
	surface.OnMap(view.onMap)
	surface.OnUnmap(view.onUnmap)
	return view
}

func (v *View) onMap(surface wlroots.XDGSurface) {
	v.mapped = true
}

func (v *View) onUnmap(surface wlroots.XDGSurface) {
	v.mapped = false
}

func (v *View) Mapped() bool {
	return v.mapped
}

func (v *View) Surface() wlroots.Surface {
	return v.surface.Surface()
}

func (v *View) XDGSurface() wlroots.XDGSurface {
	return v.surface
}
