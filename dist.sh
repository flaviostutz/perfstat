#!/bin/sh

echo "Removing previous dist"
rm /dist/perfstat-linux-amd64
rm /dist/perfstat-linux-raspberry
rm /dist/perfstat-darwin-amd64
rm /dist/perfstat-windows-amd64.exe

set -e

echo "Compile for linux-amd64"
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -a -installsuffix cgo -o /dist/perfstat-linux-amd64
chmod +x /dist/perfstat-linux-amd64
echo "Saved to /dist/perfstat-linux-amd64"

echo "Compile for linux-arm (works on Raspberry)"
GOOS=linux GOARCH=arm GOARM=5 CGO_ENABLED=0 go build -a -installsuffix cgo -o /dist/perfstat-linux-raspberry
chmod +x /dist/perfstat-linux-amd64
echo "Saved to /dist/perfstat-linux-amd64"

echo "Compile for darwin-amd64"
GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -a -installsuffix cgo -o /dist/perfstat-darwin-amd64
chmod +x /dist/perfstat-darwin-amd64
echo "Saved to /dist/perfstat-darwin-amd64"

echo "Compile for windows-amd64"
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -a -installsuffix cgo -o /dist/perfstat-windows-amd64.exe
echo "Saved to /dist/perfstat-windows-amd64"

