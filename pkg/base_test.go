package pkg

import (
	"github.com/prashantv/gostub"
	"github.com/spf13/afero"
)

type testBase struct {
	fs   afero.Fs
	stub *gostub.Stubs
}

func newTestBase() *testBase {
	t := new(testBase)
	t.fs = afero.NewMemMapFs()
	t.stub = gostub.Stub(&FsFactory, func() afero.Fs {
		return t.fs
	})
	return t
}

func (t *testBase) teardown() {
	t.stub.Reset()
}

func (t *testBase) dummyFsWithFiles(fileNames []string, contents []string) {
	for i := range fileNames {
		_ = afero.WriteFile(t.fs, fileNames[i], []byte(contents[i]), 0644)
	}
}
