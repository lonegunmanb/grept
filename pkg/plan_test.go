package pkg

import (
	"fmt"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestPlan_String(t *testing.T) {
	tests := []struct {
		name string
		plan *GreptPlan
		want []string
	}{
		{
			name: "Test Plan String",
			plan: &GreptPlan{
				FailedRules: []*FailedRule{
					{
						Rule: &FileHashRule{
							BaseBlock: &BaseBlock{
								blockAddress: "rule.file_hash.test-rule",
								name:         "test-rule",
								id:           "1",
								hb: newHclBlock(&hclsyntax.Block{
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
				Fixes: map[string]Fix{
					"test_id": &LocalFileFix{
						BaseBlock: &BaseBlock{
							name:         "test-fix",
							blockAddress: "fix.local_file.test-fix",
							hb: newHclBlock(&hclsyntax.Block{
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
