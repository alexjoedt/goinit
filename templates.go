package main

// makefile Makefile template
const makefile = `## help: print this help message
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'
	
## tidy: format code and tidy modfile
.PHONY: tidy
tidy:
	go fmt ./...
	go mod tidy -v
	
## audit: run quality control checks
.PHONY: audit
audit:
	go vet ./...
	go test -race -vet=off ./...
	go mod verify
	
## build: build the {{.ProjectName}} application
.PHONY: build
build:
	go mod verify
	go build -ldflags='-s' -o=./bin/{{.ProjectName}} .
	GOOS=linux GOARCH=amd64 go build -ldflags='-s' -o=./bin/linux_amd64/{{.ProjectName}} .
	
## run: run the {{.ProjectName}} application
.PHONY: run
run: build
	./bin/{{.ProjectName}}

`

// gitignore .gitignore template
const gitignore = `
# Binaries for programs and plugins
*.exe
*.exe~
*.dll
*.so
*.dylib

# Test binary, built with 'go test -c'
*.test

# Output of the go coverage tool, specifically when used with LiteIDE
*.out

# Dependency directories (remove the comment below to include it)
# vendor/
bin/
dist/

# Dotfiles
.env
.envrc
`

// readme README.md template
const readme = `# {{.ProjectName}}

`

// mainGoFile main.go template
const mainGoFile = `package main

import "fmt"

func main() {
	fmt.Println("Hello {{.ProjectName}}")
}

`
