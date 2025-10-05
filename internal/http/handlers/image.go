package handlers

import (
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/Dragodui/diploma-server/internal/utils"
	"github.com/google/uuid"
)

type ImageHandler struct {
}

func NewImageHandler() *ImageHandler {
	return &ImageHandler{}
}

func (h *ImageHandler) UploadImage(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(10 << 20) // 10mb file limit

	if err != nil {
		utils.JSONError(w, "Error uploading image: "+err.Error(), http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("image")
	if err != nil {
		utils.JSONError(w, "File required: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	uploadDir := "./uploads"
	os.MkdirAll(uploadDir, os.ModePerm)

	ext := filepath.Ext(header.Filename)
	newName := uuid.New().String() + ext

	filePath := uploadDir + newName
	out, err := os.Create(filePath)
	if err != nil {
		utils.JSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer out.Close()
	io.Copy(out, file)

	publicPath := "/uploads/" + newName
	utils.JSON(w, http.StatusCreated, map[string]interface{}{
		"status":  true,
		"message": "File uploaded successfully",
		"url":     publicPath,
	})
}
