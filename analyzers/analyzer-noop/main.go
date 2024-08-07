package main

import (
	"github.com/aprimetechnology/derisk-sql/pkg/subprocess"
	"github.com/aprimetechnology/derisk-sql/pkg/types"
)

func analyze(summary types.ParsedMigrationsSummary) types.AnalyzedMigrationsSummary {
	reports := []types.Report{}
	for _, migration := range summary.Migrations {
		reports = append(reports, types.Report{
			Migration: migration,
			Text:      "Noop, nothing to see here",
			Diagnostics: []types.Diagnostic{
				types.Diagnostic{
					LineNumber:   -1,
					LinePosition: -1,
					Text:         "Empty diagnostic",
					Code:         "NOOP",
					Level:        types.DiagnosticLevelWarning,
				},
			},
			Actions: []string{},
		})

	}
	return types.AnalyzedMigrationsSummary{
		Reports: reports,
	}
}
func main() {
	// standard input expected to have JSON containing:
	// - a list of migration objects
	// - an overall metadata object
	input := subprocess.Input()

	// analyze the input. ie, do *something* meaningful here
	// for now, this is just a no-op operation
	output := analyze(input)

	// standard output expected to print JSON containing:
	// - a list of report objects
	subprocess.Output(output)
}
