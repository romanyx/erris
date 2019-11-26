package main_test

import (
	"log"
	"os"
	"os/exec"
	"path"
	"strings"
	"testing"
)

const binaryName = "erris"

func TestMain(m *testing.M) {
	err := os.Chdir("..")
	if err != nil {
		log.Fatalf("could not change dir: %v", err)
	}
	dir, err := os.Getwd()
	if err != nil {
		log.Fatalf("run os getwd: %v", err)
	}

	dir = path.Join(dir, binaryName)
	build := exec.Command("go", "build")
	build.Dir = dir
	err = build.Run()
	if err != nil {
		log.Fatalf("could not make binary for %s: %v\n", binaryName, err)
	}

	code := m.Run()

	os.RemoveAll(path.Join(dir, binaryName))

	os.Exit(code)
}

func TestCliArgs(t *testing.T) {
	tt := []struct {
		name      string
		args      []string
		numIssues int
	}{
		{"no arguments", []string{}, 2},
		{"ignore tests", []string{"-ignoretests"}, 1},
	}

	for _, tc := range tt {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			dir, err := os.Getwd()
			if err != nil {
				t.Fatal(err)
			}

			dir = path.Join(dir, binaryName)
			tc.args = append(tc.args, "./testdata")
			cmd := exec.Command(path.Join(dir, binaryName), tc.args...)
			cmd.Dir = dir
			output, err := cmd.CombinedOutput()
			if err == nil {
				t.Error("expected error")
			}

			lines := strings.Split(string(output), "\n")

			// 1 line is blank
			if tc.numIssues != len(lines)-1 {
				t.Errorf(
					"expected: %d lines\nactual: %d lines\n",
					tc.numIssues,
					len(lines)-1,
				)
			}
		})
	}
}
