package handlers

import (
	"net/http"

	"github.com/Dragodui/diploma-server/internal/services"
	"github.com/Dragodui/diploma-server/internal/utils"
)

type ImageHandler struct {
	svc services.IImageService
}

func NewImageHandler(svc services.IImageService) *ImageHandler {
	return &ImageHandler{svc: svc}
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

	publicPath, err := h.svc.Upload(file, header)

	if err != nil {
		utils.JSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	utils.JSON(w, http.StatusCreated, map[string]interface{}{
		"status":  true,
		"message": "File uploaded successfully",
		"url":     publicPath,
	})
}
