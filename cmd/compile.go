// this file contains utilities related to compiling and running solutions from source

package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type Language struct {
	compile string // empty compile string means that we just need to copy file to tmp
	run     string // empty run string corresponds to "@"
}

// @ is later replaced with filename
var cpp = Language{compile: "g++ %s --std=c++17 -o @"}
var python = Language{run: "python3 @"}

var knownExtensions = map[string]Language{
	".cpp": cpp,
	".cc":  cpp,
	".py":  python,
}

// parseSolutionString figures out how to compile and run solution
func parseSolutionString(solution string) (compileCmd string, runCmd string) {
	if isFile(solution) {
		// map returns empty strings if key is not found,
		// so in case of binaries both compile and run will be empty
		extension := filepath.Ext(solution)
		language := knownExtensions[extension]
		runCmd = language.run
		if language.compile != "" {
			compileCmd = fmt.Sprintf(language.compile, solution)
		}
	} else {
		// it is either compile string on run string
		if strings.Contains(solution, "@") {
			compileCmd = solution
		} else {
			runCmd = solution
		}
	}

	return
}

// prepareSolution takes solution string and returns a path to a executable
func prepareExecutable(solutionString string, runDir string) (runCmd string, err error) {
	if solutionString == "" {
		return "", nil
	}

	compileCmd, runCmd := parseSolutionString(solutionString)

	if runCmd == "" {
		runCmd = "@"
	}

	// there is no need to do anything if it was a custom run command
	if !strings.Contains(runCmd, "@") {
		return
	} else {
		file, _ := ioutil.TempFile(runDir+"/bin/", "")
		name := file.Name()
		file.Close()
		runCmd = strings.ReplaceAll(runCmd, "@", name)
		if compileCmd != "" {
			strings.ReplaceAll(runCmd, "@", name)
			fmt.Println("Compiling:", compileCmd)
			_, err = bash(strings.ReplaceAll(compileCmd, "@", name))
		} else {
			bash("cp", solutionString, name)
		}
		os.Chmod(name, 0700)
	}

	return
}
