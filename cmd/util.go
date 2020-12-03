// misc utilities

package cmd

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

func initRunDir() (runDir string, err error) {
	os.Mkdir(".jay", 0700)
	runDir, err = ioutil.TempDir(".jay", "")
	for _, dir := range []string{"bin", "tests", "outputs"} {
		os.Mkdir(runDir+"/"+dir, 0700)
	}
	return
}

func compare(file1 string, file2 string) (diff string, err error) {
	cleanOutputs := fmt.Sprintf("<(awk 'NF' %s) <(awk 'NF' %s)", file1, file2)
	_, err = bash("cmp", cleanOutputs)
	if err != nil {
		diff, _ = bash("diff -y -W 40", cleanOutputs, "| colordiff")
	}
	return
}

func showHead(message string, file string) {
	head, _ := bash("head", file)
	linesStr, _ := bash("cat", file, "| wc -l")
	lines, _ := strconv.Atoi(linesStr)
	if lines > 10 {
		fmt.Printf("%s (%s, 10/%d lines shown):\n", message, file, lines)
	} else {
		fmt.Printf("%s:\n", message)
	}
	fmt.Println(head)
}

// bash executes string or slice of strings as command and blocks until completion
func bash(command ...string) (string, error) {
	if verbose {
		fmt.Println("$", strings.Join(command, " "))
	}
	out, err := exec.Command("bash", "-c", strings.Join(command, " ")).Output()
	if err != nil {
		//fmt.Println(err)
		err = errors.New(string(err.(*exec.ExitError).Stderr))
	}
	return string(out), err
}

func isFile(path string) bool {
	info, err := os.Stat(path)
	return !os.IsNotExist(err) && !info.IsDir()
}
