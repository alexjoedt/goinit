package main

import (
	"bytes"
	"embed"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"
)

func init() {
	if !binExists("git") {
		log.Fatal("git command not found in PATH. Please install Git: https://git-scm.com/downloads")
	}

	if !binExists("go") {
		log.Fatal("go command not found in PATH. Please install Go: https://golang.org/doc/install")
	}
}

var (
	version     = "unknown"
	versionFlag bool
	help        bool
	verbose     bool

	withTaskfile   bool
	withMakefile   bool
	withDockerfile bool

	targetDir   string
	projectName string
	moduleName  string
)

const (
	usage = `Usage: goinit [OPTIONS] <project-name>

Options:
  -t, --taskfile		init project with a Taskfile
  -m, --makefile 		init project with a Makefile
  -d, --dockerfile 	init project with a Dockerfile
  -gm, --module 		go module name
  -v, --verbose 		prints detailed logs
  -h, --help				prints this help message
  --version					prints the version
`
)

func main() {
	flag.BoolVar(&help, "help", false, "prints help message")
	flag.BoolVar(&help, "h", false, "prints help message")

	flag.BoolVar(&verbose, "verbose", false, "prints detailed logs")
	flag.BoolVar(&verbose, "v", false, "prints detailed logs")

	flag.BoolVar(&versionFlag, "version", false, "prints the version")

	flag.BoolVar(&withTaskfile, "taskfile", false, "init project with a Taskfile")
	flag.BoolVar(&withTaskfile, "t", false, "init project with a Taskfile")

	flag.BoolVar(&withMakefile, "makefile", false, "init project with a Makefile")
	flag.BoolVar(&withMakefile, "m", false, "init project with a Makefile")

	flag.BoolVar(&withDockerfile, "dockerfile", false, "init project with a Dockerfile")
	flag.BoolVar(&withDockerfile, "d", false, "init project with a Dockerfile")

	flag.StringVar(&moduleName, "module", projectName, "sets the go module name (default: project-name)")
	flag.StringVar(&moduleName, "gm", projectName, "sets the go module name (default: project-name)")

	flag.Parse()

	flag.Usage = func() { fmt.Print(usage) }

	if help {
		flag.Usage()
		os.Exit(0)
	}

	if versionFlag {
		fmt.Println(version)
		os.Exit(0)
	}

	projectName = flag.Arg(0)

	if err := run(); err != nil {
		fatal("%v", err)
	}
}

// Templates
var (
	//go:embed templates
	tplFiles embed.FS
)

// run creates the go project directory
func run() error {

	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current working directory: %w", err)
	}

	if projectName == "" {
		return errors.New("project name is required. Usage: goinit [OPTIONS] <project-name>")
	}
	projectName = filepath.Base(projectName)

	targetDir = filepath.Join(wd, projectName)
	if _, err := os.Stat(targetDir); !os.IsNotExist(err) {
		return fmt.Errorf("target directory '%s' already exists. Please choose a different project name or remove the existing directory", targetDir)
	}

	info("Creating project in %s", targetDir)

	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("failed to create project directory '%s': %w", targetDir, err)
	}

	// Init Go Module
	info("Initializing Go module")
	if err := initGoMod(); err != nil {
		return fmt.Errorf("failed to initialize Go module: %w", err)
	}

	// Init Git Repo
	info("Initializing Git repository")
	if err := initGitRepo(); err != nil {
		return fmt.Errorf("failed to initialize Git repository: %w", err)
	}

	// Create .gitignore
	info("Creating .gitignore")
	if err := createGitignore(); err != nil {
		return fmt.Errorf("failed to create .gitignore file: %w", err)
	}

	// Create README.md
	info("Creating README.md")
	if err := createReadme(); err != nil {
		return fmt.Errorf("failed to create README.md file: %w", err)
	}

	// Create main.go
	info("Creating main.go")
	if err := createMainDotGo(); err != nil {
		return fmt.Errorf("failed to create main.go file: %w", err)
	}

	if withTaskfile {
		if !binExists("task") {
			warn("task binary not found in PATH. You can install it from: https://taskfile.dev/installation/")
		}

		info("Creating Taskfile.yml")
		if err := createTaskfile(); err != nil {
			warn("failed to create Taskfile.yml: %v", err)
		}
	}

	if withMakefile {
		if !binExists("make") {
			warn("make binary not found in PATH. You may need to install build tools for your system")
		}

		info("Creating Makefile")
		if err := createMakefile(); err != nil {
			warn("failed to create Makefile: %v", err)
		}
	}

	if withDockerfile {
		if !binExists("docker") {
			warn("docker binary not found in PATH. You can install it from: https://docs.docker.com/get-docker/")
		}

		info("Creating Dockerfile")
		if err := createDockerfile(); err != nil {
			warn("failed to create Dockerfile: %v", err)
		}
	}

	return nil
}

func initGitRepo() error {
	return execCommand("git", "init", "--initial-branch=main")
}

func initGoMod() error {
	module := projectName
	if moduleName != "" {
		module = moduleName
	}
	return execCommand("go", "mod", "init", module)
}

func createTaskfile() error {
	tf, err := tplFiles.ReadFile("templates/Taskfile.yml")
	if err != nil {
		return fmt.Errorf("failed to read Taskfile.yml template: %w", err)
	}

	tpl, err := template.New("taskfile").Delims("[[", "]]").Parse(string(tf))
	if err != nil {
		return fmt.Errorf("failed to parse Taskfile.yml template: %w", err)
	}

	buf := &bytes.Buffer{}
	err = tpl.Execute(buf, struct{ ProjectName string }{projectName})
	if err != nil {
		return fmt.Errorf("failed to execute Taskfile.yml template: %w", err)
	}

	taskfilePath := filepath.Join(targetDir, "Taskfile.yml")
	if _, err = os.Stat(taskfilePath); !os.IsNotExist(err) {
		return fmt.Errorf("taskfile.yml already exists at '%s'", taskfilePath)
	}

	if err := writeStringToFile(taskfilePath, buf.String()); err != nil {
		return fmt.Errorf("failed to write Taskfile.yml: %w", err)
	}

	return nil
}

func createDockerfile() error {
	df, err := tplFiles.ReadFile("templates/Dockerfile")
	if err != nil {
		return fmt.Errorf("failed to read Dockerfile template: %w", err)
	}

	tpl, err := template.New("dockerfile").Parse(string(df))
	if err != nil {
		return fmt.Errorf("failed to parse Dockerfile template: %w", err)
	}

	buf := &bytes.Buffer{}
	err = tpl.Execute(buf, struct{ ProjectName string }{projectName})
	if err != nil {
		return fmt.Errorf("failed to execute Dockerfile template: %w", err)
	}

	dockerfilePath := filepath.Join(targetDir, "Dockerfile")
	if _, err = os.Stat(dockerfilePath); !os.IsNotExist(err) {
		return fmt.Errorf("dockerfile already exists at '%s'", dockerfilePath)
	}

	if err := writeStringToFile(dockerfilePath, buf.String()); err != nil {
		return fmt.Errorf("failed to write Dockerfile: %w", err)
	}

	return nil
}

func createMakefile() error {
	mf, err := tplFiles.ReadFile("templates/Makefile")
	if err != nil {
		return fmt.Errorf("failed to read Makefile template: %w", err)
	}

	t, err := template.New("makefile").Parse(string(mf))
	if err != nil {
		return fmt.Errorf("failed to parse Makefile template: %w", err)
	}

	buf := &bytes.Buffer{}
	err = t.Execute(buf, struct{ ProjectName string }{projectName})
	if err != nil {
		return fmt.Errorf("failed to execute Makefile template: %w", err)
	}

	makefilePath := filepath.Join(targetDir, "Makefile")
	if _, err = os.Stat(makefilePath); !os.IsNotExist(err) {
		return fmt.Errorf("makefile already exists at '%s'", makefilePath)
	}

	if err := writeStringToFile(makefilePath, buf.String()); err != nil {
		return fmt.Errorf("failed to write Makefile: %w", err)
	}

	return nil
}

func createReadme() error {
	rd, err := tplFiles.ReadFile("templates/README.md")
	if err != nil {
		return fmt.Errorf("failed to read README.md template: %w", err)
	}

	t, err := template.New("readme").Parse(string(rd))
	if err != nil {
		return fmt.Errorf("failed to parse README.md template: %w", err)
	}

	buf := &bytes.Buffer{}
	err = t.Execute(buf, struct{ ProjectName string }{projectName})
	if err != nil {
		return fmt.Errorf("failed to execute README.md template: %w", err)
	}

	readmePath := filepath.Join(targetDir, "README.md")
	if _, err = os.Stat(readmePath); !os.IsNotExist(err) {
		return fmt.Errorf("README.md already exists at '%s'", readmePath)
	}

	if err := writeStringToFile(readmePath, buf.String()); err != nil {
		return fmt.Errorf("failed to write README.md: %w", err)
	}

	return nil
}

func createGitignore() error {
	gi, err := tplFiles.ReadFile("templates/gitignore")
	if err != nil {
		return fmt.Errorf("failed to read gitignore template: %w", err)
	}

	gitignorePath := filepath.Join(targetDir, ".gitignore")
	if _, err := os.Stat(gitignorePath); !os.IsNotExist(err) {
		return fmt.Errorf(".gitignore already exists at '%s'", gitignorePath)
	}

	if err := writeStringToFile(gitignorePath, string(gi)); err != nil {
		return fmt.Errorf("failed to write .gitignore: %w", err)
	}

	return nil
}

func createMainDotGo() error {
	mg, err := tplFiles.ReadFile("templates/main.go")
	if err != nil {
		return fmt.Errorf("failed to read main.go template: %w", err)
	}

	t, err := template.New("main.go").Parse(string(mg))
	if err != nil {
		return fmt.Errorf("failed to parse main.go template: %w", err)
	}

	buf := new(bytes.Buffer)
	err = t.Execute(buf, struct{ ProjectName string }{projectName})
	if err != nil {
		return fmt.Errorf("failed to execute main.go template: %w", err)
	}

	mainPath := filepath.Join(targetDir, "main.go")
	if _, err = os.Stat(mainPath); !os.IsNotExist(err) {
		return fmt.Errorf("main.go already exists at '%s'", mainPath)
	}

	if err := writeStringToFile(mainPath, buf.String()); err != nil {
		return fmt.Errorf("failed to write main.go: %w", err)
	}

	return nil
}

func execCommand(cmd string, args ...string) error {
	c := exec.Command(cmd, args...)
	c.Dir = targetDir
	
	if output, err := c.CombinedOutput(); err != nil {
		if len(output) > 0 {
			return fmt.Errorf("command '%s %v' failed: %s", cmd, args, string(output))
		}
		return fmt.Errorf("command '%s %v' failed: %w", cmd, args, err)
	}
	
	return nil
}

func binExists(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

func writeStringToFile(path string, content string) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create file '%s': %w", path, err)
	}
	
	defer func() {
		_ = f.Close()
	}()

	if _, err = f.WriteString(content); err != nil {
		return fmt.Errorf("failed to write content to file '%s': %w", path, err)
	}
	return nil
}
