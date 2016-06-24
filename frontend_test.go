package main

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

var (
	server *httptest.Server

	testFetcher FileFetcher
)

func init() {

	cfgs := WebServerConfigs{
		fetcher: NewTestFetcher("psycopg2-2.5.3-cp27-none-linux_x86_64.whl"),
	}

	server = httptest.NewServer(http.HandlerFunc(cfgs.handleRequest))
}

func TestDirList(t *testing.T) {
	resp, err := http.Get(server.URL + "/psycopg2/")

	if err != nil {
		t.Fatal("failed to handle DirList, error:", err)
	}
	if resp.StatusCode != 200 {
		t.Fatal("bad statuscode", resp)
	}
	defer resp.Body.Close()
	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal("failed to handle DirList, error:", err)
	}
	t.Logf("body: %v", string(buf))
}

func TestFetchFile(t *testing.T) {
	resp, err := http.Get(server.URL + "/psycopg2/psycopg2-2.5.3-cp27-none-linux_x86_64.whl")
	if err != nil {
		t.Fatal("failed to handle FetchFile, error:", err)
	}
	if resp.StatusCode != 200 {
		t.Fatal("bad statuscode", resp)
	}
	defer resp.Body.Close()
	buf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal("failed to handle DirList, error:", err)
	}
	t.Logf("body: %v", string(buf))
}
