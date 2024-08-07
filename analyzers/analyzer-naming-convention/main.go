package main

import (
	"context"
	"fmt"
	"regexp"

	"github.com/aprimetechnology/derisk-sql/pkg/analysis"
	"github.com/aprimetechnology/derisk-sql/pkg/pgquery"
	"github.com/aprimetechnology/derisk-sql/pkg/subprocess"
	"github.com/aprimetechnology/derisk-sql/pkg/types"

	pg_query "github.com/pganalyze/pg_query_go/v5"
)

const (
	NamingRegexKey           = "naming_regex"
	DefaultRegex             = "^[a-zA-Z_]+$"
	DiagnosticCode           = "NMC-000"
	DiagnosticCodeSchemaName = "NMC-001"
	DiagnosticCodeTableName  = "NMC-002"
	DiagnosticCodeIndexName  = "NMC-003"
	DiagnosticCodeColumnName = "NMC-004"
)

type RenameInfo struct {
	ObjectType     string
	DiagnosticCode string
}

var RenameCodeToInfo = map[int]RenameInfo{
	int(pg_query.ObjectType_OBJECT_SCHEMA): RenameInfo{
		ObjectType:     "schema",
		DiagnosticCode: DiagnosticCodeSchemaName,
	},
	int(pg_query.ObjectType_OBJECT_TABLE): RenameInfo{
		ObjectType:     "table",
		DiagnosticCode: DiagnosticCodeTableName,
	},
	int(pg_query.ObjectType_OBJECT_INDEX): RenameInfo{
		ObjectType:     "index",
		DiagnosticCode: DiagnosticCodeIndexName,
	},
	int(pg_query.ObjectType_OBJECT_COLUMN): RenameInfo{
		ObjectType:     "column",
		DiagnosticCode: DiagnosticCodeColumnName,
	},
}

func NewNamingDiagnostic(name string, regex string, location pgquery.TextLocation, objectType int32) types.Diagnostic {
	info, ok := RenameCodeToInfo[int(objectType)]
	if !ok {
		return types.Diagnostic{
			LineNumber:   -1,
			LinePosition: -1,
			Code:         DiagnosticCode,
			Level:        types.DiagnosticLevelFatal,
			Text:         fmt.Sprintf("FAILURE to validate naming for unsupported object type code %d", int(objectType)),
		}
	}
	return types.Diagnostic{
		LineNumber:   location.LineNumber,
		LinePosition: location.LineCharPosition,
		Code:         info.DiagnosticCode,
		Level:        types.DiagnosticLevelFatal,
		Text:         fmt.Sprintf("New %q name %q does not meet naming requirement regex %q", info.ObjectType, name, regex),
	}
}

type NamingConventionAnalyzer struct{}

func (a *NamingConventionAnalyzer) Analyze(ctx context.Context, migration string, options map[string]string) []types.Diagnostic {
	regexString, ok := analysis.GetConfigValue(ctx, NamingRegexKey)
	if !ok {
		regexString = DefaultRegex
	}
	regex, err := regexp.Compile(regexString)
	if err != nil || regex == nil {
		return []types.Diagnostic{types.Diagnostic{
			LineNumber:   -1,
			LinePosition: -1,
			Code:         DiagnosticCode,
			Level:        types.DiagnosticLevelFatal,
			Text:         fmt.Errorf("error compiling regex string %q: %w", regexString, err).Error(),
		}}
	}

	parseTree, err := pg_query.Parse(migration)
	if err != nil {
		return []types.Diagnostic{types.Diagnostic{
			LineNumber:   -1,
			LinePosition: -1,
			Code:         DiagnosticCode,
			Level:        types.DiagnosticLevelFatal,
			Text:         fmt.Errorf("error compiling regex string %q: %w", regexString, err).Error(),
		}}
	}

	diagnostics := []types.Diagnostic{}
	for _, statement := range parseTree.Stmts {
		byteOffset := pgquery.SkipWhitespaceAndComments(migration, int(statement.StmtLocation))
		textLocation := pgquery.GetTextLocation(migration, byteOffset)
		// create schema -> schema name
		if create := statement.Stmt.GetCreateSchemaStmt(); create != nil {
			name := create.Schemaname
			if !regex.MatchString(name) {
				diagnostics = append(diagnostics, NewNamingDiagnostic(name, regexString, textLocation, int32(pg_query.ObjectType_OBJECT_SCHEMA)))
			}
		}

		// rename (any object) -> new name
		if rename := statement.Stmt.GetRenameStmt(); rename != nil {
			name := rename.Newname
			if !regex.MatchString(name) {
				diagnostics = append(diagnostics, NewNamingDiagnostic(name, regexString, textLocation, int32(rename.RenameType)))
			}
		}

		// create table -> table name, column names
		if create := statement.Stmt.GetCreateStmt(); create != nil {
			name := create.Relation.Relname
			if !regex.MatchString(name) {
				diagnostics = append(diagnostics, NewNamingDiagnostic(name, regexString, textLocation, int32(pg_query.ObjectType_OBJECT_TABLE)))
			}
			for _, col := range create.TableElts {
				if colDef := col.GetColumnDef(); colDef != nil {
					name = colDef.Colname
					if !regex.MatchString(name) {
						diagnostics = append(diagnostics, NewNamingDiagnostic(name, regexString, textLocation, int32(pg_query.ObjectType_OBJECT_COLUMN)))
					}
				}
			}
		}

		// alter table -> add column names
		if alter := statement.Stmt.GetAlterTableStmt(); alter != nil {
			for _, cmd := range alter.Cmds {
				if colDef := cmd.GetAlterTableCmd().GetDef().GetColumnDef(); colDef != nil {
					name := colDef.Colname
					if !regex.MatchString(name) {
						diagnostics = append(diagnostics, NewNamingDiagnostic(name, regexString, textLocation, int32(pg_query.ObjectType_OBJECT_COLUMN)))
					}
				}
			}
		}

		// create index -> index name
		if create := pgquery.GetCreateIndexStatement(statement); create != nil {
			name := create.Idxname
			if !regex.MatchString(name) {
				diagnostics = append(diagnostics, NewNamingDiagnostic(name, regexString, textLocation, int32(pg_query.ObjectType_OBJECT_INDEX)))
			}
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
	// - any naming operation (CREATE, ALTER, RENAME)
	// - meets a provided (or default) regex
	output := analysis.DoSimpleAnalysis(
		input,
		&NamingConventionAnalyzer{},
		"Errors occurred around enforcing naming convention for database objects",
		[]string{},
	)

	// standard output expected to print JSON containing:
	// - a list of report objects
	subprocess.Output(output)
}
