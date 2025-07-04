package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/Dragodui/diploma-server/internal/models"
	"github.com/Dragodui/diploma-server/internal/services"
	"github.com/Dragodui/diploma-server/internal/utils"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/markbates/goth/gothic"
)

// Create new validator
var validate = validator.New()

type AuthHandler struct {
	svc *services.AuthService
}

func NewAuthHandler(svc *services.AuthService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var input models.RegisterInput

	// Take arguments from response body
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.JSONError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate the fields
	if err := validate.Struct(input); err != nil {
		utils.JSONValidationErrors(w, err)
		return
	}

	if err := h.svc.Register(input.Email, input.Password, input.Name); err != nil {
		utils.JSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	utils.JSON(w, http.StatusCreated, map[string]string{"message": "Registered successfully"})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var input models.LoginInput

	// Take arguments from response body
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.JSONError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate the fields
	if err := validate.Struct(input); err != nil {
		utils.JSONValidationErrors(w, err)
		return
	}

	// Get token from Service
	token, err := h.svc.Login(input.Email, input.Password)
	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
	}

	// Response to client
	utils.JSON(w, http.StatusAccepted, map[string]string{
		"token": token,
	})
}

func (h *AuthHandler) SignInWithProvider(w http.ResponseWriter, r *http.Request) {
	provider := chi.URLParam(r, "provider")
	q := r.URL.Query()
	q.Add("provider", provider)

	r.URL.RawQuery = q.Encode()

	gothic.BeginAuthHandler(w, r)
}

func (h *AuthHandler) CallbackHandler(w http.ResponseWriter, r *http.Request) {
	provider := chi.URLParam(r, "provider")
	q := r.URL.Query()
	q.Add("provider", provider)

	r.URL.RawQuery = q.Encode()
	user, err := gothic.CompleteUserAuth(w, r)
	if err != nil {
		utils.JSONError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	redirectURL, err := h.svc.HandleCallback(user)

	if err != nil {
		utils.JSONError(w, err.Error(), http.StatusInternalServerError)
	}

	http.Redirect(w, r, redirectURL, http.StatusTemporaryRedirect)
}
