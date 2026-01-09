package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/Dragodui/diploma-server/internal/models"
	"github.com/Dragodui/diploma-server/internal/services"
	"github.com/Dragodui/diploma-server/internal/utils"
	"github.com/go-playground/validator/v10"
)

type OCRHandler struct {
	svc      services.IOCRService
	validate *validator.Validate
}

func NewOCRHandler(svc services.IOCRService) *OCRHandler {
	return &OCRHandler{
		svc:      svc,
		validate: validator.New(),
	}
}

// ProcessImage godoc
// @Summary      Process receipt image with OCR
// @Description  Extract text and structured data from receipt/invoice image
// @Tags         ocr
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body models.OCRRequest true "Image URL"
// @Success      200  {object}  models.OCRResult
// @Failure      400  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /ocr/process [post]
func (h *OCRHandler) ProcessImage(w http.ResponseWriter, r *http.Request) {
	var req models.OCRRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.JSONError(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		utils.JSONError(w, "Validation error: "+err.Error(), http.StatusBadRequest)
		return
	}

	result, err := h.svc.ProcessImage(req.ImageURL)
	if err != nil {
		utils.JSONError(w, "OCR processing failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	utils.JSON(w, http.StatusOK, result)
}
