package pkg

import "github.com/Azure/golden"

func init() {
	golden.MetaAttributeNames.Add("rule_ids")
	golden.RegisterBaseBlock(func() golden.BlockType {
		return new(BaseRule)
	})
	golden.RegisterBaseBlock(func() golden.BlockType {
		return new(BaseData)
	})
	golden.RegisterBaseBlock(func() golden.BlockType {
		return new(BaseFix)
	})
	registerRule()
	registerFix()
	registerData()
}

func registerFix() {
	golden.RegisterBlock(new(CopyFileFix))
	golden.RegisterBlock(new(LocalFileFix))
	golden.RegisterBlock(new(RenameFileFix))
	golden.RegisterBlock(new(RmLocalFileFix))
	golden.RegisterBlock(new(LocalShellFix))
	golden.RegisterBlock(new(GitIgnoreFix))
	golden.RegisterBlock(new(YamlTransformFix))
	golden.RegisterBlock(new(GitHubRepositoryCollaboratorsFix))
	golden.RegisterBlock(new(GitHubRepositoryEnvironmentsFix))
	golden.RegisterBlock(new(GitHubTeamFix))
	golden.RegisterBlock(new(GitHubTeamMembersFix))
	golden.RegisterBlock(new(GitHubTeamRepositoryFix))
	golden.RegisterBlock(new(GitHubRepositoryOidcSubjectClaimFix))
}

func registerRule() {
	golden.RegisterBlock(new(FileExistRule))
	golden.RegisterBlock(new(FileHashRule))
	golden.RegisterBlock(new(MustBeTrueRule))
	golden.RegisterBlock(new(DirExistRule))
}

func registerData() {
	golden.RegisterBlock(new(HttpDatasource))
	golden.RegisterBlock(new(GitIgnoreDatasource))
	golden.RegisterBlock(new(GitHubRepositoryCollaboratorsDatasource))
	golden.RegisterBlock(new(GitHubRepositoryEnvironmentsDatasource))
	golden.RegisterBlock(new(GitHubRepositoryTeamsDatasource))
	golden.RegisterBlock(new(GitHubTeamDatasource))
	golden.RegisterBlock(new(GitHubRepositoryOidcSubjectClaimDatasource))
}
