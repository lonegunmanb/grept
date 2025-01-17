package pkg_test

import (
	"context"
	"os"
	"runtime"
	"strconv"
	"strings"
	"testing"

	"github.com/Azure/grept/pkg"
	"github.com/Azure/grept/pkg/githubclient"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGithubRepositoryOidcSubjectFix_IntegrationTest(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("run integration test only under linux to avoid parallel testing issue")
	}
	testToken := os.Getenv("INTEGRATION_TEST_GITHUB_TOKEN")
	if testToken == "" {
		t.Skip("to run this test you must set env INTEGRATION_TEST_GITHUB_TOKEN first")
	}
	t.Setenv("GITHUB_TOKEN", testToken)
	client, err := githubclient.GetGithubClient()
	require.NoError(t, err)
	owner := readEssentialEnv(t, "INTEGRATION_TEST_GITHUB_OWNER")
	repoName := readEssentialEnv(t, "INTEGRATION_TEST_GITHUB_REPO_NAME")
	claimKeysEnv := readEssentialEnv(t, "INTEGRATION_TEST_GITHUB_OIDC_CLAIMKEYS") // This should be a semicolon separated list of claimKeys
	claimKeys := parseClaimKeys(claimKeysEnv)
	useDefaultEnv := readEssentialEnv(t, "INTEGRATION_TEST_GITHUB_OIDC_USE_DEFAULT")
	useDefault := strings.EqualFold(useDefaultEnv, "true")
	require.NoError(t, err)

	// Backup the current settings and restore after the test is complete
	backupOidcSettings, _, err := client.Actions.GetRepoOIDCSubjectClaimCustomTemplate(context.Background(), owner, repoName)
	require.NoError(t, err)
	cleanup := func() {
		_, err := client.Actions.SetRepoOIDCSubjectClaimCustomTemplate(context.Background(), owner, repoName, backupOidcSettings)
		require.NoError(t, err)
	}
	defer cleanup()

	tr := true
	fa := false
	// Test scenarios here
	cases := []pkg.GitHubRepositoryOidcSubjectClaimFix{
		{
			Owner:           owner,
			RepoName:        repoName,
			ClaimUseDefault: &useDefault,
			ClaimKeys:       claimKeys,
		},
		{
			Owner:           owner,
			RepoName:        repoName,
			ClaimUseDefault: &tr,
			ClaimKeys:       nil,
		},
		{
			Owner:           owner,
			RepoName:        repoName,
			ClaimUseDefault: &fa,
			ClaimKeys:       nil,
		},
		{
			Owner:           owner,
			RepoName:        repoName,
			ClaimUseDefault: &useDefault,
			ClaimKeys:       claimKeys,
		},
	}
	for i, c := range cases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			err := c.Apply()
			require.NoError(t, err)
			assertOidcSettings(t, client, owner, repoName, c.ClaimKeys, c.ClaimUseDefault)
		})
	}
}

func parseClaimKeys(claimKeys string) []string {
	return strings.Split(claimKeys, ";")
}

func assertOidcSettings(t *testing.T, client *githubclient.Client, owner, repoName string, subjects []string, useDefault *bool) {
	response, _, err := client.Actions.GetRepoOIDCSubjectClaimCustomTemplate(context.Background(), owner, repoName)
	require.NoError(t, err)
	assert.EqualValues(t, response.UseDefault, useDefault)
	assert.Equal(t, subjects, response.IncludeClaimKeys)
}
