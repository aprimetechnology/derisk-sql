package check

import (
	"github.com/aprimetechnology/derisk-sql/internal/cmd/check/ci"
	"github.com/aprimetechnology/derisk-sql/internal/cmd/check/run"
	"github.com/spf13/cobra"
)

var CheckCmd = &cobra.Command{
	Use:          "check",
	SilenceUsage: true,
	Short:        "Check SQL migration files for any linting rule violations",
}

func init() {
	// `run`:
	//      - executes all the actual analyzers against the migration files
	//      - prints reports to stdout
	//      - optionally writes the reports to JSON files on disk
	CheckCmd.AddCommand(run.RunCheckCmd)

	// `ci`:
	//      - reads JSON report files from disk
	//      - performs any requested actions, eg:
	//          - comment on PRs
	//          - assign PR reviewers
	//          - etc
	CheckCmd.AddCommand(ci.CiCheckCmd)
}
