package storage

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type FileSystemStorage interface {
	Save(path string, data io.Reader) (location string, err error)
	Get(path string) (file *os.File, err error)
	Move(oldPath, newPath string) (location string, err error)
}

type TmpFileSystem struct {
	baseDirectory string
}

func NewTmpFileSystem() FileSystemStorage {
	tmpDir := "/tmp"

	return &TmpFileSystem{
		baseDirectory: tmpDir,
	}
}

func (fs *TmpFileSystem) Save(path string, data io.Reader) (location string, err error) {
	directory := filepath.Dir(path)
	name := filepath.Base(path)

	baseDir := filepath.Join(fs.baseDirectory, directory)

	os.MkdirAll(baseDir, os.ModePerm)

	fp := filepath.Join(baseDir, name)

	file, err := os.Create(fp)

	if err != nil {
		return "", err
	}

	content, err := ioutil.ReadAll(data)

	if err != nil {
		return "", err
	}

	_, err = file.Write(content)

	if err != nil {
		return "", err
	}

	return file.Name(), nil
}

func (fs *TmpFileSystem) Get(path string) (file *os.File, err error) {
	realPath := path

	if !strings.HasPrefix(path, fs.baseDirectory) {
		realPath = filepath.Join(fs.baseDirectory, path)
	}

	return os.Open(realPath)
}

func (fs *TmpFileSystem) Move(oldPath, newPath string) (string, error) {
	realNewPath := newPath
	realOldPath := oldPath

	if !strings.HasPrefix(newPath, fs.baseDirectory) {
		realNewPath = filepath.Join(fs.baseDirectory, newPath)
	}

	if !strings.HasPrefix(oldPath, fs.baseDirectory) {
		realOldPath = filepath.Join(fs.baseDirectory, oldPath)
	}

	err := os.Rename(realOldPath, realNewPath)

	if err != nil {
		return "", err
	}

	return newPath, nil
}
