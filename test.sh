#!/bin/bash

cd /detectors

echo "Launching stress-ng to cause bottlenecks on system"
stress-ng 

echo "Launching tests to (hopefully) detect bottlenecks"
go test -v

