package wlroots

/*
 * This an unstable interface of wlroots. No guarantees are made regarding the
 * future consistency of this API.
 */

// #cgo pkg-config: wlroots-0.18 wayland-server
// #cgo CFLAGS: -D_GNU_SOURCE -DWLR_USE_UNSTABLE
// #include <wlr/types/wlr_buffer.h>
import "C"

/**
 * A buffer containing pixel data.
 *
 * A buffer has a single producer (the party who created the buffer) and
 * multiple consumers (parties reading the buffer). When all consumers are done
 * with the buffer, it gets released and can be re-used by the producer. When
 * the producer and all consumers are done with the buffer, it gets destroyed.
 */
type Buffer struct {
	p *C.struct_wlr_buffer
}

/**
 * Unreference the buffer. This function should be called by producers when
 * they are done with the buffer.
 */
func (b Buffer) Drop() {
	C.wlr_buffer_drop(b.p)
}

/**
 * Lock the buffer. This function should be called by consumers to make
 * sure the buffer can be safely read from. Once the consumer is done with the
 * buffer, they should call wlr_buffer_unlock().
 */
func (b Buffer) Lock() Buffer {
	p := C.wlr_buffer_lock(b.p)
	return Buffer{p: p}
}

/**
 * Unlock the buffer. This function should be called by consumers once they are
 * done with the buffer.
 */
func (b Buffer) Unlock() {
	C.wlr_buffer_unlock(b.p)
}
