#!/bin/bash

# echo "TEST stats"
# cd /app/stats && go test -v

# echo "TEST detectors"
# cd /app/detectors && go test -v

cd /detectors
echo "Launching stress-ng to cause bottlenecks on system"
stress-ng 

echo "Launching tests to (hopefully) detect bottlenecks"
go test -v

