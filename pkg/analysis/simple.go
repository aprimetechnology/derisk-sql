package analysis

import (
	"context"

	"github.com/aprimetechnology/derisk-sql/pkg/types"
)

const ConfigKey = "config"

type SimpleOneMigrationAnalyzer interface {
	Analyze(ctx context.Context, migration string, options map[string]string) []types.Diagnostic
}

func GetConfigValue(ctx context.Context, key string) (string, bool) {
	config, ok := ctx.Value(ConfigKey).(map[string]string)
	if !ok {
		return "", false
	}
	value, ok := config[key]
	return value, ok
}

func PadDownMigration(up string, down string) string {
	padding := ""
	for i, char := range up {
		if i == len(up)-2 {
			// have the second to last character be a SQL statement terminator
			// so that all this leading whitespace is treated as its own *empty* statement
			padding += ";"
		} else if i == len(up)-1 {
			// have the very last character be a new line so that the down migration
			// begins on its very own line
			padding += "\n"
		} else if char == '\n' {
			padding += string(char)
		} else {
			padding += " "
		}
	}
	return padding + "\n" + down
}

// Runs a typical very simple analysis:
// - run a given SimpleAnalyzer for each migration's Up contents
// - store any resulting diagnostic
// - run a given SimpleAnalyzer for each migration's Down contents
// - store any resulting diagnostic
// - if any diagnostic is produced, store a Report for that migration file
// - eventually, return all Reports generated
func DoSimpleAnalysis(
	input types.ParsedMigrationsSummary,
	simpleAnalyzer SimpleOneMigrationAnalyzer,
	reportText string,
	actions []string,
) types.AnalyzedMigrationsSummary {
	// for this convenience "simple" package we pass a context object
	// to maintain a consistent interface even as the contents of the
	// context object may be modified over time, the Analyze() interface
	// will take the same things as it did before
	ctx := context.Background()
	ctx = context.WithValue(ctx, ConfigKey, input.Metadata.Config)

	reports := []types.Report{}
	for _, migration := range input.Migrations {
		diagnostics := []types.Diagnostic{}

		upDiagnostics := simpleAnalyzer.Analyze(ctx, migration.Up, migration.UpOptions)
		if len(upDiagnostics) != 0 {
			for _, diagnostic := range upDiagnostics {
				diagnostics = append(diagnostics, diagnostic)
			}
		}

		// to ensure that Diagnostic.LineNumber and Diagnostic.LinePosition are correct
		// we pad the down migration with the contents of the up migration,
		// where every non-'\n' character is replaced with a space ' ' character
		paddedDown := PadDownMigration(migration.Up, migration.Down)
		downDiagnostics := simpleAnalyzer.Analyze(ctx, paddedDown, migration.DownOptions)
		if len(downDiagnostics) != 0 {
			for _, diagnostic := range downDiagnostics {
				diagnostics = append(diagnostics, diagnostic)
			}
		}

		if len(diagnostics) == 0 {
			continue
		}
		reports = append(reports, types.Report{
			Migration:   migration,
			Text:        reportText,
			Diagnostics: diagnostics,
			Actions:     actions,
			Config:      input.Metadata.Config,
		})
	}
	return types.AnalyzedMigrationsSummary{Reports: reports}
}
