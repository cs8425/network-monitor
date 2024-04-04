package main

import (
	"flag"
	"net"
	"sync"
	"time"
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

		{"1.1.1.1_syn", "1.1.1.1:443"},
		{"8.8.8.8_syn", "8.8.8.8:443"},
	}
)

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
		go pingWorker(time.Duration(*timeout)*time.Millisecond, &wg, queueCh)
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
