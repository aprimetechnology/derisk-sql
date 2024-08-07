package main

import (
	"context"
	"fmt"

	"github.com/aprimetechnology/derisk-sql/pkg/analysis"
	"github.com/aprimetechnology/derisk-sql/pkg/pgquery"
	"github.com/aprimetechnology/derisk-sql/pkg/subprocess"
	"github.com/aprimetechnology/derisk-sql/pkg/types"

	pg_query "github.com/pganalyze/pg_query_go/v5"
)

const DiagnosticCode = "IND-002"

type DropIndexConcurrentlyAnalyzer struct{}

func (a *DropIndexConcurrentlyAnalyzer) Analyze(ctx context.Context, migration string, options map[string]string) []types.Diagnostic {
	parseTree, err := pg_query.Parse(migration)
	if err != nil {
		return []types.Diagnostic{types.Diagnostic{
			LineNumber:   -1,
			LinePosition: -1,
			Code:         DiagnosticCode,
			Level:        types.DiagnosticLevelFatal,
			Text:         fmt.Errorf("error parsing migration: `%s`: %w", migration, err).Error(),
		}}
	}

	diagnostics := []types.Diagnostic{}
	for _, statement := range parseTree.Stmts {
		if drop := pgquery.GetDropIndexStatement(statement); drop != nil && !drop.Concurrent {
			byteOffset := pgquery.SkipWhitespaceAndComments(migration, int(statement.StmtLocation))
			textLocation := pgquery.GetTextLocation(migration, byteOffset)
			diagnostics = append(diagnostics, types.Diagnostic{
				LineNumber:   textLocation.LineNumber,
				LinePosition: textLocation.LineCharPosition,
				Code:         DiagnosticCode,
				Level:        types.DiagnosticLevelWarning,
				Text:         "DROP INDEX statement missing CONCURRENTLY option",
			})
		}
	}
	return diagnostics
}

func main() {
	// standard input expected to have JSON containing:
	// - a list of migration objects
	// - an overall metadata object
	input := subprocess.Input()

	// analyze the input. ie, ensure that for every migration:
	// - any DROP INDEX operation
	// - has a CONCURRENTLY keyword attached to it
	output := analysis.DoSimpleAnalysis(
		input,
		&DropIndexConcurrentlyAnalyzer{},
		"Errors occurred around DROP INDEX statement(s) with missing CONCURRENTLY option",
		[]string{},
	)

	// standard output expected to print JSON containing:
	// - a list of report objects
	subprocess.Output(output)
}