#!/bin/bash

echo "Testing socket connection with added logging"

# Check if the daemon binary exists
DAEMON_PATH="/usr/bin/free-mind-daemon"
if [ ! -f "$DAEMON_PATH" ]; then
    echo "ERROR: Daemon binary not found at $DAEMON_PATH"
    exit 1
fi
echo "Daemon binary found at $DAEMON_PATH"

# Check if the socket file exists
SOCKET_PATH="/tmp/tech.tanay.free-mind.sock"
if [ -S "$SOCKET_PATH" ]; then
    echo "Socket file exists at $SOCKET_PATH"
    ls -la "$SOCKET_PATH"
else
    echo "Socket file does not exist at $SOCKET_PATH"
    echo "Checking /tmp directory:"
    ls -la /tmp | grep free-mind
fi

# Check if the daemon process is running
PID=$(pgrep -f free-mind-daemon)
if [ -n "$PID" ]; then
    echo "Daemon process is running with PID: $PID"
    ps -p $PID -o pid,ppid,cmd,stat
else
    echo "Daemon process is not running"
fi

# Try to start the daemon if it's not running
if [ -z "$PID" ]; then
    echo "Attempting to start the daemon..."
    # Use nohup to ensure the daemon keeps running even after the script exits
    sudo sh -c "nohup $DAEMON_PATH > /tmp/free-mind-daemon.log 2>&1 &"
    echo "Waiting for daemon to initialize..."
    
    # Wait for the socket file to be created with multiple retries
    MAX_RETRIES=20
    for i in $(seq 1 $MAX_RETRIES); do
        echo "Checking for socket file (attempt $i/$MAX_RETRIES)..."
        sleep 1
        
        if [ -S "$SOCKET_PATH" ]; then
            echo "Socket file created at $SOCKET_PATH"
            ls -la "$SOCKET_PATH"
            break
        elif [ $i -eq $MAX_RETRIES ]; then
            echo "Socket file was not created after $MAX_RETRIES attempts"
            echo "Checking daemon log:"
            if [ -f "/tmp/free-mind-daemon.log" ]; then
                cat /tmp/free-mind-daemon.log
            else
                echo "No daemon log file found"
            fi
        fi
    done
fi

# Run the daemon test client
echo "Running daemon test client..."
cd daemon-test-client
go run main.go

echo "Test completed"