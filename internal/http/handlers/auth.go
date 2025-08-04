package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/Dragodui/diploma-server/internal/models"
	"github.com/Dragodui/diploma-server/internal/services"
	"github.com/Dragodui/diploma-server/internal/utils"
	"github.com/go-chi/chi/v5"
	"github.com/markbates/goth/gothic"
)

type AuthHandler struct {
	svc *services.AuthService
}

func NewAuthHandler(svc *services.AuthService) *AuthHandler {
	return &AuthHandler{svc: svc}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var input models.RegisterInput

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.JSONError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := utils.Validate.Struct(input); err != nil {
		utils.JSONValidationErrors(w, err)
		return
	}
	err := h.svc.Register(input.Email, input.Password, input.Name)
	if err != nil {
		utils.JSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.svc.SendVerificationEmail(input.Email); err != nil {
		utils.JSONError(w, "Failed to send verification email", http.StatusInternalServerError)
		return
	}

	utils.JSON(w, http.StatusCreated, map[string]string{
		"message": "Registered successfully. Please check your email to verify your account.",
	})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var input models.LoginInput

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		utils.JSONError(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := utils.Validate.Struct(input); err != nil {
		utils.JSONValidationErrors(w, err)
		return
	}

	token, err := h.svc.Login(input.Email, input.Password)
	if err != nil {
		utils.JSONError(w, "Invalid credentials", http.StatusUnauthorized)
		return
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

func (h *AuthHandler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	token := chi.URLParam(r, "token")
	err := h.svc.VerifyEmail(token)
	if err != nil {
		utils.JSONError(w, "Incorrect or expired token", http.StatusBadRequest)
		return
	}
	utils.JSON(w, http.StatusOK, map[string]string{"message": "Email verified"})
}

func (h *AuthHandler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	h.svc.SendResetPassword(email)
	utils.JSON(w, http.StatusOK, map[string]string{"message": "Reset link was sended to your email"})
}

func (h *AuthHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	// ?token=...&password=...
	token := r.FormValue("token")
	pass := r.FormValue("password")
	if err := h.svc.ResetPassword(token, pass); err != nil {
		utils.JSONError(w, "Incorrect or expired token", http.StatusBadRequest)
		return
	}
	utils.JSON(w, http.StatusOK, map[string]string{"message": "Password changed successfully"})
}
