package pkg

import (
	"bufio"
	"fmt"
	"github.com/ahmetb/go-linq/v3"
	"github.com/alexellis/go-execute/v2"
	"github.com/zclconf/go-cty/cty"
	"io"
	"net/http"
	"os"
	"runtime"
)

var _ Fix = &LocalShellFix{}

type LocalShellFix struct {
	*BaseBlock
	baseFix
	RuleId         string   `hcl:"rule_id"`
	ExecuteCommand []string `hcl:"execute_command" default:"['/bin/sh', '-c']"` // The command used to execute the script.
	InlineShebang  string   `hcl:"inline_shebang,optional" validate:"required_with=Inlines"`
	Inlines        []string `hcl:"inlines,optional" validate:"conflict_with=Script RemoteScript,at_least_one_of=Inlines Script RemoteScript"`
	Script         string   `hcl:"script,optional" validate:"conflict_with=Inlines RemoteScript,at_least_one_of=Inlines Script RemoteScript"`
	RemoteScript   string   `hcl:"remote_script,optional" validate:"conflict_with=Inlines Script,at_least_one_of=Inlines Script RemoteScript,eq=|http_url"`
	OnlyOn         []string `hcl:"only_on,optional" validate:"all_string_in_slice=windows linux darwin openbsd netbsd freebsd dragonfly android solaris plan9"`
}

func (l *LocalShellFix) Type() string {
	return "local_shell"
}

func (l *LocalShellFix) Values() map[string]cty.Value {
	return map[string]cty.Value{
		"inlines":       ToCtyValue(l.Inlines),
		"script":        ToCtyValue(l.Script),
		"remote_script": ToCtyValue(l.RemoteScript),
	}
}

var stopByOnlyOnStub = func() {}

func (l *LocalShellFix) ApplyFix() (err error) {
	if len(l.OnlyOn) > 0 && !linq.From(l.OnlyOn).Contains(runtime.GOOS) {
		stopByOnlyOnStub()
		return nil
	}
	script := l.Script
	if l.RemoteScript != "" {
		script, err = l.downloadFile(l.RemoteScript)
		defer func() {
			_ = os.RemoveAll(script)
		}()
	} else if len(l.Inlines) > 0 {
		if l.InlineShebang == "" {
			l.InlineShebang = "/bin/sh -e"
		}
		script, err = l.createTmpFileForInlines(l.InlineShebang, l.Inlines)
		if script != "" {
			defer func() {
				_ = os.RemoveAll(script)
			}()
		}
		if err != nil {
			return err
		}
	}
	l.ExecuteCommand = append(l.ExecuteCommand, script)
	cmd := execute.ExecTask{
		Command:     l.ExecuteCommand[0],
		Args:        l.ExecuteCommand[1:],
		Env:         nil,
		StreamStdio: false,
	}
	result, err := cmd.Execute(l.Context())
	fmt.Printf("%s\n", result.Stdout)
	if err != nil {
		return err
	}
	if result.ExitCode != 0 {
		return fmt.Errorf("non-zero exit code: %d fix.%s.%s", result.ExitCode, l.Type(), l.Name())
	}
	return nil
}

func (l *LocalShellFix) GetRuleId() string {
	return l.RuleId
}

func (l *LocalShellFix) downloadFile(url string) (string, error) {
	out, err := os.CreateTemp("", "")
	if err != nil {
		return "", err
	}
	defer out.Close()

	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return "", err
	}

	return out.Name(), nil
}

func (l *LocalShellFix) createTmpFileForInlines(shebang string, inlines []string) (string, error) {
	tmp, err := os.CreateTemp("", "grept-local-shell")
	if err != nil {
		return "", err
	}
	defer func() {
		_ = tmp.Close()
	}()
	writer := bufio.NewWriter(tmp)
	_, err = writer.WriteString(fmt.Sprintf("#!%s\n", shebang))
	if err != nil {
		return tmp.Name(), err
	}
	for _, inline := range inlines {
		_, err := writer.WriteString(inline)
		if err != nil {
			return tmp.Name(), err
		}

		_, err = writer.WriteString("\n")
		if err != nil {
			return tmp.Name(), err
		}
	}
	if err := writer.Flush(); err != nil {
		return tmp.Name(), fmt.Errorf("error preparing inlines script %+v", err)
	}
	err = os.Chmod(tmp.Name(), 0700)
	if err != nil {
		return tmp.Name(), err
	}
	return tmp.Name(), nil
}
