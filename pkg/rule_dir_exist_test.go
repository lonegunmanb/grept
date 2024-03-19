package pkg

import (
	"context"
	"fmt"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/suite"
	"testing"
)

type dirExistRuleSuite struct {
	suite.Suite
	*testBase
}

func TestDirExistRuleSuite(t *testing.T) {
	suite.Run(t, new(dirExistRuleSuite))
}

func (s *dirExistRuleSuite) SetupTest() {
	s.testBase = newTestBase()
}

func (s *dirExistRuleSuite) TearDownTest() {
	s.teardown()
}

func (s *dirExistRuleSuite) TestConfig_DirExistRule() {
	_ = s.fs.Mkdir("./testdir", 0644)
	hcl := `  
rule dir_exist test {  
  dir = "./%s"
  fail_on_exist = %t
}`
	cases := []struct {
		desc        string
		dir         string
		failOnExist bool
		wantError   bool
	}{
		{
			desc:        "exist_fail_on_not_exist",
			dir:         "testdir",
			failOnExist: false,
			wantError:   false,
		},
		{
			desc:        "not_exist_fail_on_not_exist",
			dir:         "not_exist",
			failOnExist: false,
			wantError:   true,
		},
		{
			desc:        "exist_fail_on_exist",
			dir:         "testdir",
			failOnExist: true,
			wantError:   true,
		},
		{
			desc:        "not_exist_fail_on_exist",
			dir:         "not_exist",
			failOnExist: true,
			wantError:   false,
		},
	}
	for _, c := range cases {
		s.Run(c.desc, func() {
			code := fmt.Sprintf(hcl, c.dir, c.failOnExist)
			_ = afero.WriteFile(s.fs, fmt.Sprintf("/%s/test.grept.hcl", c.desc), []byte(code), 0644)
			config, err := NewConfig("", fmt.Sprintf("/%s", c.desc), context.Background())
			s.NoError(err)
			_, err = config.Plan()
			s.NoError(err)
			rules := Blocks[Rule](config)
			s.Len(rules, 1)
			rule, ok := rules[0].(*DirExistRule)
			s.True(ok)
			runtimeError := rule.ExecuteDuringPlan()
			s.NoError(runtimeError)
			checkError := rule.CheckError()
			if c.wantError {
				s.NotNil(checkError)
			} else {
				s.NoError(checkError)
			}
		})
	}
}
