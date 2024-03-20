package pkg

import (
	"github.com/ahmetb/go-linq/v3"
	"github.com/stretchr/testify/suite"
	"github.com/zclconf/go-cty/cty"
	"math/big"
	"testing"
)

type localSuite struct {
	suite.Suite
	*testBase
}

func TestLocalSuite(t *testing.T) {
	suite.Run(t, new(localSuite))
}

func (s *localSuite) SetupTest() {
	s.testBase = newTestBase()
}

func (s *localSuite) TearDownTest() {
	s.teardown()
}

func (s *localSuite) TestLocalsBlockShouldBeParsedIntoMultipleLocalBlocks() {
	code := `
locals {
  a = "a"
  b = 1
  c = tolist([for i in [1,2,3,4] : tostring(i) if i%2==0])
}
`
	s.dummyFsWithFiles([]string{"test.grept.hcl"}, []string{code})
	c, err := LoadConfig(NewGreptConfig(), "/", "", nil)
	s.NoError(err)
	_, err = RunGreptPlan(c)
	s.NoError(err)
	var locals []Local
	linq.From(Blocks[Local](c)).OrderBy(func(i interface{}) interface{} {
		return i.(Local).Name()
	}).ToSlice(&locals)
	s.True(AreCtyValuesEqual(cty.StringVal("a"), locals[0].(*LocalBlock).LocalValue))
	s.True(AreCtyValuesEqual(cty.NumberVal(big.NewFloat(1)), locals[1].(*LocalBlock).LocalValue))
	s.True(AreCtyValuesEqual(cty.ListVal([]cty.Value{cty.StringVal("2"), cty.StringVal("4")}), locals[2].(*LocalBlock).LocalValue))
}

func AreCtyValuesEqual(val1, val2 cty.Value) bool {
	// Check if types are equal
	if !val1.Type().Equals(val2.Type()) {
		return false
	}

	// Check value equality based on type
	switch {
	case val1.Type().IsListType() || val1.Type().IsTupleType():
		if val1.LengthInt() != val2.LengthInt() {
			return false
		}
		for it := val1.ElementIterator(); it.Next(); {
			k1, v1 := it.Element()
			v2 := val2.Index(k1)
			if !AreCtyValuesEqual(v1, v2) {
				return false
			}
		}
	case val1.Type().IsMapType() || val1.Type().IsObjectType():
		if val1.LengthInt() != val2.LengthInt() {
			return false
		}
		for it := val1.ElementIterator(); it.Next(); {
			k, v1 := it.Element()
			v2 := val2.Index(k)
			if !AreCtyValuesEqual(v1, v2) {
				return false
			}
		}
	case val1.Type().IsSetType():
		if val1.LengthInt() != val2.LengthInt() {
			return false
		}
		for it := val1.ElementIterator(); it.Next(); {
			_, v1 := it.Element()
			if !val2.HasElement(v1).True() {
				return false
			}
		}
	default:
		// For simple types, we can use the Equal method
		if !val1.Equals(val2).True() {
			return false
		}
	}

	return true
}
