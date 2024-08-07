package reportwriter

import (
	"fmt"

	"github.com/aprimetechnology/derisk-sql/pkg/types"
)

func GetLogMessage(level string, migration string, lineNumber int, charPosition int, code string, message string) string {
	text := ""
	if level != "" {
		text += fmt.Sprintf("[%s]: ", level)
	}
	if migration != "" {
		text += fmt.Sprintf("%s:", migration)
		if lineNumber > 0 && charPosition >= 0 {
			// special case: if we're pointing at the newline right before the start of the line
			// instead point at the following character, ie the 1st non-whitespace character on that line
			if charPosition == 0 {
				charPosition = 1
			}
			text += fmt.Sprintf("%d:%d:", lineNumber, charPosition)
		}
		text += " "
	}
	if code != "" {
		text += fmt.Sprintf("(%s) ", code)
	}
	if message != "" {
		text += message
	}
	text += "\n"
	return text
}

func GetReportString(analyzer string, reports []types.Report, verbose bool) string {
	reportStr := ""
	for _, report := range reports {
		if verbose {
			reportStr += GetLogMessage(analyzer, report.Migration.FileName, -1, -1, "", report.Text)
		}
		for _, diag := range report.Diagnostics {
			reportStr += GetLogMessage(diag.Level, report.Migration.RelativeFilePath, diag.LineNumber, diag.LinePosition, diag.Code, diag.Text)
		}
		if verbose {
			// add a newline after each massive report block
			reportStr += "\n"
		}
	}
	return reportStr
}

func GetReportFatalityStatus(reports []types.Report) bool {
	for _, report := range reports {
		for _, diag := range report.Diagnostics {
			if diag.Level == types.DiagnosticLevelFatal {
				return true
			}
		}
	}
	return false
}

func WriteReportsToStdout(analyzer string, reports []types.Report, verbose bool) {
	// output the final report string to standard out
	reportString := GetReportString(analyzer, reports, verbose)

	// skip outputting anything for this analyzer if it produced no reports
	if reportString == "" {
		return
	}

	fmt.Print(reportString)
}
