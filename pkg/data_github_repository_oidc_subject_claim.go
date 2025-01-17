package pkg

import (
	"context"
	"fmt"

	"github.com/Azure/golden"
	"github.com/Azure/grept/pkg/githubclient"
)

var _ Data = &GitHubRepositoryOidcSubjectClaimDatasource{}

type GitHubRepositoryOidcSubjectClaimDatasource struct {
	*golden.BaseBlock
	*BaseData
	Owner           string   `hcl:"owner"`
	RepoName        string   `hcl:"repo_name"`
	ClaimKeys       []string `attribute:"claim_keys"`
	ClaimUseDefault *bool    `attribute:"claim_use_default"`
}

func (g *GitHubRepositoryOidcSubjectClaimDatasource) Type() string {
	return "github_repository_oidc_subject_claim"
}

func (g *GitHubRepositoryOidcSubjectClaimDatasource) ExecuteDuringPlan() error {
	client, err := githubclient.GetGithubClient()
	if err != nil {
		return fmt.Errorf("cannot create github client: %s", err.Error())
	}
	claimKeys, claimUseDefault, err := getGitHubRepositoryOidcSubjectClaims(client, g.Owner, g.RepoName)
	if err != nil {
		return fmt.Errorf("cannot list OIDC subject claims for %s/%s: %+v", g.Owner, g.RepoName, err)
	}
	g.ClaimKeys = claimKeys
	g.ClaimUseDefault = claimUseDefault
	return nil
}

func getGitHubRepositoryOidcSubjectClaims(client *githubclient.Client, owner, repoName string) ([]string, *bool, error) {
	template, resp, err := client.Actions.GetRepoOIDCSubjectClaimCustomTemplate(context.Background(), owner, repoName)
	if err != nil {
		return nil, nil, err
	}
	if resp.StatusCode != 200 {
		return nil, nil, nil
	}
	return template.IncludeClaimKeys, template.UseDefault, nil
}
