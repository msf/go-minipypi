package main

import (
	"fmt"
	"html"
	"html/template"
	"io"
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
	// webServer will use this FileFetcher to serve requests.
	fetcher FileFetcher
}

// RunWebServer starts a proxy pypi server that handles HTTP requests by leveraging any implementation of FileFetcher to do the grunt work.
func RunWebServer(cfg WebServerConfigs, fileFetcher FileFetcher) {

	http.HandleFunc(cfg.BasePath, cfg.handleRequest)
	hostPortStr := fmt.Sprintf("%v:%v", cfg.Host, cfg.Port)
	log.Printf("serving on HTTP://%v/%v\n", hostPortStr, cfg.BasePath)
	log.Fatal(http.ListenAndServe(hostPortStr, nil))
}

func (ctx WebServerConfigs) handleRequest(w http.ResponseWriter, req *http.Request) {

	path := html.EscapeString(req.URL.Path)
	if strings.HasSuffix(path, "/") {
		ctx.handleListDir(w, path)
	} else {
		// must be file, lets find it.
		ctx.handleServeS3File(w, path)
	}
}

func (ctx WebServerConfigs) handleServeS3File(w http.ResponseWriter, path string) {
	start := time.Now()

	fileName := handlePypiFileNames(path)
	file, err := ctx.fetcher.GetFile(fileName)
	if err != nil {
		desc := fmt.Sprintf("\"GET %v\" 404 - time: %.3fs, error: %v",
			fileName, time.Since(start).Seconds(), err)
		log.Println(desc)
		http.Error(w, desc, http.StatusNotFound)
		return
	}
	timeFetching := time.Since(start)

	defer file.Payload.Close()
	nbytes, err := io.Copy(w, file.Payload)
	if err != nil {
		desc := fmt.Sprintf("\"GET %v\" 404 - time: %.3fs, network write error: %v",
			fileName, time.Since(start).Seconds(), err)
		log.Println(desc)
		http.Error(w, desc, http.StatusNotFound)
		return
	}
	elapsed := time.Since(start)
	log.Printf("\"GET %v\" 200 - size: %v, fetch: %.3fs, total: %.3fs",
		path, nbytes, timeFetching.Seconds(), elapsed.Seconds())
	return
}

func (ctx WebServerConfigs) handleListDir(w http.ResponseWriter, path string) {
	start := time.Now()

	fileList, err := handlePypiListDir(ctx.fetcher, path)
	if err != nil {
		desc := fmt.Sprintf("\"LISTDIR %v\" 400 - err: %v", path, err)
		http.Error(w, desc, http.StatusBadRequest)
		return
	}

	type templateData struct {
		Results []ListDirEntry
		Elapsed time.Duration
		BaseDir string
	}

	if err := pypiResultsTemplate.Execute(w, templateData{
		Results: fileList,
		Elapsed: time.Since(start),
		BaseDir: ctx.BasePath,
	}); err != nil {
		desc := fmt.Sprintf("\"LISTDIR %v\" 404 - time: %.3fs, template write error: %v",
			path, time.Since(start).Seconds(), err)
		log.Print(desc)
		http.Error(w, desc, http.StatusInternalServerError)
	}
	log.Printf("\"LISTDIR %v\" 200 - items:%v, time: %.3fs",
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
