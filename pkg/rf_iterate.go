package pkg

import (
	"fmt"
	"strings"

	"github.com/hashicorp/hcl/v2"
)

type refIterator func(t []hcl.Traverser, i int) []string

// iterator return a refIterator that travel a list of tokens, and return referenced block's address inside this expression
func iterator(keyword string, addressLength int) refIterator {
	return func(ts []hcl.Traverser, i int) []string {
		var r []string
		if len(ts) == 0 {
			return nil
		}
		if name(ts[i]) != keyword {
			return nil
		}
		if len(ts) < i+addressLength {
			return nil
		}
		remain := addressLength
		sb := strings.Builder{}
		// data.xyz.abc
		for j := i; remain > 0; j++ {
			sb.WriteString(name(ts[j]))
			remain--
			if remain > 0 {
				sb.WriteString(".")
			}
		}
		r = []string{sb.String()}
		//potential index, like data.xyz.abc["foo"]
		if len(ts) > i+addressLength+1 {
			index, ok := ts[i+addressLength].(hcl.TraverseIndex)
			if ok {
				sb.WriteString(fmt.Sprintf(`[%s]`, CtyValueToString(index.Key)))
				r = append(r, sb.String())
			}
		}
		return r
	}
}

func name(t hcl.Traverser) string {
	switch tp := t.(type) {
	case hcl.TraverseRoot:
		{
			return tp.Name
		}
	case hcl.TraverseAttr:
		{
			return tp.Name
		}
	default:
		{
			return ""
		}
	}
}
