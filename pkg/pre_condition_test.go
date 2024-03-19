package pkg

import (
	"fmt"
	"github.com/hashicorp/hcl/v2"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"net/http"
	"net/http/httptest"
	"testing"
)

type preConditionSuite struct {
	suite.Suite
	*testBase
}

func TestPreConditionSuite(t *testing.T) {
	suite.Run(t, new(preConditionSuite))
}

func (s *preConditionSuite) SetupTest() {
	s.testBase = newTestBase()
}

func (s *preConditionSuite) TearDownTest() {
	s.teardown()
}

func (s *preConditionSuite) TestPreCondition_PassedHardcodedCondition() {
	content := `
    rule "file_hash" sample {
        glob = "*.txt"
        hash = "abc123"
        algorithm = "sha256"
        precondition {
            condition = true
        }
    }
    `
	s.dummyFsWithFiles([]string{"test.grept.hcl"}, []string{content})
	config, err := LoadConfig(NewGreptConfig(), "", "", nil)
	s.NoError(err)
	_, err = RunGreptPlan(config)
	s.NoError(err)
	rules := Blocks[Rule](config)
	s.Len(rules, 1)
	fhr, ok := rules[0].(*FileHashRule)
	require.True(s.T(), ok)
	check, err := fhr.PreConditionCheck(new(hcl.EvalContext))
	s.NoError(err)
	s.Len(check, 0)
}

func (s *preConditionSuite) TestPreCondition_FaileddHardcodedConditionShouldFailedPlan() {
	content := `
    rule "file_hash" sample {
        glob = "*.txt"
        hash = "abc123"
        algorithm = "sha256"
        precondition {
            condition = false
			error_message = "this precondition must be true"
        }
    }
    `
	s.dummyFsWithFiles([]string{"test.grept.hcl"}, []string{content})
	config, err := LoadConfig(NewGreptConfig(), "", "", nil)
	s.NoError(err)
	_, err = RunGreptPlan(config)
	s.NotNil(err)
	s.Contains(err.Error(), "this precondition must be true")
}

func (s *preConditionSuite) TestPreCondition_FaileddHardcodedCondition() {
	content := `
    rule "file_hash" sample {
        glob = "*.txt"
        hash = "abc123"
        algorithm = "sha256"
        precondition {
            condition = false
			error_message = "this precondition must be true"
        }
    }
    `
	s.dummyFsWithFiles([]string{"test.grept.hcl"}, []string{content})
	config, err := LoadConfig(NewGreptConfig(), "", "", nil)
	s.NoError(err)
	_, err = RunGreptPlan(config)
	s.NotNil(err)
	s.Contains(err.Error(), "this precondition must be true")
	rules := Blocks[Rule](config)
	s.Len(rules, 1)
	fhr, ok := rules[0].(*FileHashRule)
	require.True(s.T(), ok)
	check, err := fhr.PreConditionCheck(new(hcl.EvalContext))
	s.NoError(err)
	s.Len(check, 1)
	s.Equal("this precondition must be true", check[0].ErrorMessage)
}

func (s *preConditionSuite) TestPreCondition_FunctionCallInCondition() {
	s.T().Setenv("KEY", "VALUE")
	content := `
    rule "file_hash" sample {
        glob = "*.txt"
        hash = "abc123"
        algorithm = "sha256"
        precondition {
            condition = env("KEY") != "VALUE"
			error_message = "this precondition must be true"
        }
    }
    `
	s.dummyFsWithFiles([]string{"test.grept.hcl"}, []string{content})
	config, err := LoadConfig(NewGreptConfig(), "", "", nil)
	s.NoError(err)
	_, err = RunGreptPlan(config)
	s.NotNil(err)
	s.Contains(err.Error(), "this precondition must be true")
}

func (s *preConditionSuite) TestMultiplePreConditions_containFailedCheck() {
	// Test case when one of the preconditions fails
	content := `
        rule "file_hash" sample {
            glob = "*.txt"
            hash = "abc123"
            algorithm = "sha256"
            precondition {
                condition = true
            }
            precondition {
                condition = false
                error_message = "this precondition must be true"
            }
        }
    `
	s.dummyFsWithFiles([]string{"test.grept.hcl"}, []string{content})
	config, err := LoadConfig(NewGreptConfig(), "", "", nil)
	s.NoError(err)
	_, err = RunGreptPlan(config)
	s.NotNil(err)
	s.Contains(err.Error(), "this precondition must be true")
}

func (s *preConditionSuite) TestMultiplePreConditions_allPassedCheck() {
	// Test case when all preconditions pass
	content := `
        rule "file_hash" sample {
            glob = "*.txt"
            hash = "abc123"
            algorithm = "sha256"
            precondition {
                condition = true
            }
            precondition {
                condition = true
            }
        }
    `
	s.dummyFsWithFiles([]string{"test.grept.hcl"}, []string{content})
	config, err := LoadConfig(NewGreptConfig(), "", "", nil)
	s.NoError(err)
	_, err = RunGreptPlan(config)
	s.NoError(err)
}

func (s *preConditionSuite) TestPreCondition_ReferOtherBlockAttribute() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("Mock server content"))
	}))
	defer server.Close()

	sampleConfig := fmt.Sprintf(`
    data "http" "foo" {
        url = "%s"
    }

    rule "must_be_true" "bar" {
		condition = true
        precondition {
            condition = data.http.foo.response_body == "Mock server content"
            error_message = "Precondition check failed"
        }
    }
    `, server.URL)
	s.dummyFsWithFiles([]string{"test.grept.hcl"}, []string{sampleConfig})

	// Parse the config
	config, err := LoadConfig(NewGreptConfig(), "", "", nil)
	s.NoError(err)

	// Plan the parsed configuration
	_, err = RunGreptPlan(config)
	s.NoError(err) // Expect no error as the precondition should pass
}

func (s *preConditionSuite) TestPreCondition_ReferOtherBlockAttributeFailedCheck() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("Mock server content"))
	}))
	defer server.Close()

	sampleConfig := fmt.Sprintf(`
    data "http" "foo" {
        url = "%s"
    }

    rule "must_be_true" "bar" {
		condition = true
        precondition {
            condition = data.http.foo.response_body != "Mock server content"
            error_message = "Precondition check failed"
        }
    }
    `, server.URL)
	s.dummyFsWithFiles([]string{"test.grept.hcl"}, []string{sampleConfig})

	// Parse the config
	config, err := LoadConfig(NewGreptConfig(), "", "", nil)
	s.NoError(err)

	// Plan the parsed configuration
	_, err = RunGreptPlan(config)
	s.Contains(err.Error(), "Precondition check failed")
}

func (s *preConditionSuite) TestPreCondition_ReferMultipleBlockAttributes() {
	// Create two mock HTTP servers that return specific contents
	server1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("content1"))
	}))
	defer server1.Close()

	server2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("content2"))
	}))
	defer server2.Close()

	// Define a sample config for testing
	sampleConfig := fmt.Sprintf(`
    data "http" "foo1" {
        url = "%s"
    }

    data "http" "foo2" {
        url = "%s"
    }

    rule "must_be_true" "bar" {
		condition = true
        precondition {
            condition = data.http.foo1.response_body == "content1" && data.http.foo2.response_body == "content2"
            error_message = "Precondition check failed"
        }
    }
    `, server1.URL, server2.URL)
	s.dummyFsWithFiles([]string{"test.grept.hcl"}, []string{sampleConfig})

	// Parse the config
	config, err := LoadConfig(NewGreptConfig(), "", "", nil)
	s.NoError(err)

	// Plan the parsed configuration
	_, err = RunGreptPlan(config)
	s.NoError(err) // Expect no error as the precondition should pass
}

func (s *preConditionSuite) TestPreCondition_ReferMultipleBlockAttributesFailedCheck() {
	// Create two mock HTTP servers that return specific contents
	server1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("content1"))
	}))
	defer server1.Close()

	server2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("content2"))
	}))
	defer server2.Close()

	// Define a sample config for testing
	sampleConfig := fmt.Sprintf(`
    data "http" "foo1" {
        url = "%s"
    }

    data "http" "foo2" {
        url = "%s"
    }

    rule "must_be_true" "bar" {
		condition = true
        precondition {
            condition = data.http.foo1.response_body != "content1" && data.http.foo2.response_body == "content2"
            error_message = "Precondition check failed"
        }
    }
    `, server1.URL, server2.URL)
	s.dummyFsWithFiles([]string{"test.grept.hcl"}, []string{sampleConfig})

	// Parse the config
	config, err := LoadConfig(NewGreptConfig(), "", "", nil)
	s.NoError(err)

	// Plan the parsed configuration
	_, err = RunGreptPlan(config)
	s.Contains(err.Error(), "Precondition check failed")
}

func (s *preConditionSuite) TestPreCondition_ReferOtherBlockAttributeWithForEach() {
	// Create a mock HTTP server that returns specific contents
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("Mock server content"))
	}))
	defer server.Close()

	// Define a sample config for testing
	sampleConfig := fmt.Sprintf(`
        locals {
            items = toset(["item1", "item2", "item3"])
        }

        data "http" "foo" {
            for_each = local.items

            url = "%s"
        }

        rule "must_be_true" "bar" {
            for_each = local.items

            condition = true
            precondition {
                condition = data.http.foo[each.value].response_body == "Mock server content"
                error_message = "Precondition check failed"
            }
        }
    `, server.URL)
	s.dummyFsWithFiles([]string{"test.grept.hcl"}, []string{sampleConfig})

	// Parse the config
	config, err := LoadConfig(NewGreptConfig(), "", "", nil)
	s.NoError(err)

	// Plan the parsed configuration
	_, err = RunGreptPlan(config)
	s.NoError(err) // Expect no error as the precondition should pass
}

func (s *preConditionSuite) TestPreCondition_ReferOtherBlockAttributeWithForEach_FailedCheck() {
	// Create a mock HTTP server that returns specific contents
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("Mock server content"))
	}))
	defer server.Close()

	// Define a sample config for testing
	sampleConfig := fmt.Sprintf(`
        locals {
            items = toset(["item1", "item2", "item3"])
        }

        data "http" "foo" {
            for_each = local.items

            url = "%s"
        }

        rule "must_be_true" "bar" {
            for_each = local.items

            condition = true
            precondition {
                condition = data.http.foo[each.value].response_body != "Mock server content"
                error_message = "Precondition check failed ${each.value}"
            }
        }
    `, server.URL)
	s.dummyFsWithFiles([]string{"test.grept.hcl"}, []string{sampleConfig})

	// Parse the config
	config, err := LoadConfig(NewGreptConfig(), "", "", nil)
	s.NoError(err)

	// Plan the parsed configuration
	_, err = RunGreptPlan(config)
	s.NotNil(err)
	s.Contains(err.Error(), "Precondition check failed item1")
	s.Contains(err.Error(), "Precondition check failed item2")
	s.Contains(err.Error(), "Precondition check failed item3")
}
