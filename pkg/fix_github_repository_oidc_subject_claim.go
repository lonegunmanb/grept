package pkg

import (
	"context"
	"fmt"

	"github.com/Azure/golden"
	"github.com/Azure/grept/pkg/githubclient"
	"github.com/google/go-github/v61/github"
)

var _ Fix = &GitHubRepositoryOidcSubjectClaimFix{}

type GitHubRepositoryOidcSubjectClaimFix struct {
	*golden.BaseBlock
	*BaseFix
	Owner           string   `hcl:"owner"`
	RepoName        string   `hcl:"repo_name"`
	ClaimKeys       []string `hcl:"claim_keys"`
	ClaimUseDefault *bool    `hcl:"claim_use_default"`
}

func (g *GitHubRepositoryOidcSubjectClaimFix) Type() string {
	return "github_repository_oidc_subject_claim"
}

func (g *GitHubRepositoryOidcSubjectClaimFix) Apply() error {
	client, err := githubclient.GetGithubClient()
	if err != nil {
		return fmt.Errorf("cannot create github client: %s", err.Error())
	}
	template := &github.OIDCSubjectClaimCustomTemplate{
		IncludeClaimKeys: g.ClaimKeys,
		UseDefault:       g.ClaimUseDefault,
	}
	resp, err := client.Actions.SetRepoOIDCSubjectClaimCustomTemplate(context.Background(), g.Owner, g.RepoName, template)
	if err != nil {
		return fmt.Errorf("cannot set OIDC subject claims for %s/%s: %+v", g.Owner, g.RepoName, err)
	}
	if resp.StatusCode != 201 {
		return fmt.Errorf("unexpected status code when setting OIDC subject claims for %s/%s: %d", g.Owner, g.RepoName, resp.StatusCode)
	}
	return nil
}
