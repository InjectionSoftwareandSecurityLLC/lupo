package core

import (
	"bufio"
	"encoding/base64"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type downloadOutputKey struct {
	sessionID int
	filename  string
}

var pendingDownloadOutputs = struct {
	sync.Mutex
	paths map[downloadOutputKey][]string
}{paths: make(map[downloadOutputKey][]string)}

// UploadFile - Reads a file to be uploaded and converts it to base64 to pass to the server as a response for the session
func UploadFile(file string) string {

	WarningColorBold.Println("\nUploading file: " + file + "...")

	// Open file on disk.
	f, err := os.Open(file)

	if err != nil {
		ErrorColorUnderline.Println("could not read file: " + file)
		return ""
	}

	reader := bufio.NewReader(f)
	content, _ := ioutil.ReadAll(reader)

	// Encode as base64.
	encoded := base64.StdEncoding.EncodeToString(content)

	return encoded
}

// DownloadFile - Reads a base64 encoded string and writes it out to a local file
func DownloadFile(filename string, fileb64 string) {
	localName := filepath.Base(strings.ReplaceAll(filename, `\`, "/"))
	DownloadFileAs(localName, fileb64)
}

// DownloadFileAs writes a downloaded file to the specified local path.
func DownloadFileAs(filename string, fileb64 string) {
	WarningColorBold.Println("\nDownloading file: " + filename + "...")

	file, err := base64.StdEncoding.DecodeString(fileb64)
	if err != nil {
		var decoded string
		decoded, err = url.QueryUnescape(fileb64)
		if err == nil {
			file, err = base64.StdEncoding.DecodeString(decoded)
		}
	}
	if err != nil {
		WarningColorBold.Println("could not base64 decode downloaded file, the raw string will be written instead...")
		f, err := os.Create(filename)
		if err != nil {
			ErrorColorUnderline.Println("error: could not create file: " + filename)
			return
		}
		defer f.Close()

		if _, err := f.Write([]byte(fileb64)); err != nil {
			ErrorColorUnderline.Println("error: could not write data to file: " + filename)
			return
		}
		if err := f.Sync(); err != nil {
			ErrorColorUnderline.Println("error: sync error, there was a problem writing data to file: " + filename)
			return
		}
	} else {
		f, err := os.Create(filename)
		if err != nil {
			ErrorColorUnderline.Println("error: could not create file: " + filename)
			return
		}
		defer f.Close()

		if _, err := f.Write(file); err != nil {
			ErrorColorUnderline.Println("error: could not write data to file: " + filename)
			return
		}
		if err := f.Sync(); err != nil {
			ErrorColorUnderline.Println("error: sync error, there was a problem writing data to file: " + filename)
			return
		}
	}

	SuccessColorBold.Println("Downloaded file: " + filename + "!")

}

func SetPendingDownloadOutput(sessionID int, remoteFilename string, outputPath string) {
	if outputPath == "" {
		return
	}

	key := downloadOutputKey{
		sessionID: sessionID,
		filename:  strings.ReplaceAll(remoteFilename, `\`, "/"),
	}

	pendingDownloadOutputs.Lock()
	pendingDownloadOutputs.paths[key] = append(pendingDownloadOutputs.paths[key], outputPath)
	pendingDownloadOutputs.Unlock()
}

func ConsumePendingDownloadOutput(sessionID int, remoteFilename string) string {
	key := downloadOutputKey{
		sessionID: sessionID,
		filename:  strings.ReplaceAll(remoteFilename, `\`, "/"),
	}

	pendingDownloadOutputs.Lock()
	defer pendingDownloadOutputs.Unlock()

	paths := pendingDownloadOutputs.paths[key]
	if len(paths) == 0 {
		return ""
	}

	outputPath := paths[0]
	if len(paths) == 1 {
		delete(pendingDownloadOutputs.paths, key)
	} else {
		pendingDownloadOutputs.paths[key] = paths[1:]
	}
	return outputPath
}

func SaveDownloadedFile(remoteFilename string, outputPath string, fileb64 string) {
	if outputPath == "" {
		DownloadFile(remoteFilename, fileb64)
		return
	}
	DownloadFileAs(outputPath, fileb64)
}
