package handlers

import (
	"net/http"

	"github.com/Dragodui/diploma-server/internal/http/middleware"
	"github.com/Dragodui/diploma-server/internal/services"
	"github.com/Dragodui/diploma-server/internal/utils"
)

type UserHandler struct {
	svc      services.IUserService
	imageSvc services.IImageService
}

func NewUserHandler(svc services.IUserService, imgSvc services.IImageService) *UserHandler {
	return &UserHandler{svc: svc, imageSvc: imgSvc}
}

func (h *UserHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	if userID == 0 {
		utils.JSONError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	user, err := h.svc.GetUserByID(userID)
	if err != nil {
		utils.JSONError(w, err.Error(), http.StatusInternalServerError)
	}

	if user == nil {
		utils.JSONError(w, "User not found", http.StatusNotFound)
		return
	}

	utils.JSON(w, http.StatusAccepted, map[string]interface{}{
		"status": true,
		"user":   user,
	})
}

func (h *UserHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	if userID == 0 {
		utils.JSONError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if err := r.ParseMultipartForm(10 << 20); err != nil { // до 10 МБ
		utils.JSONError(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	name := r.FormValue("name")

	file, fileHeader, err := r.FormFile("avatar")
	hasAvatar := err == nil

	if name == "" && !hasAvatar {
		utils.JSONError(w, "No fields to update", http.StatusBadRequest)
		return
	}

	if name != "" {
		if err := h.svc.UpdateUser(userID, name); err != nil {
			utils.JSONError(w, "Failed to update name: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	if hasAvatar {
		defer file.Close()

		imagePath, err := h.imageSvc.Upload(file, fileHeader)
		if err != nil {
			utils.JSONError(w, "Failed to upload avatar: "+err.Error(), http.StatusInternalServerError)
			return
		}

		if err := h.svc.UpdateUserAvatar(userID, imagePath); err != nil {
			utils.JSONError(w, "Failed to update avatar: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"status":  true,
		"message": "User updated successfully",
	})
}
