# KayTrade

A comprehensive trading platform built with Go, featuring a robust client-server architecture for managing and executing trading operations in the terminal.

## 📋 Table of Contents

- [Overview](#overview)
- [Features](#features)
- [Architecture](#architecture)
- [Installation](#installation)
- [Getting Started](#getting-started)
- [Project Structure](#project-structure)
- [Usage](#usage)
- [Configuration](#configuration)
- [Development](#development)
- [Contributing](#contributing)
- [License](#license)

## 🎯 Overview

KayTrade is a modern trading platform designed to provide efficient and reliable trading capabilities. Built entirely in Go, it leverages the language's concurrency features and performance characteristics to deliver a responsive trading experience straight from within your terminal.

## ✨ Features

- **Client-Server Architecture**: Separated concerns with dedicated client and server components
- **High Performance**: Built with Go for optimal speed and efficiency
- **Scalable Design**: Modular architecture supporting future enhancements
- **Command-Line Interface**: Easy-to-use TUI for trading operations
- **Cross-Platform**: Works on Linux, macOS, and *maybe* Windows

## 🏗️ Architecture

KayTrade consists of two main components:

### Client
The client application (`kay_trade`) provides the user interface and handles communication with the server.

### Server
The server component manages trading logic, data persistence, and client connections.

## 📦 Installation

### Prerequisites

- Go 1.21 or higher
- Git

### Method 1: Install via `go install` (Recommended)

#### Install Specific Version (v0.1 - Submission Version)
```sh
go install github.com/Phantomvv1/KayTrade/client/cmd/kay_trade@v0.1
```

#### Install Latest Stable Release
```sh
go install github.com/Phantomvv1/KayTrade/client/cmd/kay_trade@latest
```

#### Install from Source (Development Version)
```sh
go install github.com/Phantomvv1/KayTrade/client/cmd/kay_trade
```

### Method 2: Install via docker

1. Clone the repository:
```sh
git clone https://github.com/Phantomvv1/KayTrade
cd KayTrade
```
2. Go into the client directory and run the Dockerfile with the -it flag:
```sh
cd client

docker build -t kaytrade ./client
docker run -it --rm kaytrade-client
```

### Method 3: Build from Source

1. Clone the repository:
```sh
git clone https://github.com/Phantomvv1/KayTrade
cd KayTrade
```

2. Build the client:
```sh
cd client/cmd/kay_trade
go build -o kaytrade
```

3. Build the server:
```sh
cd ../../../server/cmd/kay_trade
go build -o kaytrade-server
```
or

```sh
cd ../../../server/cmd/kay_trade
go run main.go
```

## 🚀 Getting Started

### Quick Start

1. **Start the Server**:
```sh
# Navigate to server directory
cd server
./kaytrade-server
```

2. **Run the Client**:
```sh
# In a new terminal
kay_trade
```

### Verify Installation

Check that `kay_trade` is properly installed:
```sh
kay_trade --version
```

## 📁 Project Structure

```
KayTrade/
├── client/                  # Client application
│   ├── cmd/                 # Command-line interface
│   │   └── kay_trade/       # Main client executable
│   ├── internal/            # Client packages
│   └── ...
├── server/                  # Server application
│   ├── cmd/                 # Server entry point
│   ├── internal/            # Server packages
│   └── ...
└── README.md
```

## 💻 Usage

### Basic Commands

```sh
kaytrade
```

## 🛠️ Development

### Setting Up Development Environment

1. Fork the repository
2. Clone your fork:
```sh
git clone https://github.com/YOUR_USERNAME/KayTrade.git
cd KayTrade
```

3. Install dependencies:
```sh
go mod tidy
```

### Running in Development Mode

```sh
cd server/cmd/kay_trade
KAYTRADE_ENV=dev go run main.go

cd client/cmd/kay_trade
go run main.go
```

### Code Style

This project follows standard Go conventions:
- Run `gofmt` before committing
- Follow [Effective Go](https://golang.org/doc/effective_go) guidelines
- Write tests for new features

## 🤝 Contributing

At the moment I don't accept contributions due to this being my diploma project. In the future after it gets graded I will start accepting contributions. When that happens please feel free to submit a Pull Request. For major changes, please open an issue first to discuss what you would like to change.

### Contribution Guidelines

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/AmazingFeature`)
3. Commit your changes (`git commit -m 'Add some AmazingFeature'`)
4. Push to the branch (`git push origin feature/AmazingFeature`)
5. Open a Pull Request

### Reporting Issues

If you find a bug or have a suggestion, please open an issue with:
- A clear description of the problem
- Steps to reproduce (for bugs)
- Expected vs actual behavior
- Your environment (OS, Go version, etc.)

## 📄 License

This project is part of a diploma project. Please contact the repository owner for licensing information.

## 👤 Author

**Phantomvv1**
- GitHub: [@Phantomvv1](https://github.com/Phantomvv1)

## 🙏 Acknowledgments

This project was developed as part of a diploma project submission.

---

**Note**: This is a diploma project. For educational purposes only.

## 📞 Support

For questions or support, please open an issue on the GitHub repository.

---

<div align="center">
Made with ❤️ for diploma project
</div>
