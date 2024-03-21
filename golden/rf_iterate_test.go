package golden

import (
	"github.com/hashicorp/hcl/v2"
	"testing"

	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/stretchr/testify/assert"
)

func TestRefIterator(t *testing.T) {
	tests := []struct {
		name string
		hcl  string
		want string
	}{
		{
			name: "data ref iterator",
			hcl:  `data.source.attribute`,
			want: "data.source.attribute",
		},
		{
			name: "rule ref iterator",
			hcl:  `rule.my_rule.id`,
			want: "rule.my_rule.id",
		},
		{
			name: "fix ref iterator",
			hcl:  `fix.my_fix.id`,
			want: "fix.my_fix.id",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr, _ := hclsyntax.ParseExpression([]byte(tt.hcl), "", hcl.InitialPos)
			ts := expr.Variables()

			root := name(ts[0][0])
			iterator := refIters[root]
			got := iterator(ts[0], 0)

			assert.Equal(t, tt.want, got[0])
		})
	}
}
