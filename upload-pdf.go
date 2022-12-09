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

const (
	KB            = 1000
	MB            = 1000 * KB
	MaxUploadSize = 30 * MB
)

func UploadPDFHandler(w http.ResponseWriter, r *http.Request) {
	imagick.Initialize()
	defer imagick.Terminate()

	mw := imagick.NewMagickWand()
	defer mw.Destroy()

	if r.Method != "POST" {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, MaxUploadSize)
	if err := r.ParseMultipartForm(MaxUploadSize); err != nil {
		http.Error(w, "Upload size is too big. Please upload up to 30MB.", http.StatusBadRequest)
	}

	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	log.Println("Uploading filename:", fileHeader.Filename)

	outImageFilePath := fmt.Sprintf("%s", uuid.New().String())
	if err := os.MkdirAll(outImageFilePath, os.ModePerm); err != nil {
		panic(err)
	}
	defer os.RemoveAll(outImageFilePath)

	originalFile := fmt.Sprintf("%s/%s", outImageFilePath, fileHeader.Filename)
	out, _ := os.Create(originalFile)
	io.Copy(out, file)

	bytes, err := os.ReadFile(originalFile)
	if err != nil {
		log.Fatal(err)
	}
	mimeType := http.DetectContentType(bytes)
	if mimeType != "application/pdf" {
		http.Error(w, "MIME type not allowed. Please upload a PDF.", http.StatusBadRequest)
		return
	}

	// 変換元のPDFを読み込む
	if err := mw.ReadImage(originalFile); err != nil {
		log.Fatal("failed at ReadImage.", err)
	}

	// ページ数を取得する
	numberOfImages := mw.GetNumberImages()
	log.Println("number of images: ", numberOfImages)

	// フォームで選択された出力フォーマット（WebP/PNG/JPEG）に設定する
	outputImageFormat := r.FormValue("select")
	if outputImageFormat != "webp" && outputImageFormat != "png" && outputImageFormat != "jpeg" {
		http.Error(w, "Output format is not allowed.", http.StatusBadRequest)
		return
	}
	if err := mw.SetImageFormat(outputImageFormat); err != nil {
		log.Fatal("failed at SetImageFormat.", err)
	}

	zipWriter := zip.NewWriter(w)
	defer zipWriter.Close()

	// １ページずつ変換して出力する
	for i := 0; i < int(numberOfImages); i++ {
		// ページ番号を設定する
		if ret := mw.SetIteratorIndex(i); !ret {
			break
		}

		fileName := strings.TrimSuffix(fileHeader.Filename, filepath.Ext(fileHeader.Filename))
		outFilePath := fmt.Sprintf("%s/%d_%s.%s", outImageFilePath, i+1, fileName, outputImageFormat)
		// 画像を出力する
		if err := mw.WriteImage(outFilePath); err != nil {
			log.Fatal("failed at WriteImage.", err)
		}

		// 画像をZipに追加する
		if err := addToZip(outFilePath, zipWriter); err != nil {
			panic(err)
		}
	}

	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s.zip\"", "images"))
}

func addToZip(filename string, zipWriter *zip.Writer) error {
	src, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer src.Close()

	writer, err := zipWriter.Create(filename)
	if err != nil {
		return err
	}
	_, err = io.Copy(writer, src)
	if err != nil {
		return err
	}
	return nil
}
