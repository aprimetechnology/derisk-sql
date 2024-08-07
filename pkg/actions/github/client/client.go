package client

import (
	"context"
	"fmt"
	"os"
	"strconv"

	gh "github.com/google/go-github/v62/github"
)

const (
	EnvVarGithubToken             = "GITHUB_TOKEN"
	EnvVarGithubRepoName          = "GITHUB_REPOSITORY_NAME"
	EnvVarGithubRepoOwner         = "GITHUB_REPOSITORY_OWNER"
	EnvVarGithubPullRequestNumber = "GITHUB_PULL_REQUEST_NUMBER"
	GithubReviewersKey            = "github_reviewers"
	RequestReviewersAction        = "github:requestReviewers"
)

type GithubClient struct {
	Client            *gh.Client
	Token             string
	RepoName          string
	RepoOwner         string
	PullRequestNumber int
}

func New() GithubClient {
	missingEnvVars := []string{}
	for _, envVar := range []string{
		EnvVarGithubToken,
		EnvVarGithubRepoName,
		EnvVarGithubRepoOwner,
		EnvVarGithubPullRequestNumber,
	} {
		if os.Getenv(envVar) == "" {
			missingEnvVars = append(missingEnvVars, envVar)
		}
	}
	if len(missingEnvVars) > 0 {
		panic(fmt.Errorf("Missing the following required github-related environment variables: %v", missingEnvVars))
	}
	prNumber := os.Getenv(EnvVarGithubPullRequestNumber)
	prNumberInt, err := strconv.Atoi(prNumber)
	if err != nil {
		panic(fmt.Errorf("Missing a valid github pull request number. %q is not a valid number: %w", prNumber, err))
	}
	return GithubClient{
		Client:            gh.NewClient(nil).WithAuthToken(os.Getenv(EnvVarGithubToken)),
		Token:             os.Getenv(EnvVarGithubToken),
		RepoName:          os.Getenv(EnvVarGithubRepoName),
		RepoOwner:         os.Getenv(EnvVarGithubRepoOwner),
		PullRequestNumber: prNumberInt,
	}
}

func (ghc *GithubClient) CreateComment(body string) error {
	_, _, err := ghc.Client.Issues.CreateComment(
		context.Background(),
		ghc.RepoOwner,
		ghc.RepoName,
		ghc.PullRequestNumber,
		&gh.IssueComment{
			Body: &body,
		},
	)
	if err != nil {
		return fmt.Errorf(
			"Failure posting comment to pull request #%d on repo %s/%s: %w",
			ghc.PullRequestNumber,
			ghc.RepoOwner,
			ghc.RepoName,
			err,
		)
	}
	return nil
}

func (ghc *GithubClient) BuildReviewersRequest(reviewers []string) (gh.ReviewersRequest, error) {
	var individualReviewers, teamReviewers []string
	teams, _, err := ghc.Client.Repositories.ListTeams(
		context.Background(),
		ghc.RepoOwner,
		ghc.RepoName,
		&gh.ListOptions{},
	)
	if err != nil {
		return gh.ReviewersRequest{}, fmt.Errorf("Unable to lookup provided reviewer names to see which are teams: %w", err)
	}

	// NOTE: slugs are used exclusively in API calls, not the team.Name
	// the 'slug' is the URL friendly form of the team 'Name'
	// eg: Name of "Example Name" would have a slug like "example-name"
	teamSlugs := map[string]bool{}
	for _, team := range teams {
		if team.Slug != nil {
			teamSlugs[*team.Slug] = true
		}
	}

	for _, reviewer := range reviewers {
		if teamSlugs[reviewer] {
			teamReviewers = append(teamReviewers, reviewer)
		} else {
			individualReviewers = append(individualReviewers, reviewer)
		}
	}

	return gh.ReviewersRequest{
		Reviewers:     individualReviewers,
		TeamReviewers: teamReviewers,
	}, nil
}

func (ghc *GithubClient) RequestReviewers(reviewers []string) error {
	reviewersRequest, err := ghc.BuildReviewersRequest(reviewers)
	if err != nil {
		return err
	}
	_, _, err = ghc.Client.PullRequests.RequestReviewers(
		context.Background(),
		ghc.RepoOwner,
		ghc.RepoName,
		ghc.PullRequestNumber,
		reviewersRequest,
	)
	return err
}
