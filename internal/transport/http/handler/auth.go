package handler

import (
    "encoding/json"
    "net/http"

    "pvz-service/internal/usecase/auth"
)

type AuthHandler struct {
    authService *auth.Service
}

func NewAuthHandler(authService *auth.Service) *AuthHandler {
    return &AuthHandler{authService: authService}
}

func apiRoleToInternal(role string) (string, bool) {
    switch role {
    case "employee":
        return "employee", true
    case "moderator":
        return "moderator", true
    default:
        return "", false
    }
}

func internalRoleToAPI(role string) string {
    if role == "client" {
        return "employee"
    }
    return role
}

func (h *AuthHandler) DummyLogin(w http.ResponseWriter, r *http.Request) {
    var req struct {
        Role string `json:"role"`
    }
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Bad Request", http.StatusBadRequest)
        return
    }

    internalRole, ok := apiRoleToInternal(req.Role)
    if !ok {
        http.Error(w, "Bad Request", http.StatusBadRequest)
        return
    }

    token, err := h.authService.DummyLogin(r.Context(), internalRole)
    if err != nil {
        http.Error(w, "Bad Request", http.StatusBadRequest)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    _ = json.NewEncoder(w).Encode(token.Token)
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
    var req struct {
        Email    string `json:"email"`
        Password string `json:"password"`
        Role     string `json:"role"`
    }
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Bad Request", http.StatusBadRequest)
        return
    }

    internalRole, ok := apiRoleToInternal(req.Role)
    if !ok {
        http.Error(w, "Bad Request", http.StatusBadRequest)
        return
    }

    result, err := h.authService.Register(r.Context(), req.Email, req.Password, internalRole)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    resp := struct {
        ID    string `json:"id"`
        Email string `json:"email"`
        Role  string `json:"role"`
    }{
        ID:    result.UserID,
        Email: req.Email,
        Role:  internalRoleToAPI(internalRole),
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    _ = json.NewEncoder(w).Encode(resp)
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
    var req struct {
        Email    string `json:"email"`
        Password string `json:"password"`
    }
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Bad Request", http.StatusBadRequest)
        return
    }

    token, err := h.authService.Login(r.Context(), req.Email, req.Password)
    if err != nil {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    _ = json.NewEncoder(w).Encode(token.Token)
}
