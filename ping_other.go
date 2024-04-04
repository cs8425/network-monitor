//go:build !linux

package main

import (
	"syscall"
)

func Control(network, address string, c syscall.RawConn) (err error) {
	// Vln(6, "[Control]", network, address, c)
	// TODO: something like IP_BIND_ADDRESS_NO_PORT for stop running out of ephemeral ports
	return nil
}
