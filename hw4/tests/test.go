package test

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type Handler struct{}

func HealthCheckHandler(w http.ResponseWriter, r *http.Request) {
	// A very simple health check.
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")

	// In the future we could report back on the status of our DB, or our cache
	// (e.g. Redis) by performing a simple PING, and include them in the response.
	io.WriteString(w, `{"alive": true}`)
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

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

func TestGetHandler(t *testing.T) {
	// Создаем запрос с указанием нашего хендлера. Так как мы тестируем GET-эндпоинт
	// то нам не нужно передавать тело, поэтому третьим аргументом передаем nil
	req, err := http.NewRequest("GET", "/?name=John", nil)
	if err != nil {
		t.Fatal(err)
	}
	// Мы создаем ResponseRecorder(реализует интерфейс http.ResponseWriter)
	// и используем его для получения ответа
	rr := httptest.NewRecorder()
	handler := &Handler{}
	// handler := http.HandlerFunc(HealthCheckHandler)
	// Наш хендлер соответствует интерфейсу http.Handler, а значит
	// мы можем использовать ServeHTTP и напрямую указать
	// Request и ResponseRecorder

	handler.ServeHTTP(rr, req)
	// Проверяем статус-код ответа
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
	// Проверяем тело ответа
	expected := `Parsed query-param with key "name": John`
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected)
	}
}

func TestUploadHandler(t *testing.T) {
	// открываем файл, который хотим отправить
	file, _ := os.Open("testfile")
	defer file.Close()
	// действия, необходимые для того, чтобы засунуть файл в запрос
	// в качестве мультипарт-формы
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", filepath.Base(file.Name()))
	io.Copy(part, file)
	writer.Close()
	// опять создаем запрос, теперь уже на /upload эндпоинт
	req, _ := http.NewRequest(http.MethodPost, "/upload", body)
	req.Header.Add("Content-Type", writer.FormDataContentType())
	// создаем ResponseRecorder
	rr := httptest.NewRecorder()
	// создаем заглушку файлового сервера. Для прохождения тестов
	// нам достаточно чтобы он возвращал 200 статус
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "ok!")
	}))
	defer ts.Close()
	uploadHandler := &UploadHandler{
		UploadDir: "upload",
		// таким образом мы подменим адрес файлового сервера
		// и вместо реального, хэндлер будет стучаться на заглушку
		// которая всегда будет возвращать 200 статус, что нам и нужна
		HostAddr: ts.URL,
	}
	// опять же, вызываем ServeHTTP у тестируемого обработчика
	uploadHandler.ServeHTTP(rr, req)
	// Проверяем статус-код ответа
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
	expected := `testfile`
	if !strings.Contains(rr.Body.String(), expected) {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected)
	}
}
