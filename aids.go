package main

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/jboelter/notificator"
)

var StdChars = []byte("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789")
var notify *notificator.Notificator

func main() {
	notify = notificator.New(notificator.Options{AppName: "Aids", DefaultIcon: ""})
	//$TMPDIR is empty, at least on my machine...
	//tmpdir := os.TempDir()
	file := "/tmp/"
	file += newLenChars(10, StdChars)
	file += ".png"
	if runtime.GOOS == "darwin" {
		exec.Command("screencapture", "-i", file).Run()
	} else {
		exec.Command("scrot", file, "-s").Run()
	}
	notify.Push("aids", "Uploading...")
	url, err := Upload(file)
	if err != nil {
		notify.Push("aids", err.Error())
		return
	}
	notify.Push("aids", "Done!")
	if runtime.GOOS == "darwin" {
		exec.Command("open", url).Run()
	} else {
		exec.Command("xdg-open", url).Run()
	}

}
func Upload(filename string) (string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", filepath.Base(filename))
	if err != nil {
		return "", err
	}
	_, err = io.Copy(part, file)
	err = writer.Close()
	if err != nil {
		return "", err
	}
	r, err := http.NewRequest("POST", "http://i.aidskrebs.net/upload", body)
	if err != nil {
		return "", err
	}
	r.Header.Set("Content-Type", writer.FormDataContentType())
	return sendFile(r)
}

func sendFile(r *http.Request) (string, error) {
	client := &http.Client{}
	resp, err := client.Do(r)
	if err != nil {
		return "", nil
	}
	body := &bytes.Buffer{}
	_, err = body.ReadFrom(resp.Body)
	if err != nil {
		return "", nil
	}
	resp.Body.Close()
	fmt.Println(resp.Request.URL.String())
	return resp.Request.URL.String(), nil
}

func newLenChars(length int, chars []byte) string {
	b := make([]byte, length)
	r := make([]byte, length+(length/4)) // storage for random bytes.
	clen := byte(len(chars))
	maxrb := byte(256 - (256 % len(chars)))
	i := 0
	for {
		if _, err := io.ReadFull(rand.Reader, r); err != nil {
			panic("error reading from random source: " + err.Error())
		}
		for _, c := range r {
			if c >= maxrb {
				continue
			}
			b[i] = chars[c%clen]
			i++
			if i == length {
				return string(b)
			}
		}
	}
	panic("unreachable")
}
