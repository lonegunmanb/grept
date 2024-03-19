package pkg

import (
	"context"
	"github.com/ahmetb/go-linq/v3"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/suite"
	"gopkg.in/yaml.v3"
	"testing"
)

type yamlTransformSuite struct {
	suite.Suite
	*testBase
}

func (y *yamlTransformSuite) SetupTest() {
	y.testBase = newTestBase()
}

func (y *yamlTransformSuite) TearDownTest() {
	y.testBase.teardown()
}

func TestYamlTransformSuite(t *testing.T) {
	suite.Run(t, new(yamlTransformSuite))
}

func (y *yamlTransformSuite) TestMultipleTransform() {
	hcl := `
	rule "must_be_true" example {
		condition = false
	}
	fix "yaml_transform" example {
		rule_ids = [rule.must_be_true.example.id]
		file_path = "fake.yaml"
		transform {
			yaml_path = "/on/pull_request"
			string_value = "open"
		}
		transform {
			yaml_path = "/permissions/contents"
			string_value = "write"
		}
	}
`
	y.dummyFsWithFiles([]string{"/example/test.grept.hcl"}, []string{hcl})
	config, err := LoadConfig(NewGreptConfig(), "", "/example", context.TODO())
	y.NoError(err)
	_, err = RunGreptPlan(config)
	y.NoError(err)
	fixes := Blocks[Fix](config)
	y.Len(fixes, 1)
	f, ok := fixes[0].(*YamlTransformFix)
	y.True(ok)
	y.Len(f.Transform, 2)
	var paths []string
	linq.From(f.Transform).Select(func(i interface{}) interface{} {
		return i.(YamlTransform).YamlPath
	}).ToSlice(&paths)
	y.Contains(paths, "/on/pull_request")
	y.Contains(paths, "/permissions/contents")
}

func (y *yamlTransformSuite) TestTransformingYaml() {
	yamlContent := `name: pr-check
on:
  workflow_dispatch:
  pull_request:
    types: [ 'opened', 'synchronize' ]
  push:  
    branches:  
      - main

permissions:
  contents: write
  pull-requests: read
  statuses: write
  security-events: write
  actions: read

jobs:
  prepr-check:
    runs-on: ubuntu-latest
    steps:
      - name: checkout
        uses: actions/checkout@f43a0e5ff2bd294095638e18286ca9a3d1956744 #v3.6.0
      - name: pr-check
        run: |
          docker run --rm -v $(pwd):/src -w /src -e SKIP_CHECKOV -e GITHUB_TOKEN mcr.microsoft.com/azterraform:latest make pr-check
      - name: Set up Go
        uses: actions/setup-go@6edd4406fa81c3da01a34fa6f6343087c207a568 #3.5.0
        with:
          go-version: 1.21.3
`
	yamlPath := "./target.yaml"
	y.dummyFsWithFiles([]string{yamlPath}, []string{yamlContent})
	sut := &YamlTransformFix{
		BaseBlock: &BaseBlock{
			c:    NewGreptConfig(),
			name: "test",
			id:   "test",
		},
		FilePath: yamlPath,
		Transform: []YamlTransform{
			{
				YamlPath:    "/name",
				StringValue: "new-name",
			},
			// Support string replacement only for now!
			//{
			//	YamlPath: "/on/pull_request/types",
			//	Value:    "[ 'opened' ]",
			//},
			{
				YamlPath:    `/jobs/prepr-check/steps/~{"name":"checkout"}/uses`,
				StringValue: "actions/checkout@v3.7.0",
			},
		},
	}
	err := sut.Apply()
	y.NoError(err)
	yf, err := afero.ReadFile(y.fs, yamlPath)
	y.NoError(err)
	var resultYaml = make(map[string]any)
	err = yaml.Unmarshal(yf, &resultYaml)
	y.NoError(err)
	y.Equal("new-name", resultYaml["name"])
}
