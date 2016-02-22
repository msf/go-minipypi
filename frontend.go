package main

import (
	"fmt"
	"html/template"
	"io"
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
		handleS3DirListing(w, path)
	} else {
		handleS3DirListing(w, path)
		return
		// must be file, lets find it.
		handleServeS3File(w, path)
	}
}

func handleServeS3File(w http.ResponseWriter, key string) {
	log.Println("FILE", key)
	start := time.Now()

	s3file, err := fetcher.GetFile(key)
	if err != nil {
		http.Error(w,
			fmt.Sprintf("no read file: %v, %v", key, err),
			http.StatusNotFound)
		return
	}
	log.Printf("fetcher took: %.3f secs, %v Kb\n",
		time.Since(start).Seconds(), len(s3file.Payload))

	nbytes, err := w.Write(s3file.Payload)
	if err != nil {
		log.Printf("Copy bombed after %v bytes, err:%v", nbytes, err)
	}
	elapsed := time.Since(start)
	log.Printf("Served file: %v, size: %v, took: %.3f secs",
		key, nbytes, elapsed.Seconds())
	return
}

func handleServeFileLocal(w http.ResponseWriter, path string) {
	log.Println("FILE", path)
	start := time.Now()
	file, err := os.Open(path)
	if err != nil {
		http.Error(w,
			fmt.Sprintf("no read file: %v, %v", path, err),
			http.StatusNotFound)
		return
	}
	log.Printf("fetcher took: %.3f secs\n", time.Since(start).Seconds())

	nbytes, err := io.Copy(w, file)
	if err != nil {
		log.Printf("Copy bombed after %v bytes, err:%v", nbytes, err)
	}
	elapsed := time.Since(start)
	log.Printf("Served file: %v, size: %v, took: %.3f secs",
		path, nbytes, elapsed.Seconds())
	return
}

type DirListEntry struct {
	Name, LastModified string
	SizeKb             int64
}

func handleS3DirListing(w http.ResponseWriter, path string) {
	log.Println("DirList", path)
	start := time.Now()

	fileList, err := fetcher.ListBucket(path)
	if err != nil {
		http.Error(w,
			fmt.Sprintf("Can't list dir: %v, %v", path, err),
			http.StatusBadRequest)
		return
	}

	type templateData struct {
		Results []DirListEntry
		Elapsed time.Duration
		BaseDir string
	}

	log.Println("Got from fetcher: %v", fileList)
	if err := resultsTemplate.Execute(w, templateData{
		Results: fileList,
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
  <li>{{.LastModified}} - {{.SizeKb}} Kb - <a href="http://localhost:8080/{{.Name}}">{{.Name}}</a></li>
  {{end}}
  </ol>
  <p>{{len .Results}} results in {{.Elapsed}}</p>
</body>
</html>
`))
