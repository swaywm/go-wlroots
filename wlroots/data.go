package wlroots

// #include <wlr/types/wlr_data_device.h>
import "C"

type DataDeviceManager struct {
	p *C.struct_wlr_data_device_manager
}

func NewDataDeviceManager(display Display) DataDeviceManager {
	p := C.wlr_data_device_manager_create(display.p)
	return DataDeviceManager{p: p}
}

func (m DataDeviceManager) Destroy() {
	C.wlr_data_device_manager_destroy(m.p)
}
