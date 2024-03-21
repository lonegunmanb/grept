package pkg

import (
	"bufio"
	"fmt"
	"github.com/Azure/grept/golden"
	"github.com/ahmetb/go-linq/v3"
	"github.com/emirpasic/gods/sets/hashset"
	"github.com/spf13/afero"
	"strings"
)

var _ Fix = &GitIgnoreFix{}

type GitIgnoreFix struct {
	*golden.BaseBlock
	*BaseFix
	Exist    []string `hcl:"exist,optional" validate:"at_least_one_of=Exist NotExist"`
	NotExist []string `hcl:"not_exist,optional" validate:"at_least_one_of=Exist NotExist"`
}

func (g *GitIgnoreFix) Type() string {
	return "git_ignore"
}

func (g *GitIgnoreFix) Apply() error {
	fs := FsFactory()
	gitIgnoreFile := ".gitignore"
	exist, err := afero.Exists(fs, gitIgnoreFile)
	if err != nil {
		return fmt.Errorf("error on checking existing gitignore: %+v, fix.%s.%s, %s", err, g.BlockType(), g.Name(), g.HclBlock().Range().String())
	}
	if !exist {
		err := afero.WriteFile(fs, gitIgnoreFile, []byte{}, 0600)
		if err != nil {
			return fmt.Errorf("error on ensuring gitignore: %+v, fix.%s.%s, %s", err, g.BlockType(), g.Name(), g.HclBlock().Range().String())
		}
	}
	f, err := fs.Open(gitIgnoreFile)
	if err != nil {
		return fmt.Errorf("error on reading gitignore: %+v, fix.%s.%s, %s", err, g.BlockType(), g.Name(), g.HclBlock().Range().String())
	}

	linq.From(g.NotExist).Select(g.trimLineAny).ToSlice(&g.NotExist)
	linq.From(g.Exist).Select(g.trimLineAny).ToSlice(&g.Exist)
	notAllowed := hashset.New()
	for _, i := range g.NotExist {
		notAllowed.Add(i)
	}

	sb := strings.Builder{}
	scanner := bufio.NewScanner(f)
	existed := hashset.New()
	for scanner.Scan() {
		line := scanner.Text()
		raw := line
		line = g.trimLine(line)
		existed.Add(line)
		if notAllowed.Contains(line) {
			continue
		}
		sb.WriteString(raw)
		sb.WriteString("\n")
	}
	err = f.Close()
	if err != nil {
		return fmt.Errorf("error on closing gitignore: %+v, fix.%s.%s, %s", err, g.BlockType(), g.Name(), g.HclBlock().Range().String())
	}
	for _, item := range g.Exist {
		if !existed.Contains(item) {
			sb.WriteString(item)
			sb.WriteString("\n")
		}
	}
	err = afero.WriteFile(fs, gitIgnoreFile, []byte(sb.String()), 0600)
	if err != nil {
		return fmt.Errorf("error on writing gitignore: %+v, fix.%s.%s, %s", err, g.BlockType(), g.Name(), g.HclBlock().Range().String())
	}
	return nil
}

func (g *GitIgnoreFix) trimLine(line string) string {
	return strings.TrimFunc(line, func(r rune) bool {
		return r == '\t' || r == ' ' || r == '\n' || r == '\r'
	})
}

func (g *GitIgnoreFix) trimLineAny(i any) any {
	return g.trimLine(i.(string))
}
