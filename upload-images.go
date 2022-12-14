package main

import (
	"archive/zip"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"gopkg.in/gographics/imagick.v3/imagick"
)

func UploadImagesHandler(w http.ResponseWriter, r *http.Request) {
	imagick.Initialize()
	defer imagick.Terminate()

	mw := imagick.NewMagickWand()
	defer mw.Destroy()

	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, MaxUploadSize)
	if err := r.ParseMultipartForm(MaxUploadSize); err != nil {
		http.Error(w, "Upload size is too big. Please upload up to 30MB.", http.StatusBadRequest)
	}

	// フォームで選択された出力フォーマット（WebP or PNG or JPEG）を設定する
	outputImageFormat := r.FormValue("select")
	if outputImageFormat != "webp" && outputImageFormat != "png" && outputImageFormat != "jpeg" {
		http.Error(w, "Output format is not allowed.", http.StatusBadRequest)
		return
	}

	outImageFilePath := fmt.Sprintf("%s", uuid.New().String())
	if err := os.MkdirAll(outImageFilePath, os.ModePerm); err != nil {
		panic(err)
	}
	defer os.RemoveAll(outImageFilePath)

	zipWriter := zip.NewWriter(w)
	defer zipWriter.Close()

	for _, v := range r.MultipartForm.File {
		for _, fh := range v {
			log.Println("uploaded file name: ", fh.Filename)
			f, err := fh.Open()
			if err != nil {
				http.Error(w, fmt.Sprintf("Unexpected error: %s", err.Error()), http.StatusInternalServerError)
			}
			b, err := io.ReadAll(f)
			if err != nil {
				http.Error(w, fmt.Sprintf("Unexpected error: %s", err.Error()), http.StatusInternalServerError)
			}
			mimeType := http.DetectContentType(b)
			if mimeType != "image/jpeg" && mimeType != "image/png" && mimeType != "image/gif" && mimeType != "image/webp" {
				http.Error(w, "MIME type not allowed. Please upload a image.", http.StatusBadRequest)
				return
			}
			if err := mw.ReadImageBlob(b); err != nil {
				log.Fatal("failed at ReadImage.", err)
			}
			if err := mw.SetImageFormat(outputImageFormat); err != nil {
				log.Fatal("failed at SetImageFormat.", err)
			}
			fileName := strings.TrimSuffix(fh.Filename, filepath.Ext(fh.Filename))
			outFilePath := fmt.Sprintf("%s/%s.%s", outImageFilePath, fileName, outputImageFormat)
			if err := mw.WriteImage(outFilePath); err != nil {
				log.Fatal("failed at WriteImage.", err)
			}
			if err := addToZip(outFilePath, zipWriter); err != nil {
				panic(err)
			}
		}
	}
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s.zip\"", "images"))
}
