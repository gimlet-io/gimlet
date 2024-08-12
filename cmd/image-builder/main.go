package main

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
)

func main() {
	fmt.Println("App init..")

	err := godotenv.Load(".env")
	if err != nil {
		log.Warnf("could not load .env file, relying on env vars")
	}

	r := chi.NewRouter()

	// Basic CORS
	// for more ideas, see: https://developer.github.com/v3/#cross-origin-resource-sharing
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"}, // Use this to allow specific origin hosts
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	})
	r.Post("/build-image", uploadFile)

	err = http.ListenAndServe(":9000", r)
	log.Error(err)
}

func uploadFile(w http.ResponseWriter, r *http.Request) {
	fmt.Println("File Upload Endpoint Hit")

	image := r.FormValue("image")
	tag := r.FormValue("tag")
	app := r.FormValue("app")
	configJson := r.FormValue("configJson")
	// cacheImage := r.FormValue("cacheImage")
	// previousImage := r.FormValue("previousImage")
	// Parse our multipart form, 1000 << 20 specifies a maximum
	// upload of 1000 MB files.
	r.ParseMultipartForm(1000 << 20)
	// FormFile returns the first file for the given key `myFile`
	// it also returns the FileHeader so we can get the Filename,
	// the Header and the size of the file
	file, handler, err := r.FormFile("data")
	if err != nil {
		fmt.Println("Error Retrieving the File")
		fmt.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer file.Close()
	fmt.Printf("Uploaded File: %+v\n", handler.Filename)
	fmt.Printf("File Size: %+v\n", handler.Size)
	fmt.Printf("MIME Header: %+v\n", handler.Header)

	if configJson != "" {
		// write docker config
		err := os.MkdirAll("/home/cnb/.docker", 0755)
		if err != nil {
			fmt.Println(err)
			w.Write([]byte("IMAGE BUILD ERROR"))
			return
		}

		err = os.WriteFile("/home/cnb/.docker/config.json", []byte(configJson), 0755)
		if err != nil {
			fmt.Println(err)
			w.Write([]byte("IMAGE BUILD ERROR"))
			return
		}
	}

	// Create a temporary file within our temp-images directory that follows
	// a particular naming pattern
	tempFile, err := ioutil.TempFile("/tmp", "source-*.tar.gz")
	if err != nil {
		fmt.Println(err)
		w.Write([]byte("IMAGE BUILD ERROR"))
		return
	}
	defer tempFile.Close()
	fmt.Println(tempFile.Name())

	// read all of the contents of our uploaded file into a
	// byte array
	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		fmt.Println(err)
		w.Write([]byte("IMAGE BUILD ERROR"))
		return
	}
	// write this byte array to our temporary file
	tempFile.Write(fileBytes)

	reader, err := os.Open(tempFile.Name())
	if err != nil {
		fmt.Println(err)
		w.Write([]byte("IMAGE BUILD ERROR"))
		return
	}

	sourcePath := "/home/cnb/" + app
	defer os.RemoveAll(sourcePath)
	err = Untar(sourcePath, reader)
	if err != nil {
		fmt.Println(err)
		w.Write([]byte("IMAGE BUILD ERROR"))
		return
	}

	w.WriteHeader(http.StatusOK)

	// shell out to buildpacks
	command := "/cnb/lifecycle/creator"
	sourcePathArg := "-app=" + sourcePath
	logLevelArg := "-log-level=debug"
	// cacheImage = "-cache-image=" + cacheImage
	// previousImage = "-previous-image=" + previousImage
	// cmd := exec.Command(command, sourcePath, cacheImage, previousImage, image)
	cmd := exec.Command(command, sourcePathArg, logLevelArg, image+":"+tag)
	pipeReader, pipeWriter := io.Pipe()
	cmd.Stdout = pipeWriter
	cmd.Stderr = pipeWriter
	go writeCmdOutput(w, pipeReader)
	err = cmd.Run()
	if err != nil {
		fmt.Println(err)
		w.Write([]byte("IMAGE BUILD ERROR"))
		return
	}
	pipeWriter.Close()

	if err != nil {
		fmt.Println(err)
		w.Write([]byte("IMAGE BUILD ERROR"))
		return
	}

	w.Write([]byte("IMAGE BUILT"))
}

func writeCmdOutput(res http.ResponseWriter, pipeReader *io.PipeReader) {
	buffer := make([]byte, 1024)
	for {
		n, err := pipeReader.Read(buffer)
		if err != nil {
			pipeReader.Close()
			break
		}

		data := buffer[0:n]
		fmt.Print(string(data))
		res.Write(data)
		if f, ok := res.(http.Flusher); ok {
			f.Flush()
		}
		//reset buffer
		for i := 0; i < n; i++ {
			buffer[i] = 0
		}
	}
}

// Untar takes a destination path and a reader; a tar reader loops over the tarfile
// creating the file structure at 'dst' along the way, and writing any files
func Untar(dst string, r io.Reader) error {

	gzr, err := gzip.NewReader(r)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()

		switch {

		// if no more files are found return
		case err == io.EOF:
			return nil

		// return any other error
		case err != nil:
			return err

		// if the header is nil, just skip it (not sure how this happens)
		case header == nil:
			continue
		}

		// the target location where the dir/file should be created
		target := filepath.Join(dst, header.Name)

		// the following switch could also be done using fi.Mode(), not sure if there
		// a benefit of using one vs. the other.
		// fi := header.FileInfo()

		// check the file type
		switch header.Typeflag {

		// if its a dir and it doesn't exist create it
		case tar.TypeDir:
			if _, err := os.Stat(target); err != nil {
				if err := os.MkdirAll(target, 0755); err != nil {
					return err
				}
			}

		// if it's a file create it
		case tar.TypeReg:
			f, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return err
			}

			// copy over contents
			if _, err := io.Copy(f, tr); err != nil {
				return err
			}

			// manually close here after each file operation; defering would cause each file close
			// to wait until all operations have completed.
			f.Close()
		}
	}
}
