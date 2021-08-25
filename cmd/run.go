package cmd

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/flanksource/commons/console"

	"github.com/spf13/cobra"

	"github.com/flanksource/canary-checker/api/context"
	"github.com/flanksource/canary-checker/checks"
	"github.com/flanksource/canary-checker/pkg"
	"github.com/flanksource/commons/logger"
)

var Run = &cobra.Command{
	Use:   "run",
	Short: "Execute checks and return",
	Run: func(cmd *cobra.Command, args []string) {
		configfile, _ := cmd.Flags().GetString("configfile")
		namespace, _ := cmd.Flags().GetString("namespace")
		junitFile, _ := cmd.Flags().GetString("junit-file")
		config, err := pkg.ParseConfig(configfile)
		if err != nil {
			logger.Fatalf("Could not parse %s: %v", configfile, err)
		}
		if namespace != "" {
			config.Namespace = namespace
		}
		if config.Name == "" {
			config.Name = CleanupFilename(configfile)
		}
		failed := 0
		kommonsClient, err := pkg.NewKommonsClient()
		if err != nil {
			logger.Warnf("Failed to get kommons client, features that read kubernetes configs will fail: %v", err)
		}

		results := checks.RunChecks(context.New(kommonsClient, *config))
		if junitFile != "" {
			report := getJunitReport(results)
			err := ioutil.WriteFile(junitFile, []byte(report), 0755)
			if err != nil {
				logger.Fatalf("%d checks failed", failed)
			}
		}
		for _, result := range results {
			fmt.Println(result)
			if !result.Pass {
				failed++
			}
		}
		if failed > 0 {
			logger.Fatalf("%d checks failed", failed)
		}
	},
}

func init() {
	Run.Flags().StringP("configfile", "c", "", "Specify configfile")
	Run.Flags().StringP("namespace", "n", "", "Specify namespace")
	Run.Flags().StringP("junit", "j", "", "Export JUnit XML formatted results to this file e.g: junit.xml")
}

func getJunitReport(results []*pkg.CheckResult) string {
	var testCases []console.JUnitTestCase
	var failed int
	var totalTime int64
	for _, result := range results {
		totalTime += result.Duration
		testCase := console.JUnitTestCase{
			Classname: result.Check.GetType(),
			Name:      result.Check.GetDescription(),
			Time:      strconv.Itoa(int(result.Duration)),
		}
		if !result.Pass {
			failed++
			testCase.Failure = &console.JUnitFailure{
				Message: result.Message,
			}
		}
		testCases = append(testCases, testCase)
	}
	testSuite := console.JUnitTestSuite{
		Tests:     len(results),
		Failures:  failed,
		Time:      strconv.Itoa(int(totalTime)),
		Name:      "canary-checker-run",
		TestCases: testCases,
	}
	testSuites := console.JUnitTestSuites{
		Suites: []console.JUnitTestSuite{
			testSuite,
		},
	}
	report, err := testSuites.ToXML()
	if err != nil {
		logger.Fatalf("error creating junit results: %v", err)
	}
	return report
}

func CleanupFilename(fileName string) string {
	removeSuffix := fileName[:len(fileName)-len(filepath.Ext(fileName))]
	return strings.Replace(removeSuffix, "_", "", -1)
}
