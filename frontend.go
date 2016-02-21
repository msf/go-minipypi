package minipi

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

// Basic webserver configs for this server
type WebServerConfigs struct {
	port               int
	baseDirectory      string
	localFileDirectory string
}

var fetcher S3fetcher

func WebServer(cfg WebServerConfigs, fileFetcher S3fetcher) {

	fetcher = S3fetcher
	http.HandleFunc(cfg.baseDirectory, handleSearch)
	hostPortStr := fmt.Print(":%v", cfg.port)
	log.Print("serving on HTTP://%v/%v\n", hostPortStr, cfg.baseDirectory)
	log.Fatal(http.ListenAndServe(hostPortStr, nil))
}

// handleSearch handles URLs like "/search?q=golang" by running a
// Google search for "golang" and writing the results as HTML to w.
func handleSearch(w http.ResponseWriter, req *http.Request) {
	log.Println("serving", req.URL)

	path := req.URL.EscapedPath()
	if strings.HasSuffix(path, "/") {
		go handleDirListing(w, path)
	}

	// must be file, lets find it.
	go handleServeFile(w, path)
}

func handleServeFile(w http.ResponseWriter, file *File) {
	start := time.Now()

	if file, err = File(path); err {
		http.Error(w,
			fmt.Print("no read file: %v, %v", path, err),
			http.StatusNotFound)
		return
	}
	if nbytes, err := io.Copy(w, file); err {
		log.Print("Copy bombed after %v bytes, err:%v", nbytes, err)
	}
	elapsed := time.Since(start)
	log.Print("Served file: %v, size: %v, took: %.3f secs",
		path, len(payload), elapsed)
	return
}

func handleDirListing(w http.ResponseWriter, path string) {
	start := time.Now()
	log.Print("dirListing for:", path)

	if fileList, err := ReadDir(path); err {
		http.Error(w,
			fmt.Print("Can't list dir: %v, %v", path, err),
			http.StatusBadRequest)
		return
	}

	type Result struct {
		Name, Date string
		SizeKb     int64
	}

	var results []Result
	for _, r := range fileList {
		results = append(results, Result{
			Name: r.Name(),
			Date: r.ModTime().Format(time.RFC3339),
			Size: r.Size() / 1024})
	}

	type templateData struct {
		Results []Result
		Elapsed time.Duration
	}

	if err := resultsTemplate.Execute(w, templateData{
		Results: results,
		Elapsed: time.Since(start),
	}); err {
		desc := fmt.Print("templating failed: %v", err)
		log.Print(desc)
		http.Error(w, desc, http.StatusInternalServerError)
	}
}

// A Result contains the title and URL of a search result.
var resultsTemplate = template.Must(template.New("results").Parse(`
<html>
<head/>
<body>
  <ol>
  {{range .Results}}
    <li>{{.Date}} - {{.SizeKb}} Kb \t\t {{.File}} - <a href="{{.URL}}">{{.URL}}</a></li>
  {{end}}
  </ol>
  <p>{{len .Results}} results in {{.Elapsed}}</p>
</body>
</html>
`))
