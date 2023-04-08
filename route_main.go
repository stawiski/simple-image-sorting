package main

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"path/filepath"
	"reflect"
	"runtime"
)

func routeLog(h http.HandlerFunc) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		funcName := runtime.FuncForPC(reflect.ValueOf(h).Pointer()).Name()
		logInfo("Route handler called: %s", funcName)
		h(writer, request)
	}
}

func routeIndex(writer http.ResponseWriter, request *http.Request) {
	generateHTML(writer, nil, "layout")
}

// Return raw image file by it's UUID
func routeImageGet(writer http.ResponseWriter, request *http.Request) {
	vals := request.URL.Query()
	uuid := vals.Get("uuid")

	if uuid == "" {
		logError("UUID is empty")
		defer writer.WriteHeader(http.StatusBadRequest)
		return
	}

	img, err := getImageByUuid(ImageList, uuid)
	if err != nil {
		logError("Image not found")
		defer writer.WriteHeader(http.StatusNotFound)
		return
	}

	logInfo("routeImageGet, image: %s", img)

	imgPath := img.PathCurrent

	// Read the image file into memory.
	imageData, err := ioutil.ReadFile(imgPath)
	if err != nil {
		logError("File not found (path: %s)", imgPath)
		defer writer.WriteHeader(http.StatusNotFound)
		return
	}

	// Get the file extension to determine the content type
	contentType := ""
	fileExtension := filepath.Ext(imgPath)
	switch fileExtension {
	case ".jpg", ".jpeg":
		contentType = "image/jpeg"
	case ".png":
		contentType = "image/png"
	case ".webp":
		contentType = "image/webp"
	default:
		// File extension is not supported
		defer writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Set the content type header
	writer.Header().Set("Content-Type", contentType)

	// Write the image to the response writer.
	io.Copy(writer, bytes.NewReader(imageData))
}

// Return next unassigned image UUID (returns empty json if there is no next image)
func routeImageNext(writer http.ResponseWriter, request *http.Request) {
	type NextImageMessage struct {
		Uuid string `json:"uuid"`
	}

	img := getFirstUnassignedImage(ImageList)

	logInfo("image %s", img)

	message := NextImageMessage{}
	if img != nil {
		message.Uuid = img.Uuid
	}

	writer.Header().Set("Content-Type", "application/json")
	json.NewEncoder(writer).Encode(message)
}

// Set a given image bucket (takes image UUID and bucket name)
func routeBucketSet(writer http.ResponseWriter, request *http.Request) {
	err := request.ParseForm()
	if err != nil {
		logError("Cannot parse POST form: %s", err.Error())
		defer writer.WriteHeader(http.StatusBadRequest)
		return
	}
	uuid := request.PostFormValue("uuid")
	bucket := request.PostFormValue("bucket")

	if uuid == "" {
		logError("UUID is empty")
		defer writer.WriteHeader(http.StatusBadRequest)
		return
	}

	if bucket == "" {
		logError("Bucket is empty")
		defer writer.WriteHeader(http.StatusBadRequest)
		return
	}

	bucketFound := false

	for _, b := range BucketList {
		if bucket == b.Name {
			bucketFound = true
			break
		}
	}

	if !bucketFound {
		logError("Bucket is invalid (%s)", bucket)
		defer writer.WriteHeader(http.StatusBadRequest)
		return
	}

	img, err := getImageByUuid(ImageList, uuid)
	if err != nil {
		logError("Image not found: %s", err.Error())
		defer writer.WriteHeader(http.StatusNotFound)
		return
	}

	logInfo("uuid %s", uuid)
	logInfo("image %s", img)

	err = putImageInBucket(img, bucket)
	if err != nil {
		logError("Couldn't put image in a bucket: %s", err.Error())
		defer writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	logInfo("image after %s", img)

	writer.Write([]byte("ok"))
}

// Undo given image bucket (takes image UUID), it will put the image back to the original location
func routeBucketUndo(writer http.ResponseWriter, request *http.Request) {
	err := request.ParseForm()
	if err != nil {
		logError("Cannot parse POST form: %s", err.Error())
		defer writer.WriteHeader(http.StatusBadRequest)
		return
	}
	uuid := request.PostFormValue("uuid")

	if uuid == "" {
		logError("UUID is empty")
		defer writer.WriteHeader(http.StatusBadRequest)
		return
	}

	img, err := getImageByUuid(ImageList, uuid)
	if err != nil {
		logError("Image not found: %s", err.Error())
		defer writer.WriteHeader(http.StatusNotFound)
		return
	}

	logInfo("uuid %s", uuid)
	logInfo("image %s", img)

	err = revertImageBucket(img)
	if err != nil {
		logError("Couldn't revert image bucket: %s", err.Error())
		defer writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	logInfo("image after %s", img)

	writer.Write([]byte("ok"))
}

// Get info about buckets. Returns a json with bucket structure and names.
func routeInfoGet(writer http.ResponseWriter, request *http.Request) {
	writer.Header().Set("Content-Type", "application/json")

	// Create temporary buffers for left and right buckets
	leftBuckets := make([]string, 0)
	rightBuckets := make([]string, 0)

	for _, bucket := range BucketList {
		if bucket.Position == "left" {
			leftBuckets = append(leftBuckets, bucket.Name)
		} else if bucket.Position == "right" {
			rightBuckets = append(rightBuckets, bucket.Name)
		}
	}

	// Create a map to return a json
	jsonData := map[string]interface{}{
		"left":  leftBuckets,
		"right": rightBuckets,
	}

	// Encode the map as JSON and write it to the response writer
	response, err := json.Marshal(jsonData)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
		return
	}
	writer.Write(response)
}
