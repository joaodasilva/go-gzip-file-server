package main

import (
  "flag"
  "io"
  "log"
  "net/http"
  "net/http/handler/gzip"
  "os"
)

const kIndex = `
<html><head><title>Index</title></head>
<body>
<a href="/dir/">/dir/: raw files</a><br>
<a href="/gzip/">/gzip/: gzipped files</a><br>
</body>
</html>`

func showIndex(w http.ResponseWriter, r *http.Request) {
  io.WriteString(w, kIndex)
}

func main() {
  cwd, err := os.Getwd()
  if err != nil {
    log.Fatal(err)
  }

  addr := flag.String("addr", ":8080", "Address to listen at")
  base := flag.String("base", cwd, "Site base path (default: cwd)")
  flag.Parse()

  http.Handle("/", http.HandlerFunc(showIndex))
  http.Handle("/dir/",
              http.StripPrefix("/dir", http.FileServer(http.Dir(*base))))
  http.Handle("/gzip/",
              http.StripPrefix("/gzip", gzip.FileServer(http.Dir(*base))))

  hostaddr := *addr
  if len(hostaddr) == 0 || hostaddr[0] == ':' {
    hostaddr = "localhost" + *addr
  }
  log.Printf("Starting server at http://%v serving from %v\n", hostaddr, *base)
  log.Fatal(http.ListenAndServe(*addr, nil))
}
