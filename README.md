# perfstat

Analyze and show tips about possible bottlenecks and risks in Linux systems regarding to diskio, networking, cpu, swapping, memory etc.

We decided to create this utility to help on the laborious job of aswering the following:

* "Is the system OK?"
* "Will it be OK?"
* "Do we have to do anything now?"

After hundred of hours looking for metrics on CLI and Prometheus/Grafana tools, correlating data to check if all is ok, now we can automate some of this work. Surely this won't answer all the doubts, but can help you on some repetitive work.

**If you are a system admin**, answering the "Is the system OK" overnight too, come and tell us what you miss from perfstat on the [Issues](https://github.com/flaviostutz/perfstat/issues). Share your experience and automate it forever!

**If you are a developer** too, help system admins find problems more quickly by implementing some of the [Issues](https://github.com/flaviostutz/perfstat/issues) so they can keep your software up! If in doubt, ask for a task in "Issues" and we'd be glad to answer.

Perfstat has various interfaces:

* **CLI**: ```perfstat``` - for local diagnostics
* **Prometheus Exporter**: ```perfstat --prom-enable``` - for remote monitoring
* **Golang lib**: ```go get github.com/flaviostutz/perfstat``` - for using this in something greater

## Usage

### CLI

```sh
perfstat
```

Output

```sh
```

### Prometheus Exporter

* Start exporter

```sh
perfstat --prom-enable --prom-bind 0.0.0.0 --prom-port 8880 --prom-path /metrics
```

* Check metrics

```sh
curl localhost:8880/metrics
```

* Add this exporter to Prometheus configuration

#### Prometheus Metrics

* **issue_criticity** - criticity score for active issues
  * label resource - cpu, mem, disk, net
  * label name - cpu:1, disk-/mnt/test, nic:eth0

* **issue_cpu_perc** - cpu load for active issues
  * label cpu - cpu:1, cpu:total
  * label type - iowait, used

* **issue_mem_perc** - mem perc for active issues
  * label type - ram used, swap used
  
* **issue_disk_perc** - disk storage for active issues
  * label mount - /mnt/test1
  * label type - used

## Issue Detectors

### Bottlenecks (already a problem)

* Low idle CPU (overall) OK
  * top cpu eater processes OK
  * high steal cpu OK
* Low idle CPU (single CPU)
  * top cpu eater processes
* High CPU wait (waiting for IO) OK
  * top io waiter processes OK
  * top "waited" disks OK
* Low available open files descriptors
  * top process by open files
* Disk nr of block read/writes seems to be in a ceil limit OK
  * top disk eater processes
* Disk bandwidth of read/writes seems to be in a ceil limit OK
  * top disk eater processes
* Network interface bandwidth seems to be in a ceil limit
  * top network bandwidth eater processes
* Network interface pps seems to be in a ceil limit
  * top network pps eater processes

### Risks (may cause problems)

* Low RAM
  * top ram eater processes
* Low Disk space
  * mapped device with lowest space
* Low Disk inodes
  * mapped device with lowest inodes
* Low available files to be open (ulimit)
  * top files open eater processes
* RAM memory growing linearly for process - there maybe a memory leak
  * process with growing memory
* High error rate in NIC
  * show processes with most net errors

### Enhancements

* High swap IO - few RAM, may slow down system by using too much disk
* High %util in disk - disk is being hammered and may not handle well spikes when needed

### Insights (top 5)

* Processes with high cpu wait
* Processes with high cpu usage
* Processes with high disk io
* Processes with high nr of open files
* Processes with high network usage
* Destination hosts with high network bandwidth
* Block devices with high reads/writes

## Perfstat developer tips

### Profiling

```golang
//run profile for an specific test case
go test -cpuprofile /tmp/cpu.prof -run ^TestProcessStatsBasic$

//see results in browser
go tool pprof -http 0.0.0.0:5050 /tmp/cpu.prof
```

## Existing tools for performance analysis

* CPU stats by CPU (%idle %wait etc): mpstat -P ALL 2 5
* RAM stats (free, cache, buffer, swap): vmstat -S M 2
* Disk stats by block device (wr/s rd/s etc): iostat -dx 1
* Network bandwidth by host: iftop
* Open files by process: lsof
* CPU usage by process: top
* Disk usage by process: iotop
* Network bandwidth by process: nethogs OR iftop with netstat -tup OR dstat --net --top-io-adv

## More info about performance analysis

* https://blog.appsignal.com/2018/03/06/understanding-cpu-statistics.html
* https://www.cyberciti.biz/tips/linux-resource-utilization-to-detect-system-bottlenecks.html
* http://web.archive.org/web/20101028025942/https://anchor.com.au/hosting/development/HuntingThePerformanceWumpus
* https://www.tecmint.com/command-line-tools-to-monitor-linux-performance/
