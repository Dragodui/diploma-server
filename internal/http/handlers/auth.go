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
	svc services.IAuthService
}

func NewAuthHandler(svc services.IAuthService) *AuthHandler {
	return &AuthHandler{svc}
}

// RegenerateVerify godoc
// @Summary      Regenerate verification email
// @Description  Resends the verification email to the user
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        email query string true "User Email"
// @Success      200  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /auth/verify/regenerate [get]
func (h *AuthHandler) RegenerateVerify(w http.ResponseWriter, r *http.Request) {
	// oldToken := chi.URLParam(r, "old_token")
	email := r.URL.Query().Get("email")
	// if oldToken != "" {
	// 	user, err := h.svc.GetUserByVerifyToken(oldToken)
	// 	if err != nil {
	// 		utils.JSONError(w, err.Error(), http.StatusInternalServerError)
	// 		return
	// 	}
	// 	email = user.Email

	// }
	if err := h.svc.SendVerificationEmail(email); err != nil {
		utils.JSONError(w, "Failed to send verification email", http.StatusInternalServerError)
		return
	}
}

// Register godoc
// @Summary      Register a new user
// @Description  Register a new user with email, password and name
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        input body models.RegisterInput true "Register Input"
// @Success      201  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /auth/register [post]
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

	utils.JSON(w, http.StatusCreated, map[string]interface{}{
		"status":  true,
		"message": "Registered successfully. Please check your email to verify your account.",
	})
}

// Login godoc
// @Summary      Login user
// @Description  Login with email and password
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        input body models.LoginInput true "Login Input"
// @Success      202  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Router       /auth/login [post]
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

	user, err := h.svc.GetUserByEmail(input.Email)
	if err != nil {
		utils.JSONError(w, err.Error(), http.StatusUnauthorized)
		return
	}

	if !user.EmailVerified {
		utils.JSONError(w, "Email is not verified", http.StatusUnauthorized)
		return
	}

	token, user, err := h.svc.Login(input.Email, input.Password)
	if err != nil {
		utils.JSONError(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	// Response to client
	utils.JSON(w, http.StatusAccepted, map[string]interface{}{"status": true,
		"token": token,
		"user":  user,
	})
}

// SignInWithProvider godoc
// @Summary      Sign in with provider
// @Description  Initiate OAuth2 login with a provider (google, etc.)
// @Tags         auth
// @Param        provider path string true "Provider"
// @Router       /auth/{provider} [get]
func (h *AuthHandler) SignInWithProvider(w http.ResponseWriter, r *http.Request) {
	provider := chi.URLParam(r, "provider")
	q := r.URL.Query()
	q.Add("provider", provider)

	r.URL.RawQuery = q.Encode()

	gothic.BeginAuthHandler(w, r)
}

// CallbackHandler godoc
// @Summary      OAuth2 Callback
// @Description  Handle OAuth2 callback
// @Tags         auth
// @Param        provider path string true "Provider"
// @Success      307
// @Failure      500  {object}  map[string]interface{}
// @Router       /auth/{provider}/callback [get]
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

// VerifyEmail godoc
// @Summary      Verify email
// @Description  Verify user email with token
// @Tags         auth
// @Param        token query string true "Verification Token"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Router       /auth/verify [get]
func (h *AuthHandler) VerifyEmail(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	err := h.svc.VerifyEmail(token)
	if err != nil {
		utils.JSONError(w, "Incorrect or expired token", http.StatusBadRequest)
		return
	}
	utils.JSON(w, http.StatusOK, map[string]interface{}{"status": true, "message": "Email verified"})
}

// ForgotPassword godoc
// @Summary      Forgot password
// @Description  Send reset password link to email
// @Tags         auth
// @Accept       x-www-form-urlencoded
// @Produce      json
// @Param        email formData string true "User Email"
// @Success      200  {object}  map[string]interface{}
// @Router       /auth/forgot [post]
func (h *AuthHandler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	h.svc.SendResetPassword(email)
	utils.JSON(w, http.StatusOK, map[string]interface{}{"status": true, "message": "Reset link was sended to your email"})
}

// ResetPassword godoc
// @Summary      Reset password
// @Description  Reset password with token
// @Tags         auth
// @Accept       x-www-form-urlencoded
// @Produce      json
// @Param        token formData string true "Reset Token"
// @Param        password formData string true "New Password"
// @Success      200  {object}  map[string]interface{}
// @Failure      400  {object}  map[string]interface{}
// @Router       /auth/reset [post]
func (h *AuthHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	// ?token=...&password=...
	token := r.FormValue("token")
	pass := r.FormValue("password")
	if err := h.svc.ResetPassword(token, pass); err != nil {
		utils.JSONError(w, "Incorrect or expired token", http.StatusBadRequest)
		return
	}
	utils.JSON(w, http.StatusOK, map[string]interface{}{"status": true, "message": "Password changed successfully"})
}
