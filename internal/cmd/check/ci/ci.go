package ci

import (
	"fmt"
	"strings"

	"github.com/aprimetechnology/derisk-sql/internal/reportwriter"
	github "github.com/aprimetechnology/derisk-sql/pkg/actions/github/client"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	flagInput       = "input-dir"
	defaultInputDir = "reports"
)

type ciCheckFlags struct {
	InputDir string
	github   github.GithubClient
}

var (
	flags      = ciCheckFlags{}
	CiCheckCmd = &cobra.Command{
		Use:   "ci",
		Short: "report checks in CI",
		Long:  `Reports and acts on pull requests from sql check output`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if viper.ConfigFileUsed() != "" {
				// by default viper.Unmarshal unwisely merges config file contents
				// on top of existing flag default values. eg:
				//  - defaultAnalyzers = ["a1", "a2", "a3"]
				//  - configAnalyzers = ["c1"]
				// flags.Analyzers would end up, confusingly, being ["c1", "a2", "a3"]

				// hence we wipe defaults here by instantiating a blank flagsobject
				configFlags := ciCheckFlags{}
				if err := viper.Unmarshal(&configFlags); err != nil {
					return fmt.Errorf(
						"Failure to unmarshal config file via viper: %w",
						err,
					)
				}

				// copy over all flags from the config file that are set and have a value
				if configFlags.InputDir != "" {
					flags.InputDir = configFlags.InputDir
				}

			}
			return ciCheckRun(cmd, args, flags)
		},
	}
)

func init() {
	CiCheckCmd.Flags().StringVar(
		&flags.InputDir,
		flagInput,
		defaultInputDir,
		"Directory to read check results from",
	)
}

func ciCheckRun(cmd *cobra.Command, args []string, flags ciCheckFlags) error {
	reports, err := reportwriter.GetAllReportsFromJsonFileDirectory(flags.InputDir)
	if err != nil {
		return err
	}

	// always write the report output as a comment to the pull request
	flags.github = github.New()
	err = reportwriter.WriteReportsToPullRequest(reports, flags.github)
	if err != nil {
		return err
	}

	// execute all requested actions
	for _, report := range reports {
		for _, action := range report.Actions {
			if action == github.RequestReviewersAction {
				reviewers, ok := report.Config[github.GithubReviewersKey]
				if !ok {
					fmt.Printf("Requested action %q but missing any values for config key %q\n", action, github.GithubReviewersKey)
					continue
				}
				reviewersSlice := strings.Split(reviewers, ",")
				if err = flags.github.RequestReviewers(reviewersSlice); err != nil {
					return err
				}
				fmt.Printf("Successfully added reviewers %q to the pull request\n", reviewers)
			}
		}
	}
	return nil
}
