package main

import (
	"errors"
	"net"
	"sync"
	"syscall"
	"time"
)

const (
	SYN_INIT_TIMEOUT = 1000 - 5
)

type PingTarget struct {
	Name string // for Label
	Addr string
}

func pingWorker(to time.Duration, wg *sync.WaitGroup, queueCh chan *PingTarget) {
	defer wg.Done()

	d := &net.Dialer{
		Timeout: to,
		Control: Control,
	}
	for info := range queueCh {
		dt, err := ping(d, info.Addr)
		Vln(4, "[dt]", info.Name, info.Addr, dt, err)
		appendResult(info.Name, dt, err)
	}
}

func ping(d *net.Dialer, addr string) (dt time.Duration, err error) {
	// resolve first, do not count DNS lookup time
	ipAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return -1, err // skip if DNS not found
	}

	t0 := time.Now()
	// d.Control = Control
	conn, err := d.Dial("tcp", ipAddr.String())
	dt = time.Since(t0)
	if err != nil {
		if errors.Is(err, syscall.ECONNREFUSED) {
			Vln(5, "[RST]:", dt, err)
			return dt, nil
		}
		// if os.IsTimeout(err) {
		// 	log.Println("[os.IsTimeout]:", err)
		// }
		// if err, ok := err.(net.Error); ok && err.Timeout() {
		// 	log.Println("[net.Timeout]:", err)
		// }
		return dt, err
	}
	// Vln(5, "addr:", conn.RemoteAddr(), conn.LocalAddr())
	conn.Close()
	return dt, nil
}
