package cmd

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	reference, generator, checker, interactor string
	tests, seed                               string
	numTests, numWorkers                      int
	full, unordered, reload, cleanup          bool
	timeLimit                                 time.Duration
	verbose                                   bool
	// memoryLimit megabytes?
	// sandbox bool
)

type TestCase struct {
	number int
	path   string
	seed   int
}

type Result struct {
	solution string
	output   string
	answer   string
	test     TestCase

	message  string
	verdict  string
	duration time.Duration
}

func (result Result) String() string {
	verdictColor := color.FgRed
	if result.verdict == "OK" {
		verdictColor = color.FgGreen
	}
	colorize := color.New(verdictColor).SprintFunc()
	return fmt.Sprintf("%3d  %s  %4dms  %s", result.test.number, colorize(result.verdict), result.duration.Milliseconds(), result.solution)
}

func combineSources(solutions []string) []string {
	sourceStrings := []string{reference, generator, checker, interactor}
	for _, solution := range solutions {
		sourceStrings = append(sourceStrings, solution)
	}
	return sourceStrings
}

func getSeed(i int) int {
	testSeed := rand.Int()
	if seed == "sequential" {
		testSeed = i
	}
	return testSeed
}

func doTest(solutions []string) error {
	// create runDir
	runDir, err := initRunDir()
	if err != nil {
		return fmt.Errorf("Failed to create tempdir: %s", err)
	}

	fmt.Println("Directory for this run:", runDir)

	if cleanup {
		// todo: delete .jay too if empty
		defer os.RemoveAll(runDir)
	}

	// prepare solutions
	runCmd := make(map[string]string)
	sourceStrings := combineSources(solutions)

	for _, sourceString := range sourceStrings {
		if sourceString != "" {
			runCmd[sourceString], err = prepareExecutable(sourceString, runDir)
			if err != nil {
				fmt.Printf("\nFailed to compile `%s`: %s", sourceString, err)
				return nil
			}
		}
	}

	testCases := make(chan TestCase, numWorkers)

	go func() {
		if globalSeed, err := strconv.Atoi(seed); err == nil {
			rand.Seed(int64(globalSeed))
		} else if seed == "random" {
			rand.Seed(time.Now().UTC().UnixNano())
		}
		if interactor != "" {
			for i := 0; i < numTests; i++ {
				testCases <- TestCase{number: i, seed: getSeed(i)}
			}
		} else {
			testNumber := 1
			if tests != "" {
				files, err := ioutil.ReadDir(tests)
				if err != nil {
					panic(fmt.Sprintf("Failed reading tests directory: %s", err))
				}
				for _, file := range files {
					path := filepath.Join(tests, file.Name())
					if isFile(path) && filepath.Ext(path) != ".a" {
						testCases <- TestCase{number: testNumber, path: path}
						testNumber++
					}
				}
			}
			if generator != "" {
				for i := 0; i < numTests; i++ {
					path := runDir + "/tests/" + strconv.Itoa(testNumber) + ".in"
					_, err := bash("SEED="+strconv.Itoa(getSeed(i)), runCmd[generator], ">", path)
					if err != nil {
						panic(fmt.Sprintf("Generator failed: %s", err))
					}
					testCases <- TestCase{number: testNumber, path: path}
					testNumber++
				}
			}
		}
		close(testCases)
	}()

	results := make(chan Result)
	var wg sync.WaitGroup

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			for test := range testCases {
				answer := test.path + ".a"
				if reference != "" && !isFile(answer) {
					if reference == "" {
						panic("No reference or answer file provided for test " + test.path)
					}
					_, err := bash(runCmd[reference], "<", test.path, ">", answer)
					if err != nil {
						panic(fmt.Sprintf("Reference solution failed on test %s: %s", test.path, err))
					}
				}
				for i, solution := range solutions {
					output := fmt.Sprintf("%s/outputs/%d-%d.out", runDir, test.number, i)

					var verdict string

					start := time.Now()
					if interactor != "" {
						file, _ := os.Create(output)
						verdict, err = testInteractive(runCmd[solution], runCmd[interactor], file, test.seed)
					} else {
						_, err = bash("timeout", fmt.Sprintf("%.3fs", timeLimit.Seconds()), runCmd[solution], "<", test.path, ">", output)
					}

					result := Result{
						solution: solution,
						test:     test,
						output:   output,
						answer:   answer,
						duration: time.Since(start),
					}

					if verdict != "" {
						result.verdict = verdict
					}

					if result.duration > timeLimit {
						result.verdict = "TL"
						result.message = fmt.Sprintf("Timed out %dms / %dms", result.duration.Milliseconds(), timeLimit.Milliseconds())
						results <- result
						continue
					}

					if err != nil {
						result.verdict = "RE"
						result.message = err.Error()
						results <- result
						continue
					}

					if interactor != "" {
						result.verdict = "OK"
					} else {
						if checker != "" {
							_, err = bash(runCmd[checker], test.path, output, answer)
							if err != nil {
								result.message = err.Error()
							}
						} else {
							result.message, err = compare(output, answer)
						}

						if err == nil {
							result.verdict = "OK"
						} else {
							result.verdict = "WA"
						}
					}

					results <- result
				}
			}
			wg.Done()
		}()
	}

	// close channel after all workers are done
	go func() {
		wg.Wait()
		close(results)
	}()

	orderedResults := results

	if !unordered {
		nextResult := 1
		buffered := make(map[int]Result)

		orderedResults = make(chan Result)

		go func() {
			for result := range results {
				buffered[result.test.number] = result
				for {
					if _, ok := buffered[nextResult]; ok {
						orderedResults <- buffered[nextResult]
						nextResult++
					} else {
						break
					}
				}
			}
			close(orderedResults)
		}()
	}

	// todo: ordered output
	fmt.Println("\nResults:")
	for result := range orderedResults {
		fmt.Println(result)
		if result.verdict != "OK" && !full {
			fmt.Printf("\n--- Test %d "+strings.Repeat("-", 50)+"\n\n", result.test.number)
			showHead("Input", result.test.path)
			switch result.verdict {
			case "WA":
				if checker != "" {
					showHead("Output", result.output)
					fmt.Println("Checker output:")
					fmt.Println(result.message)
				} else {
					fmt.Println("Output:                 Reference:")
					fmt.Printf(result.message)
				}
			default:
				showHead("Output", result.output)
				fmt.Println("Error:", result.message)
			}
			return nil
		}
	}

	fmt.Println("\nDone")

	return nil
}

// calculateChecksums returns a map from source to its checksum
func calculateChecksums(sourceFiles []string) (checksums []string) {
	for _, sourceFile := range sourceFiles {
		sum, err := bash("md5sum", sourceFile)
		if err != nil {
			panic(fmt.Sprintf("Failed to calculate checksum on file: %s", sourceFile))
		}
		checksums = append(checksums, sum)
	}
	return
}

var testCmd = &cobra.Command{
	Use:   "test [solution]",
	Short: "Tests a solution",
	Args:  cobra.ExactArgs(1), //cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, solutions []string) error {
		sourceStrings := combineSources(solutions)
		var sourceFiles []string
		for _, sourceString := range sourceStrings {
			if isFile(sourceString) {
				sourceFiles = append(sourceFiles, sourceString)
			}
		}
		checksums := calculateChecksums(sourceFiles)
		for {
			err := doTest(solutions)
			if err != nil {
				return err
			}
			if reload {
			reloader:
				for {
					newChecksums := calculateChecksums(sourceFiles)
					for i := range sourceFiles {
						if checksums[i] != newChecksums[i] {
							fmt.Printf("\n--- Source change detected: %s\n\n", sourceFiles[i])
							checksums = newChecksums
							break reloader
						}
					}
					time.Sleep(200 * time.Millisecond)
				}
				continue
			}
			break
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(testCmd)

	testCmd.Flags().StringVarP(&reference, "reference", "r", "", "Refernce solution used for answer file generation")
	testCmd.Flags().StringVarP(&generator, "generator", "g", "", "Test generator, any program that writes test case to stdout")
	testCmd.Flags().StringVarP(&checker, "checker", "c", "",
		"Checker, any program that takes input, output and answer files as argv and returns non-zero exit code for WA")
	testCmd.Flags().StringVarP(&interactor, "interactor", "i", "", "Interactor, returns non-zero exit code for WA")

	testCmd.Flags().StringVarP(&tests, "tests", "t", "",
		"Glob or path to tests. Files ending with '.a' are considered answers. Could be used together with generator and/or reference")
	testCmd.Flags().StringVar(&seed, "seed", "sequential", "Seed policy for generator ('sequential', 'random' or integer)")

	testCmd.Flags().IntVarP(&numTests, "num-tests", "n", 20, "Number of test cases to generate")
	testCmd.Flags().IntVarP(&numWorkers, "num-workers", "w", 1, "Number of parallel threads to use")

	testCmd.Flags().BoolVar(&full, "full", false, "Don't quit after first failed test")
	testCmd.Flags().BoolVar(&unordered, "unordered", false, "Report test verdicts as soon as they come, not necessarily in sequential order")
	testCmd.Flags().BoolVar(&reload, "reload", false, "Rerun tests on source changes")
	testCmd.Flags().BoolVar(&cleanup, "cleanup", false, "Clear run directory when finished")

	testCmd.Flags().DurationVar(&timeLimit, "time-limit", 2*time.Second, "Time limit")

	testCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Print what's happening in the shell (for debug purposes)")
}
