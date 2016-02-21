package main

import (
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

// Basic webserver configs for this server
type WebServerConfigs struct {
	Port               int
	BasePath           string
	LocalFileDirectory string
}

var fetcher S3fetcher
var webServerConfigs WebServerConfigs

func WebServer(cfg WebServerConfigs, fileFetcher S3fetcher) {

	fetcher = fileFetcher
	webServerConfigs = cfg
	http.HandleFunc(cfg.BasePath, handleRequest)
	hostPortStr := fmt.Sprintf(":%v", cfg.Port)
	log.Printf("serving on HTTP://%v/%v\n", hostPortStr, cfg.BasePath)
	log.Fatal(http.ListenAndServe(hostPortStr, nil))
}

func handleRequest(w http.ResponseWriter, req *http.Request) {
	log.Println("serving", req.URL)

	path := req.URL.EscapedPath()
	if strings.HasSuffix(path, "/") {
		handleDirListing(w,
			webServerConfigs.LocalFileDirectory+path)
	} else {
		// must be file, lets find it.
		handleServeFile(w, webServerConfigs.LocalFileDirectory+path)
	}
}

func handleServeFile(w http.ResponseWriter, path string) {
	log.Println("FILE", path)
	start := time.Now()

	file, err := os.Open(path)
	if err != nil {
		http.Error(w,
			fmt.Sprintf("no read file: %v, %v", path, err),
			http.StatusNotFound)
		return
	}

	nbytes, err := io.Copy(w, file)
	if err != nil {
		log.Printf("Copy bombed after %v bytes, err:%v", nbytes, err)
	}
	elapsed := time.Since(start)
	log.Printf("Served file: %v, size: %v, took: %.3f secs",
		path, nbytes, elapsed)
	return
}

func handleDirListing(w http.ResponseWriter, path string) {
	log.Println("DirList", path)
	start := time.Now()

	fileList, err := ioutil.ReadDir(path)
	if err != nil {
		http.Error(w,
			fmt.Sprintf("Can't list dir: %v, %v", path, err),
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
			Name:   r.Name(),
			Date:   r.ModTime().Format(time.RFC3339),
			SizeKb: r.Size() / 1024})
	}

	type templateData struct {
		Results []Result
		Elapsed time.Duration
		BaseDir string
	}

	if err := resultsTemplate.Execute(w, templateData{
		Results: results,
		Elapsed: time.Since(start),
		BaseDir: webServerConfigs.BasePath,
	}); err != nil {
		desc := fmt.Sprintf("templating failed: %v", err)
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
  <li>{{.Date}} - {{.SizeKb}} Kb - <a href="http://localhost:8080/{{.Name}}">{{.Name}}</a></li>
  {{end}}
  </ol>
  <p>{{len .Results}} results in {{.Elapsed}}</p>
</body>
</html>
`))
