package files

import (
	"chat-server/storage"
	"fmt"
	"net/http"
)

type FileHandler struct {
	storage storage.FileStorage
}

func NewFileHandler(storage storage.FileStorage) *FileHandler {
	return &FileHandler{
		storage: storage,
	}
}

func (h *FileHandler) Upload(w http.ResponseWriter, r *http.Request) {

	//Get Uploaded file
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "file is required", http.StatusBadRequest)
		return

	}
	defer file.Close()

	fmt.Println("Received file:", header.Filename)
	fmt.Println("Size:", header.Size)
	fmt.Println("Content-Type:", header.Header.Get("Content-Type"))

	// Temporary object name
	objectName := header.Filename

	err = h.storage.Upload(
		r.Context(),
		objectName,
		file,
		header.Size,
		header.Header.Get("Content-Type"),
	)

	if err != nil {
		http.Error(w, "failed to upload file", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	fmt.Fprintf(w, "file uploaded successfully: %s", objectName)

}
