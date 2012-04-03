package gzip

import (
  "fmt"
  "html"
  "io"
  "mime"
  "net/http"
  "os"
  "path"
  "path/filepath"
  "strings"
  "time"
)

type gzipFileHandler struct {
  fs http.FileSystem
}

func (h *gzipFileHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
  p := r.URL.Path
  if !strings.HasPrefix(p, "/") {
    p = "/" + p
    r.URL.Path = p
  }
  serveFile(w, r, h.fs, path.Clean(p), true)
}

// Similar to net.http.FileServer, but serves <file>.gz instead of <file> if
// it exists, has a later modification time, and the request supports gzip
// encoding. Also serves the .gz file if the original doesn't exist.
func FileServer(root http.FileSystem) http.Handler {
  return &gzipFileHandler{root}
}

// Similar to net.http.ServeFile, but serves <file>.gz instead of <file> if
// it exists, has a later modification time, and the request supports gzip
// encoding. Also serves the .gz file if the original doesn't exist.
func ServeFile(w http.ResponseWriter, r *http.Request, name string) {
  dir, file := filepath.Split(name)
  serveFile(w, r, http.Dir(dir), file, false)
}

func serveFile(w http.ResponseWriter, r *http.Request, fs http.FileSystem,
               name string, redirect bool) {
  const kIndex = "/index.html"

  if strings.HasSuffix(r.URL.Path, kIndex) {
    localRedirect(w, r, "./")
    return
  }

  modtime := time.Time{}
  isGzip := false

  file, stat := open(fs, name)
  if file != nil {
    defer file.Close()
    if maybeRedirect(w, r, stat.IsDir()) {
      return
    }
    if stat.IsDir() {
      // Use index.html, if present.
      index := name + kIndex
      ifile, istat := open(fs, index)
      if ifile != nil {
        defer ifile.Close()
        name = index
        modtime = istat.ModTime()
        file = ifile
        stat = istat
      }
    } else {
      name = stat.Name()
      modtime = stat.ModTime()
    }
  }

  if supportsGzip(r) && canTryGzip(stat) {
    zfile, zstat := open(fs, name + ".gz")
    if zfile != nil {
      defer zfile.Close()
      if shouldUseGzip(stat, zstat) {
        if file == nil {
          name = zstat.Name()
          name = name[:len(name) - 3]
          modtime = zstat.ModTime()
        }
        file = zfile
        stat = zstat
        isGzip = true
      }
    }
  }

  if file == nil {
    http.NotFound(w, r)
    return
  }

  if stat.IsDir() {
    dirList(w, file)
  } else {
    if isGzip {
      setContentType(w, name, file)
      w.Header().Set("Content-Encoding", "gzip")
    }
    http.ServeContent(w, r, name, modtime, file)
  }
}

// localRedirect gives a Moved Permanently response.
// It does not convert relative paths to absolute paths like Redirect does,
// which would be a problem when using StripPrefix.
func localRedirect(w http.ResponseWriter, r *http.Request, newPath string) {
	if q := r.URL.RawQuery; q != "" {
		newPath += "?" + q
	}
	w.Header().Set("Location", newPath)
	w.WriteHeader(http.StatusMovedPermanently)
}

func maybeRedirect(w http.ResponseWriter, r *http.Request, isDir bool) bool {
  // redirect to canonical path: / at end of directory url
  // r.URL.Path always begins with /
  url := r.URL.Path
  if isDir {
    if url[len(url)-1] != '/' {
      localRedirect(w, r, path.Base(url) + "/")
      return true
    }
  } else {
    if url[len(url)-1] == '/' {
      localRedirect(w, r, "../" + path.Base(url))
      return true
    }
  }
  return false
}

func supportsGzip(r *http.Request) bool {
  for _, encodings := range r.Header["Accept-Encoding"] {
    for _, encoding := range strings.Split(encodings, ",") {
      if encoding == "gzip" {
        return true
      }
    }
  }
  return false
}

func canTryGzip(stat os.FileInfo) bool {
  if stat == nil {
    return true
  }
  if stat.IsDir() {
    return false
  }
  if strings.HasSuffix(strings.ToLower(stat.Name()), ".gz") {
    return false
  }
  return true
}

func shouldUseGzip(stat, zstat os.FileInfo) bool {
  if stat == nil {
    return true
  }
  return !stat.ModTime().After(zstat.ModTime())
}

func dirList(w http.ResponseWriter, f http.File) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, "<pre>\n")
	for {
		dirs, err := f.Readdir(100)
		if err != nil || len(dirs) == 0 {
			break
		}
		for _, d := range dirs {
			name := d.Name()
			if d.IsDir() {
				name += "/"
			}
      escaped := html.EscapeString(name)
			fmt.Fprintf(w, "<a href=\"%s\">%s</a>\n", escaped, escaped)
		}
	}
	fmt.Fprintf(w, "</pre>\n")
}

func setContentType(w http.ResponseWriter, name string, file http.File) {
  t := mime.TypeByExtension(filepath.Ext(name))
  if t == "" {
    var buffer [512]byte
    n, _ := io.ReadFull(file, buffer[:])
    t = http.DetectContentType(buffer[:n])
    if _, err := file.Seek(0, os.SEEK_SET); err != nil {
      http.Error(w, "Can't seek", http.StatusInternalServerError)
      return
    }
  }
  w.Header().Set("Content-Type", t)
}

func open(fs http.FileSystem, name string) (http.File, os.FileInfo) {
  f, err := fs.Open(name)
  if err != nil {
    return nil, nil
  }
  s, err := f.Stat()
  if err != nil {
    f.Close()
    return nil, nil
  }
  return f, s
}
