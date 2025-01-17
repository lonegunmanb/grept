package pkg_test

import (
	"os"
	"runtime"
	"testing"

	"github.com/Azure/grept/pkg"
	"github.com/Azure/grept/pkg/githubclient"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGitHubRepositoryOidcSubjectClaimRead_IntegrationTest(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("run integration test only under linux to avoid parallel testing issue")
	}
	testToken := os.Getenv("INTEGRATION_TEST_GITHUB_TOKEN")
	if testToken == "" {
		t.Skip("to run this test you must set env INTEGRATION_TEST_GITHUB_TOKEN first")
	}
	t.Setenv("GITHUB_TOKEN", testToken)
	owner := readEssentialEnv(t, "INTEGRATION_TEST_GITHUB_OWNER")
	repoName := readEssentialEnv(t, "INTEGRATION_TEST_GITHUB_REPO_NAME")

	sut := &pkg.GitHubRepositoryOidcSubjectClaimDatasource{
		Owner:    owner,
		RepoName: repoName,
	}
	err := sut.ExecuteDuringPlan()
	require.NoError(t, err)
	assert.NotNil(t, sut.ClaimUseDefault)
	client, err := githubclient.GetGithubClient()
	require.NoError(t, err)
	assertOidcSettings(t, client, owner, repoName, sut.ClaimKeys, sut.ClaimUseDefault)
}
