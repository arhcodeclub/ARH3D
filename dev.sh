#!/bin/bash

echo "======================================="
echo "======================================="

echo "[1/2] Starting livereload server on 0.0.0.0:35729..."
go run livereload.go -dirs=internal/http/templates -addr=0.0.0.0:35729 &
LIVERELOAD_PID=$!
sleep 1

echo "[2/2] Starting reflex..."
reflex -r '\.go$|\.html$' --start-service -- sh -c 'go run cmd/server/main.go' &
SERVER_PID=$!
sleep 1

ip=$(ifconfig | grep -Eo 'inet (addr:)?([0-9]*\.){3}[0-9]*' | grep -Eo '([0-9]*\.){3}[0-9]*' | grep -v '127.0.0.1' | head -n 1)

echo ""
echo "Started:"
echo "  - http://localhost:8080"
echo "  - http://$ip:8080"
echo ""
echo "Press Ctrl+C to stop."
echo "======================================="

trap "echo; echo 'Shutting down...'; kill $LIVERELOAD_PID $SERVER_PID" SIGINT

wait
