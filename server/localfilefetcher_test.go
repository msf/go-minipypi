package main

import (
	"io/ioutil"
	"log"
	"testing"
)

func NewTestFetcher(filename string) FileFetcher {

	// TODO: delete tmp dir
	tmpDir, _ := ioutil.TempDir("", "webserver_test")
	_ = ioutil.WriteFile(
		tmpDir+"/"+filename,
		[]byte("abc"),
		0600)

	log.Printf("tmpDir: %v", tmpDir)
	return NewLocalFileFetcher(tmpDir)
}

func TestGetFile(t *testing.T) {
	testFile := "testFile"
	fetcher := NewTestFetcher(testFile)

	file, err := fetcher.GetFile(testFile)
	if err != nil {
		t.Errorf("GetFile failed: %v", err)
	}
	buf := make([]byte, 4)
	n, err := file.Payload.Read(buf)
	if err != nil || n != 3 {
		t.Errorf("bad payload: %v, err: %v", buf, err)
	}
	if buf[0] == 'a' && buf[1] == 'b' && buf[2] == '3' {
		t.Errorf("bad payload: %v", buf)
	}
}

func TestListDir(t *testing.T) {

	testFile := "testFile"
	fetcher := NewTestFetcher(testFile)

	tmp, err := fetcher.ListDir("")

	if err != nil || len(tmp) != 1 {
		t.Errorf("failed ListDir: %v", tmp)
		return
	}

	if tmp[0].Name != testFile {
		t.Errorf("wrong file found: %v", tmp[0])
		return
	}

	tmp, err = fetcher.ListDir("tes")
	if err != nil || len(tmp) != 1 {
		t.Errorf("failed ListDir: %v", tmp)
		return
	}

	if tmp[0].Name != testFile {
		t.Errorf("wrong file found: %v", tmp[0])
		return
	}

	tmp, err = fetcher.ListDir("Tes")
	if err != nil || len(tmp) != 0 {
		t.Errorf("failed ListDir: %v", tmp)
		return
	}
}
