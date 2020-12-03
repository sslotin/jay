package cmd

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
)

func redirect(from *io.ReadCloser, to *io.WriteCloser, log *os.File, prefix string) {
	scanner := bufio.NewScanner(*from)
	for scanner.Scan() {
		line := scanner.Text()
		line += "\n"
		// in this exact order so that race conditions are less likely to happen
		// ideally we would want a mutex or something to guarantee that I and S won't write to it at the same time
		// but the logs are only used for debugging so that's not a critical issue
		(*log).WriteString(prefix + line)
		io.WriteString(*to, line)
	}
}

func testInteractive(solution string, interactor string, log *os.File, seed int) (string, error) {
	timeout := fmt.Sprintf("timeout %.3fs ", timeLimit.Seconds())

	solCmd := exec.Command("bash", "-c", timeout+solution)
	intCmd := exec.Command("bash", "-c", "SEED="+strconv.Itoa(seed)+" "+timeout+interactor)

	solErr := new(bytes.Buffer)
	solCmd.Stderr = solErr

	intErr := new(bytes.Buffer)
	intCmd.Stderr = intErr

	solIn, _ := solCmd.StdinPipe()
	intIn, _ := intCmd.StdinPipe()
	solOut, _ := solCmd.StdoutPipe()
	intOut, _ := intCmd.StdoutPipe()

	go redirect(&intOut, &solIn, log, "i: ")
	go redirect(&solOut, &intIn, log, "s: ")

	solCmd.Start()
	intCmd.Start()

	if err := solCmd.Wait(); err != nil {
		return "RE", fmt.Errorf("%s", solErr)
	}

	if err := intCmd.Wait(); err != nil {
		return "WA", fmt.Errorf("%s", intErr)
	}

	return "OK", nil
}
