package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

func jsonDirList(w http.ResponseWriter, r *http.Request) {
	dir := imagePath
	fileInfos, err := listFiles(dir)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonBytes, err := json.Marshal(fileInfos)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonBytes)
}

type fileInfo struct {
	Name      string `json:"name"`
	ThumbName string `json:"thumb_name"`
	ThumbPath string `json:"thumb_path"`
	ImgPath   string `json:"img_path"`
}

func listFiles(dir string) ([]fileInfo, error) {
	var files []fileInfo
	dirEntries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	for _, de := range dirEntries {
		if !de.IsDir() {
			files = append(files, fileInfo{
				Name:      de.Name(),
				ThumbName: fmt.Sprintf("%s%s", de.Name(), thumbSuffix),
				ThumbPath: fmt.Sprintf("%s%s%s", thumbPath, de.Name(), thumbSuffix),
				ImgPath:   fmt.Sprintf("%s%s", imagePath, de.Name()),
			})
		}
	}
	return files, nil
}
