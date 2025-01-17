package pkg

import (
	"context"
	"fmt"

	"github.com/Azure/golden"
	"github.com/Azure/grept/pkg/githubclient"
	"github.com/google/go-github/v61/github"
)

var _ Data = &GitHubRepositoryEnvironmentsDatasource{}

type EnvironmentForGitHubRepositoryEnvironmentsDatasource struct {
	Name   string `attribute:"name"`
	NodeId string `attribute:"node_id"`
}

type GitHubRepositoryEnvironmentsDatasource struct {
	*golden.BaseBlock
	*BaseData
	Owner        string                                                 `hcl:"owner"`
	RepoName     string                                                 `hcl:"repo_name"`
	Environments []EnvironmentForGitHubRepositoryEnvironmentsDatasource `attribute:"environments"`
}

func (g *GitHubRepositoryEnvironmentsDatasource) Type() string {
	return "github_repository_environments"
}

func (g *GitHubRepositoryEnvironmentsDatasource) ExecuteDuringPlan() error {
	client, err := githubclient.GetGithubClient()
	if err != nil {
		return fmt.Errorf("cannot create github client: %s", err.Error())
	}
	results, err := listGitHubRepositoryEnvironments(client, g.Owner, g.RepoName)
	if err != nil {
		return fmt.Errorf("cannot list environments for %s/%s: %+v", g.Owner, g.RepoName, err)
	}
	g.Environments = results
	return nil
}

func listGitHubRepositoryEnvironments(client *githubclient.Client, owner, repoName string) ([]EnvironmentForGitHubRepositoryEnvironmentsDatasource, error) {
	var results []EnvironmentForGitHubRepositoryEnvironmentsDatasource
	var listOptions *github.EnvironmentListOptions
	for {
		environments, resp, err := client.Repositories.ListEnvironments(context.Background(), owner, repoName, listOptions)
		if err != nil {
			return nil, err
		}
		if environments == nil {
			return nil, nil
		}
		for _, environment := range environments.Environments {
			results = append(results, EnvironmentForGitHubRepositoryEnvironmentsDatasource{
				Name:   environment.GetName(),
				NodeId: environment.GetNodeID(),
			})
		}
		if resp.NextPage == 0 {
			break
		}
		listOptions.Page = resp.NextPage
	}
	return results, nil
}
