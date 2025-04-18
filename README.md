
# Improved TCP Echo Server

## Overview
An enhanced TCP echo server implementing concurrency, logging, command handling, and timeout features for Test #2.

## Features
- 🚀 Concurrent client handling using goroutines
- 📝 Connection/disconnection logging
- 💾 Message logging to individual client files
- ⚙️ Command-line configuration (port/timeout)
- ⏳ 30-second inactivity timeout
- 📏 Message length validation (1024 byte limit)
- 😊 Personality responses ("hello", "bye")
- ⌨️ Command protocol (/time, /quit, /echo)

## How to Run

### Prerequisites
- Go 1.16+ installed
- Netcat or telnet for testing

### Running the Server
1. **Build the executable:**
   ```bash
   go build -o echo-server
2. **Second Terminal - connect a client:**
   nc localhost 4808

***Testing Commands:***
hello            # Returns "Hi there!"
bye              # Returns "Goodbye!" and disconnects
/time            # Returns current server time
/echo [message]  # Returns the message back
/quit            # Closes the connection

Demo Video
