package utils

// import (
// 	"fmt"
// 	"io/ioutil"
// 	"log"
// 	"net/http"
// )

// type UploadHandler struct {
// 	HostAddr  string
// 	UploadDir string
// }

// func (h *UploadHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
// 	file, header, err := r.FormFile("file")
// 	if err != nil {
// 		http.Error(w, "Unable to read file", http.StatusBadRequest)
// 		return
// 	}
// 	defer file.Close()
// 	data, err := ioutil.ReadAll(file)
// 	if err != nil {
// 		http.Error(w, "Unable to read file", http.StatusBadRequest)
// 		return
// 	}
// 	filePath := h.UploadDir + "/" + header.Filename
// 	err = ioutil.WriteFile(filePath, data, 0777)
// 	if err != nil {
// 		log.Println(err)
// 		http.Error(w, "Unable to save file", http.StatusInternalServerError)
// 		return
// 	}
// 	fileLink := h.HostAddr + "/" + header.Filename
// 	fmt.Fprintln(w, fileLink)
// }
