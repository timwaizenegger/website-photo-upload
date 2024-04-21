package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sort"
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
	TimeStamp string `json:"time_stamp"`
	GroupName string `json:"group_name"`
}

type ByNameAplhabetically []fileInfo

func (a ByNameAplhabetically) Len() int {
	return len(a)
}

func (a ByNameAplhabetically) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a ByNameAplhabetically) Less(i, j int) bool {
	return a[i].Name > a[j].Name
}

func listFiles(dir string) ([]fileInfo, error) {
	var files []fileInfo
	dirEntries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	for _, de := range dirEntries {
		if !de.IsDir() {
			t := filenameToTime(de.Name())
			files = append(files, fileInfo{
				Name:      de.Name(),
				ThumbName: fmt.Sprintf("%s%s", de.Name(), thumbSuffix),
				ThumbPath: fmt.Sprintf("%s%s%s", thumbPath, de.Name(), thumbSuffix),
				ImgPath:   fmt.Sprintf("%s%s", imagePath, de.Name()),
				TimeStamp: fmt.Sprintf("%s", t),
				GroupName: groupNameForDate(t),
			})
		}
	}
	sort.Sort(ByNameAplhabetically(files))
	return files, nil
}
