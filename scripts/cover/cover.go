package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var omitDirs = map[string]bool{
	".git":     true,
	"cmd":      true,
	"scripts":  true,
	"testdata": true,
}

func main() {
	gopath := os.Getenv("GOPATH")
	lacodexPath :=
		filepath.Join(gopath, "src", "github.com", "konkers", "lacodex")

	files, err := ioutil.ReadDir(lacodexPath)
	if err != nil {
		log.Fatal(err)
	}

	packages := []string{"."}

	for _, f := range files {
		if f.IsDir() {
			if _, ok := omitDirs[f.Name()]; !ok {
				packages = append(packages, "./"+f.Name())
			}
		}
	}
	coverpkg := fmt.Sprintf("-coverpkg=%s", strings.Join(packages, ","))

	cmd := exec.Command("go", "test", "./...",
		"-coverprofile=coverage.out", coverpkg)
	cmd.Dir = lacodexPath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		fmt.Printf("Command run failed: %v\n", err)
		os.Exit(1)
	}
}
