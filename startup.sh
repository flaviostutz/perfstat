#!/bin/sh

if [ "$RUN_TESTS" == "true" ]; then

    # echo "TEST stats"
    # cd /app/stats && go test -v

    # echo "TEST detectors"
    # cd /app/detectors && go test -v

    cd /app/
    echo "Launching stress-ng to cause bottlenecks on system"

    #cpu
    # stress-ng -c 4
    #io
    # stress-ng -i 2
    #disk
    # stress-ng -d 5
    #socket
    # stress-ng -S 5
    # stress-ng --udp 5
    #mem
    # stress-ng --vm 5
    # stress-ng -c 2 -i 1 -m 1 --vm-bytes 128M -t 10s
    # stress-ng --disk 2
    # stress-ng --cpu 4 --io 2 --vm 1 --vm-bytes 1G --timeout 60s

    echo "Launching tests to check for bottleneck detections"
    go test -v -run TestRollingDetections

else

    echo "Starting Perfstat..."
    # perfstat
    # top

    go get github.com/cjbassi/gotop
    echo "GOTOP"
    gotop

fi
