package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"path"
)

var InputDirectory string = ""
var OutputDirectory string = ""
var DirectoryPathList []string
var ImageList []*Img
var BucketList []Bucket

type stringList []string

func (s *stringList) String() string {
	return fmt.Sprintf("%v", *s)
}

func (s *stringList) Set(value string) error {
	*s = append(*s, value)
	return nil
}

func main() {
	var leftBuckets stringList
	var rightBuckets stringList

	flag.StringVar(&InputDirectory, "input", "", "Input image directory")
	flag.StringVar(&OutputDirectory, "output", "", "Output image directory")
	flag.Var(&leftBuckets, "left", "Left buckets")
	flag.Var(&rightBuckets, "right", "Right buckets")

	flag.Parse()

	if InputDirectory == "" || OutputDirectory == "" || (len(leftBuckets)+len(rightBuckets)) < 2 {
		flag.Usage()
		os.Exit(1)
	}

	fmt.Println("Input directory:", InputDirectory)
	fmt.Println("Output directory:", OutputDirectory)
	fmt.Println("Left buckets:", leftBuckets)
	fmt.Println("Right buckets:", rightBuckets)

	for _, bucket := range leftBuckets {
		BucketList = append(BucketList, Bucket{
			Name:     bucket,
			Position: "left",
		})
	}

	for _, bucket := range rightBuckets {
		BucketList = append(BucketList, Bucket{
			Name:     bucket,
			Position: "right",
		})
	}

	// Create bucket directories if they don't exist
	for _, bucket := range BucketList {
		directory := path.Join(OutputDirectory, bucket.Name)
		err := createDirectoryIfNotExists(directory)
		if err != nil {
			fmt.Println("Error creating bucket directory:", directory)
			os.Exit(1)
		}
	}

	DirectoryPathList = getAllDirectoryPaths(InputDirectory)
	ImageList = getImagesFromDirectoryPaths(DirectoryPathList)

	// handle static assets
	mux := http.NewServeMux()
	files := http.FileServer(http.Dir("public"))
	mux.Handle("/static/", http.StripPrefix("/static/", files))

	//
	// all route patterns matched here
	// route handler functions defined in other files
	//
	mux.HandleFunc("/", routeIndex)
	mux.HandleFunc("/img/get", routeLog(routeImageGet))
	mux.HandleFunc("/img/next", routeLog(routeImageNext))
	mux.HandleFunc("/bucket/set", routeLog(routeBucketSet))
	mux.HandleFunc("/bucket/undo", routeLog(routeBucketUndo))
	mux.HandleFunc("/info/get", routeLog(routeInfoGet))

	// starting up the server
	server := &http.Server{
		Addr:    "0.0.0.0:8080",
		Handler: mux,
	}
	server.ListenAndServe()
}
