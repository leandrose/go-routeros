# Mikrotik API Client (Go)

A simple and powerful Go 1.21+ client for communicating with Mikrotik's binary API (port 8728).
Supports multiple simultaneous commands, response channels, and debug logging.

---

## ✨ Features

- Direct communication with Mikrotik's binary API
- Username/password login
- Automatic `.tag` assignment for each command
- Responses streamed via individual `chan Response`
- Supports multiple concurrent commands (responses are multiplexed)
- `Debug()` method to enable full read/write logging
- Works with any `io.ReadWriteCloser` (great for testing and mocking)

---

## 📦 Installation

```bash
go get github.com/yourusername/mikrotik-go
```
