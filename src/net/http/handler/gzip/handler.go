package gzip

import (
  "errors"
  "fmt"
  "net/http"
  "os"
  "path"
  "path/filepath"
  "strings"
)

// File extends net.http.File to also support writing.
type File interface {
  http.File
  Write(b []byte) (int, error)
}

// Extends net.http.FileSystem to also support opening files to write.
type FileSystem interface {
  http.FileSystem
	OpenFile(name string, flag int, perm os.FileMode) (File, error)
}

// Dir implements FileSystem using the native file system restricted to a
// specific directory.
type Dir string

func (d Dir) getFileName(name string) (string, error) {
	if filepath.Separator != '/' &&
     strings.IndexRune(name, filepath.Separator) >= 0 {
		return "", errors.New("http: invalid character in file path")
	}
  dir := string(d)
  if dir == "" {
    dir = "."
  }
  return filepath.Join(dir, filepath.FromSlash(path.Clean("/" + name))), nil
}

func (d Dir) Open(name string) (http.File, error) {
  file, err := d.getFileName(name)
  if err == nil {
    f, err := os.Open(file)
    if err == nil {
      return f, nil
    }
  }
  return nil, err
}

func (d Dir) OpenFile(name string, flag int, perm os.FileMode) (File, error) {
  file, err := d.getFileName(name)
  if err == nil {
    f, err := os.OpenFile(file, flag, perm)
    if err == nil {
      return f, nil
    }
  }
  return nil, err
}

func FileServer(root FileSystem) http.Handler {
  return nil
}

func Foo() {
  fmt.Println("foo")
}
