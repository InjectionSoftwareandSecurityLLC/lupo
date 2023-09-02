package core

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
)

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

	WarningColorBold.Println("\nDownloading file: " + filename + "...")

	fileb64_url_decoded, err := url.QueryUnescape(fileb64)
	if err != nil {
		fmt.Println("Error decoding:", err)
		return
	}

	file, err := base64.StdEncoding.DecodeString(fileb64_url_decoded)
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
