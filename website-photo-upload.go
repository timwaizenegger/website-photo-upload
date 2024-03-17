package main

import (
	"crypto/sha256"
	"fmt"
	"gopkg.in/gographics/imagick.v2/imagick"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"path/filepath"
)

// Notes:
// export CGO_CFLAGS_ALLOW='-Xpreprocessor'
// might need export PKG_CONFIG_PATH="/opt/homebrew/opt/imagemagick@6/lib/pkgconfig"

var (
	mw *imagick.MagickWand
)

const (
	bindHost    = "0.0.0.0:3333"
	imagePath   = "./images/"
	thumbPath   = "./images/thumbs/"
	thumbSuffix = ".thumb.jpg"
)

func putUpload(w http.ResponseWriter, r *http.Request) {
	log.Printf("upload %q %q %q", r, r.URL, r.Method)
	err := r.ParseMultipartForm(1000 * 1024 * 1024)
	if err != nil {
		log.Printf("error parsing multipart form: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get a reference to the file
	_, _, err = r.FormFile("imageInputName")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	// TODO: return error if any happened
	for k, v := range r.MultipartForm.File {
		log.Printf("got a file %s", k)
		for _, header := range v {
			log.Printf("got a file header %v", header.Filename)
			err = saveUploadedFile(header)
			if err != nil {
				log.Printf("error saving file header: %v", err)
			}
		}
	}

	//io.WriteString(w, "upload processed successfully\n")
	//http.Redirect(w, r, "./upload", http.StatusFound)
	http.ServeFile(w, r, "html/reloader.html")
	// serveMainPage(w, r)
}

// saveUploadedFile ...
// TODO: save as sha sum key
func saveUploadedFile(h *multipart.FileHeader) error {
	log.Printf("writing file to disk %q size %dKB", h.Filename, h.Size/1024)
	f, err := h.Open()
	if err != nil {
		log.Printf("error opening file %q: %v", h.Filename, err)
		return err
	}
	defer f.Close()
	buffer := make([]byte, h.Size)
	f.Read(buffer)
	sum := sha256.Sum256(buffer)
	ext := filepath.Ext(h.Filename)
	if ext == "" {
		ext = ".jpg"
	}
	fileName := fmt.Sprintf("%x%s", sum[:], ext)
	targetPath := filepath.Join(imagePath, fileName)
	out, err := os.Create(targetPath)
	if err != nil {
		log.Printf("error creating local file %q: %v", targetPath, err)
		return err
	}
	_, err = out.Write(buffer)
	if err != nil {
		log.Printf("error writing to local file %q: %v", targetPath, err)
		return err
	}
	err = makeThumbnail(targetPath)
	if err != nil {
		os.Remove(targetPath)
	}
	return nil
}

func makeThumbnail(imagePath string) error {
	imageDir := path.Dir(imagePath)
	imageName := path.Base(imagePath)
	outDir := path.Join(imageDir, "thumbs")
	outName := fmt.Sprintf("%s%s", imageName, thumbSuffix)
	outPath := path.Join(outDir, outName)

	log.Printf("making a thumbnail for %q. Will pace the file at %q", imagePath, outPath)
	ret, err := imagick.ConvertImageCommand([]string{
		"convert", imagePath, "-auto-orient", "-thumbnail", "256x256", outPath,
	})
	if err != nil {
		log.Printf("ERROR making a thumbnail for %q, %s", imagePath, err)
		return err
	}
	log.Printf("thumbnail result: %v", ret)
	return nil
}

func serveMainPage(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "html/upload.html")
}

func main() {
	log.Printf("starting 'website-photo-upload' ... ")
	imagick.Initialize()
	defer imagick.Terminate()

	mw = imagick.NewMagickWand()

	mux := http.NewServeMux()
	//mux.Handle("/html/", http.StripPrefix("/html/", http.FileServer(http.Dir("./html"))))
	mux.Handle("/images/", http.StripPrefix("/images/", http.FileServer(http.Dir("./images/"))))
	mux.HandleFunc("/api/thumbs", jsonDirList)
	//mux.HandleFunc("/start", serveMainPage)
	mux.HandleFunc("GET /{$}", serveMainPage)
	mux.HandleFunc("GET /upload", serveMainPage)
	mux.HandleFunc("POST /upload", putUpload)

	log.Printf("starting server on %s", bindHost)
	//err := http.ListenAndServe("127.0.0.1:3333", mux)
	err := http.ListenAndServe(bindHost, mux)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("finished server on %s", bindHost)
}
