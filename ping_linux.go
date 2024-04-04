package main

import (
	"syscall"

	"golang.org/x/sys/unix"
)

func Control(network, address string, c syscall.RawConn) (err error) {
	// Vln(6, "[Control]", network, address, c)
	return c.Control(func(fd uintptr) {
		err = unix.SetsockoptInt(int(fd), unix.IPPROTO_IP, unix.IP_BIND_ADDRESS_NO_PORT, 1)
		// err = unix.SetsockoptInt(int(fd), unix.SOL_SOCKET, unix.SO_REUSEADDR, 1)
		// if err != nil {
		// 	return
		// }
		// err = unix.SetsockoptInt(int(fd), unix.SOL_SOCKET, unix.SO_REUSEPORT, 1)
	})
}
