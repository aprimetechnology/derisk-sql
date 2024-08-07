package reportwriter

import (
	"fmt"
	"strings"

	github "github.com/aprimetechnology/derisk-sql/pkg/actions/github/client"
	"github.com/aprimetechnology/derisk-sql/pkg/types"
)

func GetPullRequestCommentString(reports []types.Report) string {
	reportStr := ""
	for _, report := range reports {
		for _, diag := range report.Diagnostics {
			reportStr += GetLogMessage(diag.Level, report.Migration.RelativeFilePath, diag.LineNumber, diag.LinePosition, diag.Code, diag.Text)
		}
	}
	return reportStr
}

func WriteReportsToPullRequest(reports []types.Report, githubClient github.GithubClient) error {
	commentString := GetPullRequestCommentString(reports)
	if commentString == "" {
		commentString = "CI check passed successfully! No errors or warnings"
	} else {
		commentString = fmt.Sprintf(
			"CI check encountered the following warnings and errors:\n```\n%s\n```\n",
			strings.TrimRight(commentString, " \n\t"),
		)
	}
	fmt.Printf("Posting the following comment to the pull request: '''\n%s\n'''\n", commentString)

	return githubClient.CreateComment(commentString)
}
