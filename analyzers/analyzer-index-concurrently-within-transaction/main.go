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

const DiagnosticCode = "IND-003"

type CreateDropIndexConcurrentlyWithinTransactionAnalyzer struct{}

func (a *CreateDropIndexConcurrentlyWithinTransactionAnalyzer) Analyze(ctx context.Context, migration string, options map[string]string) []types.Diagnostic {
	// dbmate sets the "transaction" option to true by default
	// meaning every migration is run in a transaction block by default

	// this can be overridden by setting the option "transaction:false"
	// in which case we can't possibly have an issue around a
	// DROP INDEX/CREATE INDEX happening inside a transaction block
	if options["transaction"] == "false" {
		return nil
	}

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
		if create := pgquery.GetCreateIndexStatement(statement); create != nil && create.Concurrent {
			byteOffset := pgquery.SkipWhitespaceAndComments(migration, int(statement.StmtLocation))
			textLocation := pgquery.GetTextLocation(migration, byteOffset)
			diagnostics = append(diagnostics, types.Diagnostic{
				LineNumber:   textLocation.LineNumber,
				LinePosition: textLocation.LineCharPosition,
				Code:         DiagnosticCode,
				Level:        types.DiagnosticLevelFatal,
				Text:         "CREATE INDEX CONCURRENTLY statement is happening within a transaction block! This is prohibited",
			})
		}
		if drop := pgquery.GetDropIndexStatement(statement); drop != nil && drop.Concurrent {
			byteOffset := pgquery.SkipWhitespaceAndComments(migration, int(statement.StmtLocation))
			textLocation := pgquery.GetTextLocation(migration, byteOffset)
			diagnostics = append(diagnostics, types.Diagnostic{
				LineNumber:   textLocation.LineNumber,
				LinePosition: textLocation.LineCharPosition,
				Code:         DiagnosticCode,
				Level:        types.DiagnosticLevelFatal,
				Text:         "DROP INDEX CONCURRENTLY statement is happening within a transaction block! This is prohibited",
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
	// - any CREATE INDEX operation or DROP INDEX operation
	// - that has a CONCURRENTLY keyword attached to it
	// - is not being performed inside a TRANSACTION block (this is illegal)
	output := analysis.DoSimpleAnalysis(
		input,
		&CreateDropIndexConcurrentlyWithinTransactionAnalyzer{},
		"Errors occurred around CREATE INDEX or DROP INDEX statement with CONCURRENTLY option happening inside a transaction block",
		[]string{},
	)

	// standard output expected to print JSON containing:
	// - a list of report objects
	subprocess.Output(output)
}
