## help: print this help message
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
	go build -ldflags='-s' -o=./bin/{{.ProjectName}} {{.BuildPath}}
	GOOS=linux GOARCH=amd64 go build -ldflags='-s' -o=./bin/linux_amd64/{{.ProjectName}} {{.BuildPath}}
	
## run: run the {{.ProjectName}} application
.PHONY: run
run: build
	./bin/{{.ProjectName}}
