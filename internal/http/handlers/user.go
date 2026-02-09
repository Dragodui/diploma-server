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

// GetMe godoc
// @Summary      Get current user
// @Description  Get details of the currently logged in user
// @Tags         user
// @Produce      json
// @Security     BearerAuth
// @Success      202  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /user [post]
func (h *UserHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r)
	if userID == 0 {
		utils.JSONError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	user, err := h.svc.GetUserByID(r.Context(), userID)
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

// Update godoc
// @Summary      Update user
// @Description  Update user name or avatar
// @Tags         user
// @Accept       multipart/form-data
// @Produce      json
// @Security     BearerAuth
// @Param        name formData string false "User Name"
// @Param        avatar formData file false "User Avatar"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /user [patch]
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
	avatarURL := r.FormValue("avatar")

	file, fileHeader, err := r.FormFile("avatar_file")
	hasAvatarFile := err == nil

	if name == "" && !hasAvatarFile && avatarURL == "" {
		utils.JSONError(w, "No fields to update", http.StatusBadRequest)
		return
	}

	if name != "" {
		if err := h.svc.UpdateUser(r.Context(), userID, name); err != nil {
			utils.JSONError(w, "Failed to update name: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	if avatarURL != "" {
		if err := h.svc.UpdateUserAvatar(r.Context(), userID, avatarURL); err != nil {
			utils.JSONError(w, "Failed to update avatar: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	if hasAvatarFile {
		defer file.Close()

		imagePath, err := h.imageSvc.Upload(r.Context(), file, fileHeader)
		if err != nil {
			utils.JSONError(w, "Failed to upload avatar: "+err.Error(), http.StatusInternalServerError)
			return
		}

		if err := h.svc.UpdateUserAvatar(r.Context(), userID, imagePath); err != nil {
			utils.JSONError(w, "Failed to update avatar: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	utils.JSON(w, http.StatusOK, map[string]interface{}{
		"status":  true,
		"message": "User updated successfully",
	})
}
