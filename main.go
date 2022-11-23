package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/cli/go-gh"
	"github.com/spf13/cobra"
)

func main() {
	var Algorithm string
	var globs []string
	var concurrency, index int

	cmd := &cobra.Command{
		Use:   "split-tests",
		Short: "Utility for splitting tests suites across multiple jobs",
		Long:  "Split tests suites using globbing patterns and concurrency types to execute tests in parallel",
		RunE: func(cmd *cobra.Command, args []string) error {
			var err error
			if concurrency == -1 {
				value, ok := os.LookupEnv("TOTAL_CONCURRENCY")
				if !ok {
					return fmt.Errorf("failed to retrieve concurreny, set TOTAL_CONCURRENCY environment variable or set the --concurrency flag")
				}

				concurrency, err = strconv.Atoi(value)
				if err != nil {
					return fmt.Errorf("failed to parse TOTAL_CONCURRENCY: %w", err)
				}
			}

			if index == -1 {
				value, ok := os.LookupEnv("JOB_INDEX")
				if !ok {
					return fmt.Errorf("failed to retrieve concurreny, set JOB_INDEX environment variable or set the --index flag")
				}

				index, err = strconv.Atoi(value)
				if err != nil {
					return fmt.Errorf("failed to parse JOB_INDEX: %w", err)
				}
			}

			switch Algorithm {
			case "name":
				files, err := splitByFileName(concurrency, index, args, globs...)
				if err != nil {
					return fmt.Errorf("failed to split tests by file name: %w", err)
				}
				if len(files) == 0 {
					return fmt.Errorf("no test files found")
				}

				fmt.Println(strings.Join(files, " "))
				return nil
			case "time":
				files, err := splitByTestTime(concurrency, index, args, globs...)
				if err != nil {
					return fmt.Errorf("failed to split tests by test time: %w", err)
				}
				if len(files) == 0 {
					return fmt.Errorf("no test files found")
				}

				fmt.Println(strings.Join(files, " "))
				return nil
			default:
				return fmt.Errorf(`invalid algorithm "%s"", must be one of: name or time`, Algorithm)
			}
		},
	}

	cmd.Flags().StringVarP(&Algorithm, "algorithm", "a", "name", "Algorithm to use for splitting tests: name or time")
	cmd.Flags().StringSliceVarP(&globs, "globs", "g", []string{}, "Globs to match against test files")
	cmd.Flags().IntVarP(&concurrency, "concurrency", "c", -1, "Number of jobs to split tests across")
	cmd.Flags().IntVarP(&index, "index", "i", -1, "Index of the current job")
	cmd.MarkFlagsRequiredTogether("concurrency", "index")

	err := cmd.Execute()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(500)
	}
}

func splitByFileName(concurrency, index int, files []string, globs ...string) ([]string, error) {
	var matches []string

	for i, file := range files {
		if i%concurrency == index {
			if len(globs) > 0 {
				for _, glob := range globs {
					match, err := filepath.Match(glob, file)
					if err != nil {
						return nil, fmt.Errorf("error matching glob %s: %w", glob, err)
					}
					if match {
						matches = append(matches, file)
					}
				}
			} else {
				matches = append(matches, file)
			}
		}
	}

	return matches, nil
}

func splitByTestTime(concurrency, index int, files []string, globs ...string) ([]string, error) {

	return nil, nil
}

type WorkflowRunNotFound struct{}

func (m *WorkflowRunNotFound) Error() string {
	return "Workflow run not found"
}

func findLastWorkflowRun() (string, error) {
	client, err := gh.RESTClient(nil)
	if err != nil {
		return "", fmt.Errorf("failed to create REST client: %w", err)
	}
	response := struct {
		WorkflowRuns []struct {
			ArtifactsURL string `json:"artifacts_url"`
		} `json:"workflow_runs"`
	}{}
	fullName := os.Getenv("GITHUB_REPOSITORY")
	workflow := os.Getenv("GITHUB_WORKFLOW")
	path := fmt.Sprintf("/repos/%s/actions/workflows/%s/runs?status=success&per_page=1", fullName, workflow)
	err = client.Get(path, &response)
	if err != nil {
		return "", fmt.Errorf("failed to retrieve last run: %w", err)
	}

	if len(response.WorkflowRuns) == 0 {
		return "", &WorkflowRunNotFound{}
	}

	return response.WorkflowRuns[0].ArtifactsURL, nil
}

//func downloadFile() error {
//	client, err := gh.RESTClient(nil)
//	if err != nil {
//		return fmt.Errorf("failed to create REST client: %w", err)
//	}
//	response := struct {
//		Artifacts []struct {
//			Name string `json:"name"`
//			URL  string `json:"archive_download_url"`
//		} `json:"artifacts"`
//	}{}
//	artifactsURL, err := findLastWorkflowRun()
//	if err != nil {
//		return fmt.Errorf("failed to find last workflow run: %w", err)
//	}
//	err = client.Get(artifactsURL, &response)
//	if err != nil {
//		return fmt.Errorf("failed to retrieve artifacts: %w", err)
//	}
//
//	if len(response.Artifacts) == 0 {
//		return fmt.Errorf("no artifacts found")
//	}
//
//	artifact := response.Artifacts[0]
//	err = client.Download(artifact.URL, "artifact.zip")
//	if err != nil {
//		return fmt.Errorf("failed to download artifact: %w", err)
//	}
//
//	return nil
//}
