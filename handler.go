package main

import (
	"net/http"
	"os"
	"strings"
	"fmt"
	"archive/tar"
	"compress/gzip"
	"io"
	"path/filepath"
	"log"
)

func updateHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			w.WriteHeader(404)
			return
		}

		token, hasToken := r.Header["Token"]
		host, hasHost := r.Header["Target-Host"]

		if !hasToken || !hasHost {
			w.WriteHeader(400)
			return
		}

		log.Print("Incoming deployment for: " + host[0])

		directory := getDirectory(host[0]);
		file, err := os.ReadFile(directory + "/token")
		if err != nil {
			log.Print("Site doesn't exist")
			w.WriteHeader(404)
			return
		}

		fileToken := strings.TrimSpace(string(file))
		if token[0] != fileToken {
			log.Print("Failed token check")
			w.WriteHeader(401)
			return
		}

		r.ParseMultipartForm(32 << 28)

		archive, _, err := r.FormFile("file")
		if err != nil {
			w.WriteHeader(400)
			w.Write([]byte(fmt.Sprintf("Error: %s", err)))
			return
		}
		defer archive.Close()

		gunzipped, err := gzip.NewReader(archive)
		if err != nil {
			w.WriteHeader(400)
			w.Write([]byte(fmt.Sprintf("Error: %s", err)))
			return
		}

		os.MkdirAll(directory + "/html.new", 0755)
		tarball := tar.NewReader(gunzipped)
		if err := unpack(tarball, directory + "/html.new"); err != nil {
			w.WriteHeader(400)
			w.Write([]byte(fmt.Sprintf("Error: %s", err)))
			return
		}

		log.Print("Unpack success! Moving things in place")
		if err := os.RemoveAll(directory + "/html"); err != nil {
			w.WriteHeader(500)
			w.Write([]byte(fmt.Sprintf("Error: %s", err)))
			return
		}
		if err := os.Rename(directory + "/html.new", directory + "/html"); err != nil {
			w.WriteHeader(500)
			w.Write([]byte(fmt.Sprintf("Error: %s", err)))
			return
		}
		
		log.Print("Deployment complete")
		w.WriteHeader(200)
		w.Write([]byte("Okay!"))
	})
}

func unpack(tarball *tar.Reader, targetDir string) error {
	log.Print("Starting unpack to: " + targetDir)

	for {
        header, err := tarball.Next()
        if err == io.EOF {
            break
        } else if err != nil {
            return err
        }
 
        path := filepath.Join(targetDir, header.Name)
		path = strings.ReplaceAll(path, "..", "")

        info := header.FileInfo()
        if info.IsDir() {
			log.Print("Creating directory: " + path)
            if err = os.MkdirAll(path, 0755); err != nil {
                return err
            }
            continue
        }
 
		log.Print("Creating file: " + path)
        file, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
        if err != nil {
            return err
        }
        defer file.Close()
        _, err = io.Copy(file, tarball)
        if err != nil {
            return err
        }
    }
	
	return nil
}
