package mocks

import (
	"os"
	"time"

	"github.com/spf13/afero"
)

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

type MockFsWithError struct {
	afero.Fs
	ErrToReturn error
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (m *MockFsWithError) Stat(name string) (os.FileInfo, error) {
	if m.ErrToReturn != nil {
		return nil, m.ErrToReturn
	}

	return m.Fs.Stat(name)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (m *MockFsWithError) Name() string {
	return "MockFsWithError"
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (m *MockFsWithError) Create(name string) (afero.File, error) {
	if m.ErrToReturn != nil {
		return nil, m.ErrToReturn
	}

	return m.Fs.Create(name)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (m *MockFsWithError) Mkdir(name string, perm os.FileMode) error {
	if m.ErrToReturn != nil {
		return m.ErrToReturn
	}

	return m.Fs.Mkdir(name, perm)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (m *MockFsWithError) MkdirAll(path string, perm os.FileMode) error {
	if m.ErrToReturn != nil {
		return m.ErrToReturn
	}

	return m.Fs.MkdirAll(path, perm)
}

func (m *MockFsWithError) Open(name string) (afero.File, error) {
	if m.ErrToReturn != nil {
		return nil, m.ErrToReturn
	}

	return m.Fs.Open(name)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (m *MockFsWithError) OpenFile(name string, flag int, perm os.FileMode) (afero.File, error) {
	if m.ErrToReturn != nil {
		return nil, m.ErrToReturn
	}

	return m.Fs.OpenFile(name, flag, perm)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (m *MockFsWithError) Remove(name string) error {
	if m.ErrToReturn != nil {
		return m.ErrToReturn
	}

	return m.Fs.Remove(name)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (m *MockFsWithError) RemoveAll(path string) error {
	if m.ErrToReturn != nil {
		return m.ErrToReturn
	}

	return m.Fs.RemoveAll(path)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (m *MockFsWithError) Rename(oldname, newname string) error {
	if m.ErrToReturn != nil {
		return m.ErrToReturn
	}

	return m.Fs.Rename(oldname, newname)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (m *MockFsWithError) Chmod(name string, mode os.FileMode) error {
	if m.ErrToReturn != nil {
		return m.ErrToReturn
	}

	return m.Fs.Chmod(name, mode)
}

// ~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

func (m *MockFsWithError) Chtimes(name string, atime time.Time, mtime time.Time) error {
	if m.ErrToReturn != nil {
		return m.ErrToReturn
	}

	return m.Fs.Chtimes(name, atime, mtime)
}
