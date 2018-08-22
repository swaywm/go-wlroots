package wlroots

// #include <wayland-server.h>
import "C"
import "errors"

type Display struct {
	p *C.struct_wl_display
}

func NewDisplay() Display {
	p := C.wl_display_create()
	return Display{p: p}
}

func (d Display) Destroy() {
	C.wl_display_destroy(d.p)
}

func (d Display) Run() {
	C.wl_display_run(d.p)
}

func (d Display) Terminate() {
	C.wl_display_terminate(d.p)
}

func (d Display) AddSocketAuto() (string, error) {
	socket := C.wl_display_add_socket_auto(d.p)
	if socket == nil {
		return "", errors.New("error adding socket")
	}

	return C.GoString(socket), nil
}
