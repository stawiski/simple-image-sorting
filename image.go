package main

import (
	"errors"
	"fmt"
	"os"
	"path"

	"github.com/google/uuid"
)

type Bucket struct {
	Name     string
	Position string
}

type Img struct {
	Uuid         string
	PathOriginal string
	PathCurrent  string
	Bucket       string
}

func (i Img) String() string {
	return fmt.Sprintf("uuid: %s, path original: %s, path current: %s, bucket: %s", i.Uuid, i.PathOriginal, i.PathCurrent, i.Bucket)
}

func getImagesFromDirectoryPaths(directories []string) (imgList []*Img) {
	fmt.Println(directories)
	for _, path := range directories {
		filePaths := getFilePaths(path)

		imageExtensions := []string{".jpg", ".jpeg", ".png", ".apng", ".svg", ".webp"}
		imagePaths := filterFilesByExtensions(imageExtensions, filePaths)

		for _, imagePath := range imagePaths {
			img := Img{
				Uuid:         uuid.New().String(),
				PathOriginal: imagePath,
				PathCurrent:  imagePath,
				Bucket:       "",
			}
			imgList = append(imgList, &img)
		}
	}

	return imgList
}

func getImageByUuid(imageList []*Img, uuid string) (*Img, error) {
	for _, i := range imageList {
		if i.Uuid == uuid {
			return i, nil
		}
	}

	return nil, errors.New("image not found")
}

func isImageAssigned(image *Img) bool {
	return (*image).Bucket != ""
}

func getFirstUnassignedImage(list []*Img) *Img {
	for _, i := range list {
		if !isImageAssigned(i) {
			return i
		}
	}
	return nil
}

func putImageInBucket(image *Img, bucket string) (err error) {
	if isImageAssigned(image) {
		err = errors.New("image already assigned")
		return
	}

	if image.PathCurrent != image.PathOriginal {
		err = errors.New("image path differs from original path")
		return
	}

	logInfo("putImageInBucket image: %s", image)
	(*image).Bucket = bucket

	// Make new directory for the image
	newDirectory := path.Join(OutputDirectory, bucket)

	err = createDirectoryIfNotExists(newDirectory)
	if err != nil {
		err = errors.New("error creating new directory")
		return
	}

	// Split the path into its directory and file components
	_, filename := path.Split(image.PathCurrent)

	// Join the new directory and file components
	newPath := path.Join(newDirectory, filename)

	logInfo("Renaming \"%s\" to \"%s\"", image.PathCurrent, newPath)

	err = os.Rename(image.PathCurrent, newPath)
	if err != nil {
		err = errors.New("error moving file")
		return
	}

	(*image).PathCurrent = newPath
	return
}

func revertImageBucket(image *Img) (err error) {
	// Reset bucket
	(*image).Bucket = ""

	logInfo("Renaming \"%s\" to \"%s\"", (*image).PathCurrent, (*image).PathOriginal)
	err = os.Rename((*image).PathCurrent, (*image).PathOriginal)
	if err != nil {
		err = errors.New("error moving file")
		return
	}

	(*image).PathCurrent = image.PathOriginal
	return
}
