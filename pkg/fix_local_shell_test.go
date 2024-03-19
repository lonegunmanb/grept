package pkg

import (
	"context"
	"fmt"
	"github.com/ahmetb/go-linq/v3"
	"github.com/google/uuid"
	"github.com/prashantv/gostub"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"runtime"
	"strings"
	"testing"
)

type localExecFixSuite struct {
	suite.Suite
	*testBase
}

func TestLocalExecFixSuite(t *testing.T) {
	suite.Run(t, new(localExecFixSuite))
}

func (s *localExecFixSuite) SetupTest() {
	s.testBase = newTestBase()
}

func (s *localExecFixSuite) TearDownTest() {
	s.teardown()
}

func (s *localExecFixSuite) TestLocalExecShell_CommandValidate() {
	cases := []struct {
		desc      string
		f         *LocalShellFix
		wantError bool
	}{
		{
			desc: "Inlines only",
			f: &LocalShellFix{
				Inlines: []string{"echo hello"},
			},
			wantError: false,
		},
		{
			desc: "Script only",
			f: &LocalShellFix{
				Inlines: nil,
				Script:  "./bash.sh",
			},
			wantError: false,
		},
		{
			desc: "RemoteScript only",
			f: &LocalShellFix{
				RemoteScript: "https://raw.githubusercontent.com/cloudposse/build-harness/master/bin/install.sh",
			},
			wantError: false,
		},
		{
			desc: "RemoteScript http only",
			f: &LocalShellFix{
				RemoteScript: "http://raw.githubusercontent.com/cloudposse/build-harness/master/bin/install.sh",
			},
			wantError: false,
		},
		{
			desc: "non-http RemoteScript",
			f: &LocalShellFix{
				RemoteScript: "s3::https://s3.amazonaws.com/bucket/foo",
			},
			wantError: true,
		},
		{
			desc: "Inlines&Script",
			f: &LocalShellFix{
				Inlines: []string{"echo hello"},
				Script:  "./bash.sh",
			},
			wantError: true,
		},
		{
			desc: "Script&RemoteScrip",
			f: &LocalShellFix{
				Inlines:      nil,
				Script:       "./bash.sh",
				RemoteScript: "https://raw.githubusercontent.com/cloudposse/build-harness/master/bin/install.sh",
			},
			wantError: true,
		},
		{
			desc: "Inlines&RemoteScript",
			f: &LocalShellFix{
				Inlines:      []string{"echo hello"},
				RemoteScript: "https://raw.githubusercontent.com/cloudposse/build-harness/master/bin/install.sh",
			},
			wantError: true,
		},
		{
			desc:      "No command",
			f:         &LocalShellFix{},
			wantError: true,
		},
		{
			desc: "InlineShebang Only",
			f: &LocalShellFix{
				Script:        "/test.sh",
				InlineShebang: "#!/bin/sh",
			},
			wantError: true,
		},
		{
			desc: "InlineShebang with inlines",
			f: &LocalShellFix{
				InlineShebang: "#!/bin/sh",
				Inlines: []string{
					"echo hello",
				},
			},
			wantError: false,
		},
		{
			desc: "Invalid only_on",
			f: &LocalShellFix{
				Inlines: []string{
					"echo hello",
				},
				OnlyOn: []string{
					"Linux", // should be linux
				},
			},
			wantError: true,
		},
		{
			desc: "Valid only_on",
			f: &LocalShellFix{
				Inlines: []string{
					"echo hello",
				},
				OnlyOn: []string{
					"linux",
					"windows",
					"darwin",
				},
			},
			wantError: false,
		},
	}
	for _, c := range cases {
		s.Run(c.desc, func() {
			err := Validate.Struct(*c.f)
			if c.wantError {
				assert.Error(s.T(), err)
			} else {
				assert.NoError(s.T(), err)
			}
		})
	}
}

func (s *localExecFixSuite) TestLocalExecShell_ShouldReturnDirectlyIfOnlyOnMismatch() {
	quit := false
	stub := gostub.Stub(&stopByOnlyOnStub, func() {
		quit = true
	})
	defer stub.Reset()

	onlyOn := []string{"linux", "windows", "darwin"}
	linq.From(onlyOn).Where(func(i interface{}) bool {
		return i.(string) != runtime.GOOS // exclude the current os
	}).ToSlice(&onlyOn)

	sut := &LocalShellFix{
		Inlines: []string{"echo hello"},
		OnlyOn:  onlyOn,
	}
	err := sut.Apply()
	assert.NoError(s.T(), err)
	assert.True(s.T(), quit)
}

func (s *localExecFixSuite) TestLocalShellFix_ApplyFix_Inlines() {
	t := s.T()
	if runtime.GOOS == "windows" {
		t.Skip("cannot run this test on windows")
	}
	fix := &LocalShellFix{
		BaseBlock: &BaseBlock{
			id:   "test",
			name: "test",
			c: &BaseConfig{
				ctx: context.TODO(),
			},
		},
		ExecuteCommand: []string{"/bin/sh", "-c"},
		Inlines:        []string{`echo "Hello, World!"`},
	}

	r, w, _ := os.Pipe()
	stub := gostub.Stub(&os.Stdout, w)
	defer stub.Reset()

	err := fix.Apply()
	require.NoError(t, err)
	_ = w.Close()

	out, _ := io.ReadAll(r)
	// ExecuteDuringPlan that the command output was as expected
	assert.Contains(t, string(out), "Hello, World!")
}

func (s *localExecFixSuite) TestLocalShellFix_ApplyFix_script() {
	t := s.T()
	if runtime.GOOS == "windows" {
		t.Skip("cannot run this test on windows")
	}
	tmpScript, err := os.CreateTemp("", "grep-test")
	_ = os.Chmod(tmpScript.Name(), 0700)
	require.NoError(t, err)
	defer func() {
		_ = os.Remove(tmpScript.Name())
	}()
	_, _ = tmpScript.WriteString(`#!/bin/sh
		echo "Hello, World!"
`)
	_ = tmpScript.Close()

	fix := &LocalShellFix{
		BaseBlock: &BaseBlock{
			id:   "test",
			name: "test",
			c: &BaseConfig{
				ctx: context.TODO(),
			},
		},
		ExecuteCommand: []string{"/bin/sh", "-c"},
		Script:         tmpScript.Name(),
	}

	r, w, _ := os.Pipe()
	stub := gostub.Stub(&os.Stdout, w)
	defer stub.Reset()

	err = fix.Apply()
	require.NoError(t, err)
	_ = w.Close()

	out, _ := io.ReadAll(r)
	// ExecuteDuringPlan that the command output was as expected
	assert.Contains(t, string(out), "Hello, World!")
}

func (s *localExecFixSuite) TestLocalShellFix_ApplyFix_RemoteScript() {
	t := s.T()
	if runtime.GOOS == "windows" {
		t.Skip("cannot run this test on windows")
	}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Mock response
		w.Header().Set("Content-Type", "text/plain")
		_, _ = fmt.Fprintln(w, `#!/bin/sh
		echo "Hello, World!"`)
	}))
	defer ts.Close()

	fix := &LocalShellFix{
		BaseBlock: &BaseBlock{
			id:   "test",
			name: "test",
			c: &BaseConfig{
				ctx: context.TODO(),
			},
		},
		ExecuteCommand: []string{"/bin/sh", "-c"},
		RemoteScript:   fmt.Sprintf("%s/test.sh", ts.URL),
	}

	r, w, _ := os.Pipe()
	stub := gostub.Stub(&os.Stdout, w)
	defer stub.Reset()

	err := fix.Apply()
	require.NoError(t, err)
	_ = w.Close()

	out, _ := io.ReadAll(r)
	// ExecuteDuringPlan that the command output was as expected
	assert.Contains(t, string(out), "Hello, World!")
}

func (s *localExecFixSuite) TestLocalShellFix_ApplyFix() {
	t := s.T()
	if runtime.GOOS == "windows" {
		t.Skip("cannot run this test on windows")
	}
	rand := uuid.NewString()
	t.Setenv("TMP_VAR", rand)
	temp, err := os.CreateTemp("", "test_grept")
	require.NoError(t, err)
	defer func() {
		_ = os.Remove(temp.Name())
	}()
	err = temp.Close()
	require.NoError(t, err)
	hcl := fmt.Sprintf(`
	rule "must_be_true" "example" {
		condition = false
	}

	fix "local_shell" "example" {
		rule_ids = [rule.must_be_true.example.id]
		inlines = [
			"echo ${env("TMP_VAR")}>%s",
			"echo ${env("TMP_VAR2")}>>%s",
		]
		env = {
			TMP_VAR2 = "${strrev(env("TMP_VAR"))}"
		}
	}
`, temp.Name(), temp.Name())
	s.dummyFsWithFiles([]string{"/example/test.grept.hcl"}, []string{hcl})
	config, err := NewConfig("", "/example", context.TODO())
	require.NoError(t, err)
	plan, err := config.Plan()
	require.NoError(t, err)
	err = plan.Apply()
	require.NoError(t, err)
	tmpFileForRead, err := os.ReadFile(temp.Name())
	require.NoError(t, err)
	assert.Contains(t, string(tmpFileForRead), rand, "predefined env should be honored")
	assert.Contains(t, string(tmpFileForRead), reverse(rand), "env defined in `env` map should be honored")
}

func (s *localExecFixSuite) TestLocalShellFix_ApplyFix_UserAssignedEnvShouldBeLocal() {
	t := s.T()
	if runtime.GOOS == "windows" {
		t.Skip("cannot run this test on windows")
	}
	temp0, err := createTempFile(t, "test_grept")
	s.NoError(err)
	defer func() {
		_ = os.Remove(temp0.Name())
	}()
	temp1, err := createTempFile(t, "test_grept")
	s.NoError(err)
	defer func() {
		_ = os.Remove(temp1.Name())
	}()
	hcl := fmt.Sprintf(`
	rule "must_be_true" "example" {
		condition = false
	}

	fix "local_shell" "example0" {
		rule_ids = [rule.must_be_true.example.id]
		inlines = [
			"echo \"${env("TMP_VAR")}\">%s",
		]
		env = {
			TMP_VAR = "0"
		}
	}
	
	fix "local_shell" "example1" {
		rule_ids = [rule.must_be_true.example.id]
		inlines = [
			"echo \"${env("TMP_VAR")}\">%s",
		]
		env = {
			TMP_VAR = "1"
		}
	}
`, temp0.Name(), temp1.Name())
	s.dummyFsWithFiles([]string{"/example/test.grept.hcl"}, []string{hcl})
	config, err := NewConfig("", "/example", context.TODO())
	require.NoError(t, err)
	plan, err := config.Plan()
	require.NoError(t, err)
	err = plan.Apply()
	require.NoError(t, err)
	content0, err := os.ReadFile(temp0.Name())
	require.NoError(t, err)
	content1, err := os.ReadFile(temp1.Name())
	require.NoError(t, err)
	assert.Contains(t, string(content0), "0")
	assert.Contains(t, string(content1), "1")
}

func (s *localExecFixSuite) TestLocalShellFix_ApplyFix_scriptWithUserAssignedEnv() {
	t := s.T()
	if runtime.GOOS == "windows" {
		t.Skip("cannot run this test on windows")
	}
	resultFile, err := createTempFile(t, "testgrept")
	require.NoError(t, err)
	tmpScript, err := os.CreateTemp("", "grep-test")
	_ = os.Chmod(tmpScript.Name(), 0700)
	require.NoError(t, err)
	defer func() {
		_ = os.Remove(tmpScript.Name())
	}()
	_, _ = tmpScript.WriteString(fmt.Sprintf(`#!/bin/sh
		echo $TMP_ENV>%s
`, resultFile.Name()))
	_ = tmpScript.Close()

	rand := uuid.NewString()
	hcl := fmt.Sprintf(`
	rule "must_be_true" "example" {
		condition = false
	}

	fix "local_shell" "example" {
		rule_ids = [rule.must_be_true.example.id]
		script = "%s"
		env = {
			TMP_ENV = "%s"
		}
	}
`, tmpScript.Name(), rand)
	s.dummyFsWithFiles([]string{"/example/test.grept.hcl"}, []string{hcl})
	config, err := NewConfig("", "/example", context.TODO())
	require.NoError(t, err)
	plan, err := config.Plan()
	require.NoError(t, err)
	err = plan.Apply()
	require.NoError(t, err)
	result, err := os.ReadFile(resultFile.Name())
	require.NoError(t, err)
	assert.Equal(t, strings.TrimSuffix(string(result), "\n"), rand, "user assigned env should be honored with script")
}

func createTempFile(t *testing.T, pattern string) (*os.File, error) {
	tf, err := os.CreateTemp("", pattern)
	require.NoError(t, err)
	err = tf.Close()
	require.NoError(t, err)
	return tf, err
}

func reverse(s string) string {
	r := []rune(s)
	for i, j := 0, len(r)-1; i < len(r)/2; i, j = i+1, j-1 {
		r[i], r[j] = r[j], r[i]
	}
	return string(r)
}
