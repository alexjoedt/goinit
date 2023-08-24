# goinit

A simple CLI tool to create a new Go project with a very basic directory structure.

Contributions are welcome!

## Project structure

```bash
.
├── .git
├── .gitignore
├── Makefile (optional)
├── README.md
├── Taskfile.yml (optional)
├── go.mod
└── main.go
```

## Usage
    
```bash
$ go install github.com/alexjoedt/goinit
$ goinit my-new-project
```

### Examples

Create project with a Makefile:

```bash
$ goinit -m my-new-project
# or
$ goinit --makefile my-new-project
```

Create project with a Taskfile

```bash
$ goinit -t my-new-project
# or
$ goinit --taskfile my-new-project
```

Sets the module name

```bash
$ goinit -gm my-new-project
# or
$ goinit --module github.com/alexjoedt/goinit my-new-project
```

### Roadmap
- [x] Set the module name
- [x] Create project with a Makefile
- [x] Create project with a Taskfile
- [ ] Create project with a Dockerfile
