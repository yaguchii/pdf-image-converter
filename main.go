package main

import (
	"embed"
	"fmt"
	"log"
	"net/http"
	"time"
)

//go:embed asset/*
var asset embed.FS

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", serveUploadPdfHTML)
	mux.HandleFunc("/upload", uploadPdfHandler)

	mux.HandleFunc("/multi", serveUploadImagesHTML)
	mux.HandleFunc("/upload-multi", uploadImagesHandler)

	mux.Handle("/asset/", http.FileServer(http.FS(asset)))

	srv := &http.Server{
		ReadTimeout: 1 * time.Minute,
		Handler:     mux,
		Addr:        ":8080",
	}
	fmt.Println("Server is running on port 8080.")
	if err := srv.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

func serveUploadPdfHTML(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html")
	http.ServeFile(w, r, "html/index.html")
}

func serveUploadImagesHTML(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html")
	http.ServeFile(w, r, "html/index-multi.html")
}

func uploadPdfHandler(w http.ResponseWriter, r *http.Request) {
	UploadPDFHandler(w, r)
}

func uploadImagesHandler(w http.ResponseWriter, r *http.Request) {
	UploadImagesHandler(w, r)
}
