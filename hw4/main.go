package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"path"
	"time"
)

type Employee struct {
	Name   string  `json:"name" xml:"name"`
	Age    int     `json:"age" xml:"age"`
	Salary float32 `json:"salary" xml:"salary"`
}

type Handler struct {
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		name := r.FormValue("name")
		fmt.Fprintf(w, "Parsed query-param with key \"name\": %s", name)
	case http.MethodPost:
		var employee Employee
		contentType := r.Header.Get("Content-Type")
		switch contentType {
		case "application/json":
			err := json.NewDecoder(r.Body).Decode(&employee)
			if err != nil {
				http.Error(w, "Unable to unmarshal JSON", http.StatusBadRequest)
				return
			}
		case "application/xml":
			err := xml.NewDecoder(r.Body).Decode(&employee)
			if err != nil {
				http.Error(w, "Unable to unmarshal XML", http.StatusBadRequest)
				return
			}
		default:
			http.Error(w, "Unknown content type", http.StatusBadRequest)
			return
		}

		fmt.Fprintf(w, "Got a new employee!\nName: %s\nAge: %d y.o.\nSalary %0.2f\n",
			employee.Name,
			employee.Age,
			employee.Salary,
		)
	}
}

type UploadHandler struct {
	HostAddr  string
	UploadDir string
}

func (h *UploadHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Unable to read file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		http.Error(w, "Unable to read file", http.StatusBadRequest)
		return
	}

	filePath := h.UploadDir + "/" + header.Filename
	err = ioutil.WriteFile(filePath, data, 0777)
	if err != nil {
		log.Println(err)
		http.Error(w, "Unable to save file", http.StatusInternalServerError)
		return
	}
	fileLink := h.HostAddr + "/" + header.Filename

	fmt.Fprintln(w, fileLink)
}

func main() {
	handler := &Handler{}
	uploadHandler := &UploadHandler{
		UploadDir: "upload",
	}

	http.Handle("/", handler)
	http.Handle("/upload", uploadHandler)
	http.HandleFunc("/files", func(w http.ResponseWriter, r *http.Request) {
		extension := r.FormValue("extention")
		files, err := ioutil.ReadDir("upload")
		if err != nil {
			http.Error(w, "Unable to read files", http.StatusBadRequest)
			return
		}

		for _, file := range files {
			fileName := file.Name()
			fileExtension := path.Ext(fileName)
			if extension != "" {
				if fileExtension == extension {
					fmt.Fprintln(w, fileName, extension, file.Size())
				}
			} else {
				fmt.Fprintln(w, fileName, extension, file.Size())
			}
		}
	})

	srv := &http.Server{
		Addr:         ":80",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	dirToServe := http.Dir(uploadHandler.UploadDir)
	fs := &http.Server{
		Addr:         ":8080",
		Handler:      http.FileServer(dirToServe),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	go fs.ListenAndServe()
	srv.ListenAndServe()

}
