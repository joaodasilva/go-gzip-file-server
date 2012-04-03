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
  Create(name string) (File, error)
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

func (d Dir) Create(name string) (File, error) {
  file, err := d.getFileName(name)
  if err == nil {
    f, err := os.Create(file)
    if err == nil {
      return f, nil
    }
  }
  return nil, err
}

type gzipFileHandler struct {
  root FileSystem
}

func (h *gzipFileHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
  // TODO
}

// Similar to net.http.FileServer, but gzips content on the fly and serves it
// encoded, when supported by the client.
func FileServer(root FileSystem) http.Handler {
  return &gzipFileHandler{root}
}

func Foo() {
  fmt.Println("foo")
}
