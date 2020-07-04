# bottler
Analyze and show tips about possible bottlenecks in Linux systems regarding to diskio, networking, cpu, swapping, memory etc

https://www.cyberciti.biz/tips/linux-resource-utilization-to-detect-system-bottlenecks.html
http://web.archive.org/web/20101028025942/https://anchor.com.au/hosting/development/HuntingThePerformanceWumpus
https://www.tecmint.com/command-line-tools-to-monitor-linux-performance/

## Brainstorm

## Existing tools
* CPU stats by CPU (%idle %wait etc): mpstat -P ALL 2 5
* RAM stats (free, cache, buffer, swap): vmstat -S M 2
* Disk stats by block device (wr/s rd/s etc): iostat -dx 1
* Network bandwidth by host: iftop
* Open files by process: lsof
* CPU usage by process: top
* Disk usage by process: iotop
* Network bandwidth by process: nethogs

#### This IS a problem
* Low idle CPU (overall)
* Low idle CPU (single CPU)
* Low RAM
* High CPU wait (waiting for IO)
* Low available files open (ulimit)
* Low available network bandwidth (when max bandwidth is set by the user)

### This MIGHT be a problem
* Alert high swap io
* Alert high number of processes waiting for CPU

### Inferences

### Top 5 (give tips)
* Processes with high cpu wait
* Processes with high cpu usage
* Processes with high disk io
* Processes with lots of open files

