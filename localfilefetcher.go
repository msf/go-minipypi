package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

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
