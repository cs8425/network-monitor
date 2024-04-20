# network-monitor

a service doing TCP ping and export prometheus metrics for monitoring.


## config and run

1. build the service: `go build -ldflags="-s -w" -trimpath .`
2. copy and modify `target.example.txt` to `target.txt`
3. run: `./network-monitor -l :8080 -f target.txt`
4. get report: `curl http://127.0.0.1:8080/metrics`


## TODO

* [ ] hot-reload targets
* [ ] refactor as library
* [ ] api for load/save/reload config
* [ ] target groups with custom metrics path
* [ ] configurable metrics (packet loss, RTT, RTT distribution)
* [ ] https/tls server
