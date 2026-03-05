package handlers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/Dragodui/diploma-server/internal/services"
	"github.com/Dragodui/diploma-server/internal/utils"
	"github.com/google/uuid"
)

const maxOCRFileSize = 10 << 20 // 10 MB

var allowedOCRTypes = map[string]string{
	"application/pdf": ".pdf",
	"image/jpeg":      ".jpg",
	"image/png":       ".png",
	"image/gif":       ".gif",
	"image/webp":      ".webp",
	"image/bmp":       ".bmp",
}

type OCRHandler struct {
	svc services.IOCRService
}

func NewOCRHandler(svc services.IOCRService) *OCRHandler {
	return &OCRHandler{svc: svc}
}

// Process godoc
// @Summary      Process receipt file (image or PDF) with OCR
// @Description  Upload a file directly for OCR processing. Supports images and PDFs.
// @Tags         ocr
// @Accept       multipart/form-data
// @Produce      json
// @Security     BearerAuth
// @Param        file formData file true "Receipt file (image or PDF)"
// @Param        language formData string false "Language of the receipt text"
// @Success      200  {object}  models.OCRResult
// @Failure      400  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /ocr/process [post]
func (h *OCRHandler) Process(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(maxOCRFileSize); err != nil {
		utils.JSONError(w, "File too large or invalid form data", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		utils.JSONError(w, "Missing file field", http.StatusBadRequest)
		return
	}
	defer file.Close()

	if header.Size > maxOCRFileSize {
		utils.JSONError(w, fmt.Sprintf("File size exceeds maximum of %d bytes", maxOCRFileSize), http.StatusBadRequest)
		return
	}

	// Detect content type from file content
	buf := make([]byte, 512)
	n, err := file.Read(buf)
	if err != nil {
		utils.JSONError(w, "Failed to read file", http.StatusBadRequest)
		return
	}
	contentType := http.DetectContentType(buf[:n])

	// Reset file reader
	if _, err := file.Seek(0, 0); err != nil {
		utils.JSONError(w, "Failed to process file", http.StatusInternalServerError)
		return
	}

	// Check if PDF by extension (DetectContentType may not detect PDF reliably)
	ext := strings.ToLower(filepath.Ext(header.Filename))
	if ext == ".pdf" {
		contentType = "application/pdf"
	}

	// Validate content type
	fileExt, ok := allowedOCRTypes[contentType]
	if !ok {
		utils.JSONError(w, fmt.Sprintf("Unsupported file type: %s. Allowed: PDF, JPEG, PNG, GIF, WebP, BMP", contentType), http.StatusBadRequest)
		return
	}

	// Save to temp file
	tempPath := filepath.Join(os.TempDir(), fmt.Sprintf("ocr_%s%s", uuid.New().String(), fileExt))
	outFile, err := os.Create(tempPath)
	if err != nil {
		utils.JSONError(w, "Failed to process file", http.StatusInternalServerError)
		return
	}

	if _, err := io.Copy(outFile, file); err != nil {
		outFile.Close()
		os.Remove(tempPath)
		utils.JSONError(w, "Failed to save file", http.StatusInternalServerError)
		return
	}
	outFile.Close()
	defer os.Remove(tempPath)

	language := r.FormValue("language")

	result, err := h.svc.ProcessFile(r.Context(), tempPath, language)
	if err != nil {
		utils.SafeError(w, err, "OCR processing failed", http.StatusInternalServerError)
		return
	}

	utils.JSON(w, http.StatusOK, result)
}
