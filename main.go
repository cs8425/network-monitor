package main

import (
	"errors"
	"flag"
	"net"
	"sync"
	"syscall"
	"time"

	"golang.org/x/sys/unix"
)

const (
	SYN_INIT_TIMEOUT = 1000 - 5
)

var (
	verbosity = flag.Int("v", 2, "verbosity")
	bind      = flag.String("l", ":8080", "max worker for RTT test")

	timeout     = flag.Int("to", SYN_INIT_TIMEOUT, "SYN timeout in ms")
	dt          = flag.Int("dt", 5000, "interval between test in ms")
	workerCount = flag.Int("w", 6, "max worker for RTT test")

	target     = flag.String("t", "", "target addr with port")
	targetFile = flag.String("f", "target.txt", "target list file")
)

var (
	targetList = []*PingTarget{
		{"vultr_jp_syn", "hnd-jp-ping.vultr.com:80"},
		{"vultr_la_syn", "lax-ca-us-ping.vultr.com:80"},
		{"vultr_jp_rst", "hnd-jp-ping.vultr.com:23"},
		{"vultr_la_rst", "lax-ca-us-ping.vultr.com:23"},

		{"1.1.1.1_syn", "1.1.1.1:80"},
		{"8.8.8.8_syn", "8.8.8.8:80"},
	}
)

type PingTarget struct {
	Name string // for Label
	Addr string
}

func main() {
	flag.Parse()

	// for single test
	if *target != "" {
		// v, e := net.ResolveTCPAddr("tcp", "hnd-jp-ping.vultr.com:80")
		// Vln(1, v, e)

		dst := *target
		to := time.Duration(*timeout) * time.Millisecond
		d := &net.Dialer{
			Timeout: to,
			Control: Control,
		}
		dt, err := ping(d, dst)
		Vln(1, "[dt]", dst, dt, err)
		return
	}

	// for service
	go run(targetList, *workerCount)

	WebStart(*bind)
}

func run(targetList []*PingTarget, workerCount int) {
	var wg sync.WaitGroup
	queueCh := make(chan *PingTarget, workerCount)
	for i := 0; i < workerCount; i += 1 {
		wg.Add(1)
		go pingWorker(&wg, queueCh)
	}

	Vln(3, "[doTest]start")
	for {
		nextTime := time.Now().Add(time.Duration(*dt) * time.Millisecond)
		Vln(3, "[doTest]next:", nextTime)

		// TODO: hot reload list
		for _, dst := range targetList {
			queueCh <- dst
		}
		time.Sleep(time.Until(nextTime))
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

func pingWorker(wg *sync.WaitGroup, queueCh chan *PingTarget) {
	defer wg.Done()

	to := time.Duration(*timeout) * time.Millisecond
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

func Control(network, address string, c syscall.RawConn) (err error) {
	// Vln(5, "[Control]", network, address, c)
	return c.Control(func(fd uintptr) {
		err = unix.SetsockoptInt(int(fd), unix.IPPROTO_IP, unix.IP_BIND_ADDRESS_NO_PORT, 1)
		// err = unix.SetsockoptInt(int(fd), unix.SOL_SOCKET, unix.SO_REUSEADDR, 1)
		// if err != nil {
		// 	return
		// }
		// err = unix.SetsockoptInt(int(fd), unix.SOL_SOCKET, unix.SO_REUSEPORT, 1)
	})
}
