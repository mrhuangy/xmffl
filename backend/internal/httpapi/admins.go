package httpapi

import (
	"database/sql"
	"errors"
	"net/http"
	"strings"

	"fpxxl/backend/internal/domain"
	"golang.org/x/crypto/bcrypt"
)

type adminMutationRequest struct {
	Username    string         `json:"username"`
	Email       *string        `json:"email"`
	Password    string         `json:"password"`
	DisplayName string         `json:"displayName"`
	Role        string         `json:"role"`
	Permissions map[string]any `json:"permissions"`
	Status      string         `json:"status"`
}

func (h *Handler) listAdmins(w http.ResponseWriter, r *http.Request) {
	if !requireOwner(w, r) {
		return
	}
	admins, err := h.repo.ListAdmins(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"admins": admins})
}

func (h *Handler) getAdmin(w http.ResponseWriter, r *http.Request) {
	if !requireOwner(w, r) {
		return
	}
	id, err := pathUint64(r.URL.Path, "/api/admin/users/")
	if err != nil {
		writeBadRequest(w, err.Error())
		return
	}
	admin, err := h.repo.GetAdmin(r.Context(), id)
	if errors.Is(err, sql.ErrNoRows) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "管理员不存在"})
		return
	}
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, admin)
}

func (h *Handler) createAdmin(w http.ResponseWriter, r *http.Request) {
	claims, ok := ownerClaims(w, r)
	if !ok {
		return
	}
	var request adminMutationRequest
	if err := decodeJSON(r, &request); err != nil {
		writeBadRequest(w, err.Error())
		return
	}
	if err := validateAdminRequest(request, true); err != nil {
		writeBadRequest(w, err.Error())
		return
	}
	if _, err := h.repo.FindAdminByUsername(r.Context(), request.Username); err == nil {
		writeJSON(w, http.StatusConflict, map[string]string{"error": "用户名已存在"})
		return
	} else if !errors.Is(err, sql.ErrNoRows) {
		writeError(w, err)
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
	if err != nil {
		writeError(w, err)
		return
	}
	admin := mutationToAdmin(request)
	admin.PasswordHash = string(hash)
	admin.CreatedBy = &claims.Subject
	id, err := h.repo.CreateAdmin(r.Context(), admin)
	if err != nil {
		writeError(w, err)
		return
	}
	created, err := h.repo.GetAdmin(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (h *Handler) updateAdmin(w http.ResponseWriter, r *http.Request) {
	claims, ok := ownerClaims(w, r)
	if !ok {
		return
	}
	id, err := pathUint64(r.URL.Path, "/api/admin/users/")
	if err != nil {
		writeBadRequest(w, err.Error())
		return
	}
	current, err := h.repo.GetAdmin(r.Context(), id)
	if errors.Is(err, sql.ErrNoRows) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "管理员不存在"})
		return
	}
	if err != nil {
		writeError(w, err)
		return
	}
	var request adminMutationRequest
	if err := decodeJSON(r, &request); err != nil {
		writeBadRequest(w, err.Error())
		return
	}
	if err := validateAdminRequest(request, false); err != nil {
		writeBadRequest(w, err.Error())
		return
	}
	if id == claims.Subject && request.Status != "active" {
		writeBadRequest(w, "不能禁用当前登录账号")
		return
	}
	if current.Role == "owner" && current.Status == "active" && (request.Role != "owner" || request.Status != "active") {
		if blocked, err := h.isLastOwner(r, w); err != nil || blocked {
			return
		}
	}
	if existing, err := h.repo.FindAdminByUsername(r.Context(), request.Username); err == nil && existing.ID != id {
		writeJSON(w, http.StatusConflict, map[string]string{"error": "用户名已存在"})
		return
	} else if err != nil && !errors.Is(err, sql.ErrNoRows) {
		writeError(w, err)
		return
	}
	admin := mutationToAdmin(request)
	admin.ID = id
	updatePassword := request.Password != ""
	if updatePassword {
		hash, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
		if err != nil {
			writeError(w, err)
			return
		}
		admin.PasswordHash = string(hash)
	}
	if err := h.repo.UpdateAdmin(r.Context(), admin, updatePassword); err != nil {
		writeError(w, err)
		return
	}
	updated, err := h.repo.GetAdmin(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

func (h *Handler) deleteAdmin(w http.ResponseWriter, r *http.Request) {
	claims, ok := ownerClaims(w, r)
	if !ok {
		return
	}
	id, err := pathUint64(r.URL.Path, "/api/admin/users/")
	if err != nil {
		writeBadRequest(w, err.Error())
		return
	}
	if id == claims.Subject {
		writeBadRequest(w, "不能删除当前登录账号")
		return
	}
	admin, err := h.repo.GetAdmin(r.Context(), id)
	if errors.Is(err, sql.ErrNoRows) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "管理员不存在"})
		return
	}
	if err != nil {
		writeError(w, err)
		return
	}
	if admin.Role == "owner" && admin.Status == "active" {
		if blocked, err := h.isLastOwner(r, w); err != nil || blocked {
			return
		}
	}
	if err := h.repo.DeleteAdmin(r.Context(), id); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func mutationToAdmin(request adminMutationRequest) domain.AdminUser {
	request.Username = strings.TrimSpace(request.Username)
	request.DisplayName = strings.TrimSpace(request.DisplayName)
	if request.Email != nil {
		value := strings.TrimSpace(*request.Email)
		if value == "" {
			request.Email = nil
		} else {
			request.Email = &value
		}
	}
	return domain.AdminUser{Username: request.Username, Email: request.Email, DisplayName: request.DisplayName,
		Role: request.Role, Permissions: request.Permissions, Status: request.Status}
}

func validateAdminRequest(request adminMutationRequest, passwordRequired bool) error {
	if strings.TrimSpace(request.Username) == "" {
		return errors.New("username is required")
	}
	if passwordRequired && request.Password == "" {
		return errors.New("password is required")
	}
	if request.Password != "" && len(request.Password) < 10 {
		return errors.New("password must contain at least 10 characters")
	}
	if request.Role != "owner" && request.Role != "operator" && request.Role != "viewer" {
		return errors.New("invalid role")
	}
	if request.Status != "active" && request.Status != "disabled" {
		return errors.New("invalid status")
	}
	return nil
}

func requireOwner(w http.ResponseWriter, r *http.Request) bool { _, ok := ownerClaims(w, r); return ok }

func ownerClaims(w http.ResponseWriter, r *http.Request) (adminClaims, bool) {
	claims, ok := r.Context().Value(claimsContextKey{}).(adminClaims)
	if !ok || claims.Role != "owner" {
		writeJSON(w, http.StatusForbidden, map[string]string{"error": "仅超级管理员可执行此操作"})
		return claims, false
	}
	return claims, true
}

func (h *Handler) isLastOwner(r *http.Request, w http.ResponseWriter) (bool, error) {
	count, err := h.repo.CountActiveOwners(r.Context())
	if err != nil {
		writeError(w, err)
		return false, err
	}
	if count <= 1 {
		writeBadRequest(w, "系统必须至少保留一个可用的超级管理员")
		return true, nil
	}
	return false, nil
}
