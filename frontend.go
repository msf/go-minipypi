package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
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
		// must be file, lets find it.
		handleServeS3File(w, path)
	}
}

func handleServeS3File(w http.ResponseWriter, key string) {
	log.Println("FILE", key)
	start := time.Now()

	log.Printf("GetFile:%v", key)
	pathParts := strings.Split(key, "/")
	nparts := len(pathParts)
	if nparts > 2 {
		// requests can come in the form: /packagename/packagename-version-blblablabla.xx
		// the real key in these cases is: /packagename-version-blablalbalbla.xx
		key = "/" + pathParts[nparts-1]
	}
	s3file, err := fetcher.GetFile(key)
	if err != nil {
		desc := fmt.Sprintf("GetFile failed: %v, %v", key, err)
		log.Println(desc)
		http.Error(w, desc, http.StatusNotFound)
		return
	}
	log.Printf("fetcher took: %.3f secs, %v Kb\n",
		time.Since(start).Seconds(), len(s3file.Payload))

	nbytes, err := w.Write(s3file.Payload)
	if err != nil {
		desc := fmt.Sprintf(
			"GetFile(%v) copy to network bombed after %v bytes, %v secs, err:%v",
			key, nbytes, time.Since(start).Seconds(), err)
		log.Println(desc)
		http.Error(w, desc, http.StatusNotFound)
		return
	}
	elapsed := time.Since(start)
	log.Printf("Served file: %v, size: %v, took: %.3f secs",
		key, nbytes, elapsed.Seconds())
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
		desc := fmt.Sprintf("Can't list dir: %v, %v", path, err)
		http.Error(w, desc, http.StatusBadRequest)
		return
	}

	type templateData struct {
		Results []DirListEntry
		Elapsed time.Duration
		BaseDir string
	}

	log.Printf("DirList, Got from fetcher: %v in %.3f secs", len(fileList), time.Since(start).Seconds())
	if err := pypiResultsTemplate.Execute(w, templateData{
		Results: fileList,
		Elapsed: time.Since(start),
		BaseDir: webServerConfigs.BasePath,
	}); err != nil {
		desc := fmt.Sprintf("templating failed: %v", err)
		log.Print(desc)
		http.Error(w, desc, http.StatusInternalServerError)
	}
	log.Printf("DirList:%v in %.3f secs", path, time.Since(start).Seconds())
}

// results just like a pypi server would return
var pypiResultsTemplate = template.Must(template.New("results").Parse(`
<html>
<head/>
<body>
  {{range .Results}}
  <a href="{{.Name}}">{{.Name}}</a>
  {{end}}
</body>
</html>
`))

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
