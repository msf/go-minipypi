package main

import (
	"fmt"
	"html"
	"html/template"
	"log"
	"net/http"
	"strings"
	"time"
)

// WebServerConfigs holds configs for this server
type WebServerConfigs struct {
	Host     string
	Port     int
	BasePath string
}

// FileFetcher interface that the frontend uses to serve requests.
type FileFetcher interface {
	GetFile(path string) (*S3File, error)
	ListBucket(bucketName string) ([]DirListEntry, error)
}

var fetcher FileFetcher
var webServerConfigs WebServerConfigs

// WebServer is a proxy pypi server that handles HTTP requests by leveraging any implementation of FileFetcher to do the grunt work.
func WebServer(cfg WebServerConfigs, fileFetcher FileFetcher) {

	fetcher = fileFetcher
	webServerConfigs = cfg
	http.HandleFunc(cfg.BasePath, handleRequest)
	hostPortStr := fmt.Sprintf("%v:%v", cfg.Host, cfg.Port)
	log.Printf("serving on HTTP://%v/%v\n", hostPortStr, cfg.BasePath)
	log.Fatal(http.ListenAndServe(hostPortStr, nil))
}

func handleRequest(w http.ResponseWriter, req *http.Request) {

	path := html.EscapeString(req.URL.Path)
	if strings.HasSuffix(path, "/") {
		handleS3DirListing(w, path)
	} else {
		// must be file, lets find it.
		handleServeS3File(w, path)
	}
}

func handleServeS3File(w http.ResponseWriter, path string) {
	start := time.Now()
	key := path

	pathParts := strings.Split(key, "/")
	nparts := len(pathParts)
	if nparts > 2 {
		// requests can come in the form: /packagename/packagename-version-blblablabla.xx
		// the real key in these cases is: /packagename-version-blablalbalbla.xx
		key = "/" + pathParts[nparts-1]
	}
	s3file, err := fetcher.GetFile(key)
	if err != nil {
		desc := fmt.Sprintf("\"GET %v\" 404 - time: %.3f secs, error: %v",
			key, time.Since(start).Seconds(), err)
		log.Println(desc)
		http.Error(w, desc, http.StatusNotFound)
		return
	}
	timeFetching := time.Since(start)

	nbytes, err := w.Write(s3file.Payload)
	if err != nil {
		desc := fmt.Sprintf("\"GET %v\" 404 - time: %.3f secs, network write error: %v",
			key, time.Since(start).Seconds(), err)
		log.Println(desc)
		http.Error(w, desc, http.StatusNotFound)
		return
	}
	elapsed := time.Since(start)
	log.Printf("\"GET %v\" 200 - size: %v, fetch: %.3f, total: %.3f secs",
		path, nbytes, timeFetching.Seconds(), elapsed.Seconds())
	return
}

type DirListEntry struct {
	Name, LastModified string
	SizeKb             int64
}

func handleS3DirListing(w http.ResponseWriter, path string) {
	start := time.Now()

	fileList, err := fetcher.ListBucket(path)
	if err != nil {
		desc := fmt.Sprintf("\"DIRLIST %v\" 400 - err: %v", path, err)
		http.Error(w, desc, http.StatusBadRequest)
		return
	}

	type templateData struct {
		Results []DirListEntry
		Elapsed time.Duration
		BaseDir string
	}

	if err := pypiResultsTemplate.Execute(w, templateData{
		Results: fileList,
		Elapsed: time.Since(start),
		BaseDir: webServerConfigs.BasePath,
	}); err != nil {
		desc := fmt.Sprintf("\"DIRLIST %v\" 404 - time: %.3f secs, template write error: %v",
			path, time.Since(start).Seconds(), err)
		log.Print(desc)
		http.Error(w, desc, http.StatusInternalServerError)
	}
	log.Printf("\"DIRLIST %v\" 200 - items:%v, time: %.3f secs",
		path, len(fileList), time.Since(start).Seconds())
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
