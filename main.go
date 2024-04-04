package main

import (
	"bufio"
	"flag"
	"math"
	"net"
	"os"
	"strings"
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

	lines, err := readFile(*targetFile)
	if err != nil {
		Vln(1, "[list][err]read:", err)
		return
	}

	targetList := make([]*PingTarget, 0, 16)
	for i, line := range lines {
		args := strings.SplitN(line, "\t", 2)
		if strings.HasPrefix(args[0], "#") {
			continue
		}
		if len(args) < 2 {
			continue
		}
		Vln(4, "[list]", i, args)

		targetList = append(targetList, &PingTarget{
			args[0],
			args[1],
		})
	}
	Vln(3, "[list]total", len(targetList))

	// calc minimum worker for non-blocking in worst case
	reqps := float64(*dt) / float64(*timeout)
	n := int(math.Ceil(float64(len(targetList)) / reqps))
	if *workerCount < n {
		Vln(2, "[worker]change for avoiding worst case", n, "x", reqps, ">=", len(targetList), ", was", *workerCount)
	} else {
		n = *workerCount
	}

	// for service
	go run(targetList, n)

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

func readFile(path string) ([]string, error) {
	af, err := os.Open(path)
	if err != nil {
		Vln(2, "[open]", err)
		return nil, err
	}
	defer af.Close()

	data := make([]string, 0)
	r := bufio.NewReader(af)
	b, err := r.Peek(3)
	if err != nil {
		return nil, err
	}
	if b[0] == 0xEF && b[1] == 0xBB && b[2] == 0xBF {
		r.Discard(3)
	}
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			break
		}

		line = strings.Trim(line, "\n\r\t")
		data = append(data, line)
	}

	Vln(7, "[dbg][file]", data)
	return data, nil
}
