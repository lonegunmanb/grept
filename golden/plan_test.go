package golden

import (
	"fmt"
	"github.com/Azure/grept/pkg"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPlan_String(t *testing.T) {
	tests := []struct {
		name string
		plan *pkg.GreptPlan
		want []string
	}{
		{
			name: "Test Plan String",
			plan: &pkg.GreptPlan{
				FailedRules: []*pkg.FailedRule{
					{
						Rule: &pkg.FileHashRule{
							BaseBlock: &BaseBlock{
								blockAddress: "rule.file_hash.test-rule",
								name:         "test-rule",
								id:           "1",
								hb: NewHclBlock(&hclsyntax.Block{
									Type:   "rule",
									Labels: []string{"file_hash", "test-rule"},
								}, nil),
							},
							Glob:      "test-glob",
							Hash:      "test-hash",
							Algorithm: "sha1",
						},
						CheckError: fmt.Errorf("test error"),
					},
				},
				Fixes: map[string]pkg.Fix{
					"test_id": &pkg.LocalFileFix{
						BaseBlock: &BaseBlock{
							name:         "test-fix",
							blockAddress: "fix.local_file.test-fix",
							hb: NewHclBlock(&hclsyntax.Block{
								Type:   "fix",
								Labels: []string{"local_file", "test-fix"},
							}, nil),
						},
						Paths:   []string{"test-path"},
						Content: "test-content",
					},
				},
			},
			want: []string{
				"rule.file_hash.test-rule",
				"fix.local_file.test-fix would be apply",
				"\"test-path\"",
				"\"test-content\"",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := tt.plan
			got := p.String()
			for _, w := range tt.want {
				assert.Contains(t, got, w)
			}
		})
	}
}
