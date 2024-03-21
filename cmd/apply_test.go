package cmd

import (
	"context"
	"fmt"
	"github.com/Azure/grept/pkg"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/prashantv/gostub"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestApplyFunc(t *testing.T) {
	expectedContent := "Mock server response"
	// Create a test server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, expectedContent)
	}))
	defer ts.Close()

	// Mock config
	configContent := `      
		data "http" "test" {      
			url = "` + ts.URL + `"      
		}      
      
		rule "file_hash" "test" {      
			glob = "test.txt"      
			hash = sha1(data.http.test.response_body)      
		}    
    
		fix "local_file" "test" {    
			paths = ["test.txt"]    
			content = data.http.test.response_body    
			rule_ids = [rule.file_hash.test.id]    
		}    
	`

	mockFs := afero.NewMemMapFs()
	stub := gostub.Stub(&pkg.FsFactory, func() afero.Fs {
		return mockFs
	})
	defer stub.Reset()

	_ = afero.WriteFile(mockFs, "./test.txt", []byte("incorrect content"), 0644)
	_ = afero.WriteFile(mockFs, "./test_config.grept.hcl", []byte(configContent), 0644)

	// Redirect Stdin and Stdout
	r, w, _ := os.Pipe()
	stub.Stub(&os.Stdout, w)

	cmd := NewApplyCmd()
	cmd.SetContext(context.TODO())
	_ = cmd.Flags().Set("auto", "true")
	// Run function
	err := cmd.RunE(cmd, []string{"apply", "."})
	require.NoError(t, err)

	// Reset Stdout
	_ = w.Close()

	// Read Stdout
	out, _ := io.ReadAll(r)
	output := string(out)

	assert.Contains(t, output, "Plan applied successfully.")

	// ExecuteDuringPlan if the fix was applied
	fixedContent, _ := afero.ReadFile(mockFs, "test.txt")
	assert.Equal(t, expectedContent, string(fixedContent))
}
