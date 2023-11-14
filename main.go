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
		log.Fatal("git not found")
	}

	if !binExists("go") {
		log.Fatal("go not found")
	}
}

var (
	version     = "unknown"
	versionFlag bool
	help        bool
	verbose     bool
	debug       bool // currently unused

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
		logErr("%v", err)
		flag.Usage()
		os.Exit(1)
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
		logErr("get current dir failed")
		return err
	}

	if projectName == "" {
		return errors.New("no project name specified")
	}
	projectName = filepath.Base(projectName)

	targetDir = filepath.Join(wd, projectName)
	if _, err := os.Stat(targetDir); !os.IsNotExist(err) {
		logErr("target dir %s already exist", targetDir)
		return err
	}

	logInfo("Creating project in %s\n", targetDir)

	if err := os.MkdirAll(targetDir, 0755); err != nil {
		logErr("create target dir %s failed", targetDir)
		return err
	}

	// Init Go Module
	logInfo("Init Go Module")
	if err := initGoMod(); err != nil {
		logErr("init go mod failed")
		return err
	}

	// Init Git Repo
	logInfo("Init Git Repo")
	if err := initGitRepo(); err != nil {
		logErr("init git repo failed")
		return err
	}

	// Create .gitignore
	logInfo("Create .gitignore")
	if err := createGitignore(); err != nil {
		logErr("create .gitignore failed")
		return err
	}

	// Create README.md
	logInfo("Create README.md")
	if err := createReadme(); err != nil {
		logErr("create README.md failed")
		return err
	}

	// Create main.go
	logInfo("Create main.go")
	if err := createMainDotGo(); err != nil {
		logErr("create main.go failed")
		return err
	}

	if withTaskfile {
		if !binExists("task") {
			logWarn("task binary is not exit on this system")
		}

		logInfo("Create Taskfile.yml")
		if err := createTaskfile(); err != nil {
			logWarn("create Taskfile.yml failed: %v", err)
		}
	}

	if withMakefile {
		if !binExists("make") {
			logWarn("make binary is not exit on this system")
		}

		logInfo("Create Makefile")
		if err := createMakefile(); err != nil {
			logWarn("create Makefile failed")
		}
	}

	if withDockerfile {
		if !binExists("docker") {
			logWarn("docker binary is not exit on this system")
		}

		logInfo("Create Dockerfile")
		if err := createDockerfile(); err != nil {
			logWarn("create Dockerfile failed")
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
	logDebug("Set go module to %s", module)
	return execCommand("go", "mod", "init", module)
}

func createTaskfile() error {
	tf, err := tplFiles.ReadFile("templates/Taskfile.yml")
	if err != nil {
		return err
	}

	tpl, err := template.New("taskfile").Delims("[[", "]]").Parse(string(tf))
	if err != nil {
		return err
	}

	buf := &bytes.Buffer{}
	err = tpl.Execute(buf, struct{ ProjectName string }{projectName})
	if err != nil {
		return err
	}

	taskfilePath := filepath.Join(targetDir, "Taskfile.yml")
	_, err = os.Stat(taskfilePath)
	if !os.IsNotExist(err) {
		return err
	}

	return writeStringToFile(taskfilePath, buf.String())
}

func createDockerfile() error {
	df, err := tplFiles.ReadFile("templates/Dockerfile")
	if err != nil {
		return err
	}

	tpl, err := template.New("dockerfile").Parse(string(df))
	if err != nil {
		return err
	}

	buf := &bytes.Buffer{}
	err = tpl.Execute(buf, struct{ ProjectName string }{projectName})
	if err != nil {
		return err
	}

	dockerfilePath := filepath.Join(targetDir, "Dockerfile")
	_, err = os.Stat(dockerfilePath)
	if !os.IsNotExist(err) {
		return err
	}

	return writeStringToFile(dockerfilePath, buf.String())
}

func createMakefile() error {
	t, err := template.New("makefile").Parse(makefile)
	if err != nil {
		return err
	}

	buf := &bytes.Buffer{}
	err = t.Execute(buf, struct{ ProjectName string }{projectName})
	if err != nil {
		return err
	}

	makefilePath := filepath.Join(targetDir, "Makefile")
	_, err = os.Stat(makefilePath)
	if !os.IsNotExist(err) {
		return err
	}

	return writeStringToFile(makefilePath, buf.String())
}

func createReadme() error {
	t, err := template.New("readme").Parse(readme)
	if err != nil {
		return err
	}

	buf := &bytes.Buffer{}
	t.Execute(buf, struct{ ProjectName string }{projectName})

	readmePath := filepath.Join(targetDir, "README.md")
	_, err = os.Stat(readmePath)
	if !os.IsNotExist(err) {
		return err
	}

	return writeStringToFile(readmePath, buf.String())
}

func createGitignore() error {
	gitignorePath := filepath.Join(targetDir, ".gitignore")
	_, err := os.Stat(gitignorePath)
	if !os.IsNotExist(err) {
		return err
	}

	return writeStringToFile(gitignorePath, gitignore)
}

func createMainDotGo() error {
	t, err := template.New("main.go").Parse(mainGoFile)
	if err != nil {
		return err
	}

	buf := new(bytes.Buffer)
	err = t.Execute(buf, struct{ ProjectName string }{projectName})
	if err != nil {
		return err
	}

	mainPath := filepath.Join(targetDir, "main.go")
	_, err = os.Stat(mainPath)
	if !os.IsNotExist(err) {
		return err
	}

	return writeStringToFile(mainPath, buf.String())
}

func execCommand(cmd string, args ...string) error {
	c := exec.Command(cmd, args...)
	c.Dir = targetDir
	err := c.Run()
	if err != nil {
		return err
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
		return err
	}
	defer f.Close()

	_, err = f.WriteString(content)
	if err != nil {
		return err
	}
	return nil
}
