# Testing Plan for IPC Implementation

## Overview

This document outlines the testing plan for the new IPC implementation across different platforms.

## Test Environments

- Linux (Ubuntu/Debian)
- macOS (Darwin)
- Windows

## Test Cases

> Unit tests written for cases 1–5 (`ipc/unix_test.go`, `root-daemon/daemon_test.go`, `app_test.go`, `test-socket-connection.sh`). Cross-platform execution pending.

### 1. Daemon Installation and Startup

**Test Steps:**
1. Run the application
2. Check if the daemon is installed and running
3. If not, install and start the daemon
4. Verify the daemon is running

**Expected Results:**
- The daemon should be installed and running
- The socket/pipe path should be created at the correct location
- The socket/pipe path should be written to the correct file

### 2. Client Connection

**Test Steps:**
1. Run the application
2. Connect to the daemon using the IPC client
3. Verify the connection is successful

**Expected Results:**
- The client should connect to the daemon successfully
- The connection should be established using the correct socket/pipe path

### 3. Message Sending and Receiving

**Test Steps:**
1. Run the application
2. Connect to the daemon using the IPC client
3. Send a message to the daemon
4. Receive the response from the daemon
5. Verify the response is correct

**Expected Results:**
- The message should be sent to the daemon successfully
- The daemon should process the message correctly
- The response should be received by the client successfully
- The response should have the correct format and content

### 4. Site Blocking Functionality

**Test Steps:**
1. Run the application
2. Connect to the daemon using the IPC client
3. Send a message to update the sites to be blocked
4. Send a message to start blocking
5. Verify the sites are blocked
6. Send a message to stop blocking
7. Verify the sites are unblocked

**Expected Results:**
- The sites should be updated in the hosts-list-to-be-blocked file
- The sites should be added to the hosts file when blocking is started
- The sites should be removed from the hosts file when blocking is stopped

### 5. Error Handling

**Test Steps:**
1. Run the application
2. Try to connect to a non-existent daemon
3. Try to send a message to a non-existent daemon
4. Try to send an invalid message to the daemon
5. Try to receive a response from a non-existent daemon

**Expected Results:**
- The application should handle errors gracefully
- The application should provide meaningful error messages
- The application should not crash

## Platform-Specific Tests

### Linux/macOS

**Test Steps:**
1. Verify the Unix socket is created at `/tmp/tech.tanay.free-mind.sock`
2. Verify the socket permissions are set correctly
3. Verify the socket is removed when the daemon is stopped

**Expected Results:**
- The socket should be created at the correct location
- The socket permissions should allow the client to connect
- The socket should be removed when the daemon is stopped

### Windows

**Test Steps:**
1. Verify the named pipe is created at `\\.\pipe\tech.tanay.free-mind`
2. Verify the pipe permissions are set correctly
3. Verify the pipe is closed when the daemon is stopped

**Expected Results:**
- The pipe should be created at the correct location
- The pipe permissions should allow the client to connect
- The pipe should be closed when the daemon is stopped

## Test Tools

- The daemon-test-client can be used to test the daemon functionality
- The application itself can be used to test the client functionality
- System tools can be used to verify the socket/pipe creation and permissions

## Test Execution

- [x] Write unit tests for IPC package (`ipc/unix_test.go`)
- [x] Write unit tests for daemon (`root-daemon/daemon_test.go`)
- [x] Write unit tests for app (`app_test.go`)
- [x] Add shell integration test (`test-socket-connection.sh`)
- [ ] Run the tests on each platform (Linux, macOS, Windows)
- [ ] Document any issues or failures
- [ ] Fix any issues and retest
- [ ] Verify all tests pass on all platforms