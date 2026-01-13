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
- [Development](#development)
- [Contributing](#contributing)
- [License](#license)

<a name="overview"/>

## 🎯 Overview

KayTrade is a modern trading platform designed to provide efficient and reliable trading capabilities. Built entirely in Go, it leverages the language's concurrency features and performance characteristics to deliver a responsive trading experience straight from within your terminal.

<a name="features"/>

## ✨ Features

- **Client-Server Architecture**: Separated concerns with dedicated client and server components
- **High Performance**: Built with Go for optimal speed and efficiency
- **Scalable Design**: Modular architecture supporting future enhancements
- **Terminal User Interface**: Easy-to-use TUI for trading operations
- **Cross-Platform**: Works on Linux, macOS, and *maybe* Windows

<a name="architecture"/>

## 🏗️ Architecture

KayTrade consists of two main components:

### Client
The client application (`kaytrade`) provides the user interface and handles communication with the server.

### Server
The server component manages trading logic, data persistence, and client connections.

<a name="installation"/>

## 📦 Installation

### Prerequisites

- Go 1.25 or higher
- Git

### Method 1: Install via `go install` (Recommended)

#### Install Latest Stable Release
```sh
go install github.com/Phantomvv1/KayTrade/client/cmd/kaytrade@latest
```
> **Note**: The binary will be installed as kaytrade in your $GOPATH/bin directory. If that is not set up, go will install it in your $HOME/go/bin directory. This means that you won't be able to run it instantly. What you need to do in that case is go into that directory and move the binary to one of the /bin directories on your PC or the bin directory of your go installation. Usually that is /usr/local/go/bin. Here is how you do that:
```sh
cd ~
cd go/bin
# You have to use sudo since this is a root directory

sudo mv kaytrade /bin
# or
sudo mv kaytrade /usr/bin
# or
sudo mv kaytrade /usr/local/bin
# or
sudo mv kaytrade /usr/local/go/bin
```

### Method 2: Download the binary straight from github

Go to the releases, click on any release you like (preferably the last one). Go down and download the binary named kaytrade. After that give it rights to be executed and put it in one of the /bin in your PC as shown in the note above. That's it.

```sh
chmod +x kaytrade
./kaytrade -v
```

### Method 3: Install via Docker

1. Clone the repository:
```sh
git clone https://github.com/Phantomvv1/KayTrade
cd KayTrade
```
2. Build and run the Docker container:
```sh
cd client

docker build -t kaytrade .
docker run -it --rm kaytrade
```

### Method 4: Build from Source

1. Clone the repository:
```sh
git clone https://github.com/Phantomvv1/KayTrade
cd KayTrade
```

2. Build the client:
```sh
cd client/cmd/kaytrade
go build -o kaytrade
./kaytrade
```

or

```sh
cd client/cmd/kaytrade
go run kaytrade
```

<a name="getting-started"/>

## 🚀 Getting Started

### Quick Start

Simply launch the application:
```sh
kaytrade
```
The TUI will guide you through the trading interface.

### Verify Installation

Check that `kay_trade` is properly installed:
```sh
kaytrade --version
# or
kaytrade -v
```

<a name="project-structure"/>

## 📁 Project Structure

```
KayTrade/
├── client/                  # Client application
│   ├── cmd/                 # Command-line interface
│   │   └── kaytrade/        # Main client executable
│   ├── internal/            # Client packages
│   ├── Dockerfile           # Client container definition
│   └── ...
├── server/                  # Server application
│   ├── cmd/                 # Server entry point
│   │   └── kaytrade/        # Main server executable
│   ├── internal/            # Server packages
│   └── ...
└── README.md
```

<a name="usage"/>

## 💻 Usage

### Launching KayTrade

```sh
kaytrade
```

Once launched, navigate the TUI using:
- Arrow keys or vim keybindings for navigation
- Enter to select
- ESC to go back
- q to quit

### Command-Line Options

```sh
kaytrade --version    # Display version information
```

<a name="development"/>

## 🛠️ Development

### Setting Up Development Environment

1. Fork the repository
2. Clone your fork:
```sh
git clone https://github.com/YOUR_USERNAME/KayTrade
cd KayTrade
```

3. Install dependencies:
```sh
go mod tidy
```
4. Database and cache:

Install Postgres version 17/18 and Redis/Valkey version 8

### Running in Development Mode

```sh
cd server/cmd/kaytrade
KAYTRADE_ENV=dev go run main.go

cd client/cmd/kaytrade
go run main.go
```

### Code Style

This project follows standard Go conventions:
- Run `gofmt` before committing
- Follow [Effective Go](https://golang.org/doc/effective_go) guidelines
- Write tests for new features

<a name="contributing"/>

## 🤝 Contributing

**Note**: This repository is currently not accepting contributions as it is part of an ongoing diploma project. After the project is graded and submitted, contributions will be welcomed.

Once contributions are open, feel free to submit Pull Requests. For major changes, please open an issue first to discuss what you would like to change.

### Future Contribution Guidelines

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

<a name="license"/>

## 📄 License

This project is part of a diploma project. Please contact the repository owner for licensing information.

## 👤 Author

**Phantomvv1**
- GitHub: [Phantomvv1](https://github.com/Phantomvv1)

## 🙏 Acknowledgments

This project was developed as part of a diploma project submission.

## 📞 Support

For questions or support, please open an issue on the GitHub repository.

---

<div align="center">
Made with ❤️ for diploma project
</div>
