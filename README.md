# goinit

[![Go Version](https://img.shields.io/github/go-mod/go-version/alexjoedt/goinit)](https://golang.org/)
[![License](https://img.shields.io/github/license/alexjoedt/goinit)](LICENSE)
[![Release](https://img.shields.io/github/v/release/alexjoedt/goinit)](https://github.com/alexjoedt/goinit/releases)

A simple CLI tool to quickly initialize Go projects with a clean directory structure.

## Installation

```bash
go install github.com/alexjoedt/goinit@latest
```

## Usage

```bash
# Create a basic Go project
goinit my-project

# With optional components
goinit -t -m -d my-webapp  # Add Taskfile, Makefile, and Dockerfile

# Interactive mode
goinit --interactive
```

## Options

```
-t, --taskfile     Add Taskfile.yml for task automation
-m, --makefile     Add Makefile for build automation  
-d, --dockerfile   Add Dockerfile for containerization
    --module       Custom Go module name
-i, --interactive  Interactive mode with prompts
-v, --verbose      Show detailed output
-h, --help         Show help message
```

## Project Structure

```
my-project/
├── .git/
├── .gitignore
├── go.mod
├── main.go
├── README.md
├── Makefile      (optional)
├── Taskfile.yml  (optional)
└── Dockerfile    (optional)
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
