package pkg

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"fmt"
	"github.com/zclconf/go-cty/cty"
	"hash"

	"github.com/spf13/afero"
)

var _ Rule = &FileHashRule{}

type FileHashRule struct {
	*BaseBlock
	*BaseRule
	Glob               string   `hcl:"glob"`
	Hash               string   `hcl:"hash"`
	Algorithm          string   `hcl:"algorithm,optional" default:"sha1"`
	FailOnHashMismatch bool     `hcl:"fail_on_hash_mismatch,optional"`
	HashMismatchFiles  []string `attribute:"hash_mismatch_files"`
}

func (fhr *FileHashRule) Type() string {
	return "file_hash"
}

func (fhr *FileHashRule) Values() map[string]cty.Value {
	return map[string]cty.Value{
		"glob":                  ToCtyValue(fhr.Glob),
		"hash":                  ToCtyValue(fhr.Hash),
		"algorithm":             ToCtyValue(fhr.Algorithm),
		"fail_on_hash_mismatch": ToCtyValue(fhr.FailOnHashMismatch),
		"hash_mismatch_files":   ToCtyValue(fhr.HashMismatchFiles),
	}
}

func (fhr *FileHashRule) ExecuteDuringPlan() error {
	// Use Glob to find files matching the path pattern
	fs := FsFactory()
	files, err := afero.Glob(fs, fhr.Glob)
	if err != nil {
		return err
	}

	if len(files) == 0 {
		fhr.setCheckError(fmt.Errorf("no files match path pattern: %s", fhr.Glob))
		return nil
	}
	matchFound := false

	for _, file := range files {
		fileData, err := afero.ReadFile(fs, file)
		if err != nil {
			return err
		}

		// Calculate the hash of the file data
		var h hash.Hash
		switch fhr.Algorithm {
		case "md5":
			h = md5.New()
		case "sha256":
			h = sha256.New()
		case "sha512":
			h = sha512.New()
		case "sha1":
			fallthrough
		default: // Default to sha1
			h = sha1.New()
		}
		h.Write(fileData)
		computedHash := fmt.Sprintf("%x", h.Sum(nil))

		if computedHash == fhr.Hash {
			matchFound = true
			continue
		}
		fhr.HashMismatchFiles = append(fhr.HashMismatchFiles, file)
	}

	if !fhr.FailOnHashMismatch && matchFound {
		return nil
	}

	if len(fhr.HashMismatchFiles) == 0 {
		return nil
	}

	fhr.setCheckError(fmt.Errorf("file with glob %s and  different hash than %s found", fhr.Glob, fhr.Hash))
	return nil
}

func (fhr *FileHashRule) Validate() error {
	if fhr.Glob == "" {
		return fmt.Errorf("glob is required")
	}

	if fhr.Hash == "" {
		return fmt.Errorf("hash is required")
	}

	if fhr.Algorithm != "" {
		switch fhr.Algorithm {
		case "md5", "sha1", "sha256", "sha512":
			// valid
		default:
			return fmt.Errorf("invalid algorithm: %s", fhr.Algorithm)
		}
	}

	return nil
}
