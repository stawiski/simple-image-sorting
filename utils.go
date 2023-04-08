package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
)

var logger *log.Logger

func init() {
	file, err := os.OpenFile("server.log", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		log.Fatalln("Failed to open log file", err)
	}
	logger = log.New(file, "INFO ", log.Ldate|log.Ltime|log.Lshortfile)
}

// parse HTML templates
// pass in a list of file names, and get a template
func parseTemplateFiles(filenames ...string) (t *template.Template) {
	var files []string
	t = template.New("layout")
	for _, file := range filenames {
		files = append(files, fmt.Sprintf("templates/%s.html", file))
	}
	t = template.Must(t.ParseFiles(files...))
	return
}

func generateHTML(writer http.ResponseWriter, data interface{}, filenames ...string) {
	var files []string
	for _, file := range filenames {
		files = append(files, fmt.Sprintf("templates/%s.html", file))
	}

	templates := template.Must(template.ParseFiles(files...))
	templates.ExecuteTemplate(writer, "layout", data)
}

func getCallerFunc() string {
	result := "Unknown"
	// Need to skip two stack frames,
	// as "0" refers to current function,
	// "1" refers to caller function,
	// and we want to get the caller of the caller, hence "2"
	pc, _, _, ok := runtime.Caller(2)
	if ok {
		result = runtime.FuncForPC(pc).Name()
	}

	return result
}

func logInfo(format string, args ...interface{}) {
	logger.SetPrefix("INFO ")
	logger.Printf(format, args...)
}

func logWarning(format string, args ...interface{}) {
	logger.SetPrefix(fmt.Sprintf("WARNING [func: %s] ", getCallerFunc()))
	logger.Printf(format, args...)
}

func logError(format string, args ...interface{}) {
	logger.SetPrefix(fmt.Sprintf("ERROR [func: %s] ", getCallerFunc()))
	logger.Printf(format, args...)
}

func getAllDirectoryPaths(path string) []string {
	var dirs []string
	var err error

	path, err = filepath.Abs(path)
	if err != nil {
		logError("Getting absolute path failed for root: %s, err = %v", path, err)
		return nil
	}

	// Walk the directory tree
	err = filepath.Walk(path, func(walkPath string, info os.FileInfo, err error) error {
		if err != nil {
			logError("Nested walk failed for directory: %s, err = %v", path, err)
			return nil
		}
		if info.IsDir() {
			// Append whole path to the result
			dirs = append(dirs, walkPath)
		}
		return nil
	})

	if err != nil {
		logError("Walk failed for directory: %s, err = %v", path, err)
		return nil
	}

	return dirs
}

func getFilePaths(path string) []string {
	var files []string
	var err error

	path, err = filepath.Abs(path)
	if err != nil {
		logError("Getting absolute path failed for root: %s, err = %v", path, err)
		return nil
	}

	err = filepath.Walk(path, func(walkPath string, info os.FileInfo, err error) error {
		if err != nil {
			logError("Nested walk failed for directory: %s, err = %v", path, err)
			return nil
		}
		if !info.IsDir() {
			files = append(files, walkPath)
		}
		return nil
	})

	if err != nil {
		logError("Walk failed for directory: %s, err = %v", path, err)
		return nil
	}

	return files
}

func stringInSlice(search string, list []string) bool {
	for _, s := range list {
		if s == search {
			return true
		}
	}
	return false
}

func filterFilesByExtensions(extensions []string, files []string) (filteredFiles []string) {
	for _, f := range files {
		for _, ext := range extensions {
			if filepath.Ext(f) == ext {
				filteredFiles = append(filteredFiles, f)
			}
		}
	}

	return
}

func createDirectoryIfNotExists(path string) error {
	// Check if directory exists
	_, err := os.Stat(path)
	if err == nil {
		// Exists
		return nil
	}

	// Create the directory
	err = os.MkdirAll(path, 0755)
	if err != nil {
		return err
	}

	return nil
}
