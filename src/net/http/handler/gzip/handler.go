package gzip

import (
  "fmt"
  "net/http"
)

func FileServer(root http.FileSystem) http.Handler {
  return http.FileServer(root)
}

func Foo() {
  fmt.Println("foo")
}
