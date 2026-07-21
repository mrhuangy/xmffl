package httpapi

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"fpxxl/backend/internal/domain"
)

type systemControlRequest struct {
	ControlKey       string     `json:"controlKey"`
	ControlGroup     string     `json:"controlGroup"`
	Name             string     `json:"name"`
	Description      string     `json:"description"`
	ValueType        string     `json:"valueType"`
	ValueText        *string    `json:"valueText"`
	ValueJSON        any        `json:"valueJson"`
	DefaultValueText *string    `json:"defaultValueText"`
	DefaultValueJSON any        `json:"defaultValueJson"`
	Enabled          bool       `json:"enabled"`
	IsPublic         bool       `json:"isPublic"`
	SortOrder        int        `json:"sortOrder"`
	EffectiveFrom    *time.Time `json:"effectiveFrom"`
	EffectiveUntil   *time.Time `json:"effectiveUntil"`
}

func (h *Handler) listSystemControls(w http.ResponseWriter, r *http.Request) {
	if !requireOwner(w, r) {
		return
	}
	controls, err := h.repo.ListSystemControls(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"controls": controls})
}

func (h *Handler) getSystemControl(w http.ResponseWriter, r *http.Request) {
	if !requireOwner(w, r) {
		return
	}
	id, err := pathUint64(r.URL.Path, "/api/admin/system-controls/")
	if err != nil {
		writeBadRequest(w, err.Error())
		return
	}
	control, err := h.repo.GetSystemControl(r.Context(), id)
	if errors.Is(err, sql.ErrNoRows) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "特殊配置不存在"})
		return
	}
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, control)
}

func (h *Handler) createSystemControl(w http.ResponseWriter, r *http.Request) {
	claims, ok := ownerClaims(w, r)
	if !ok {
		return
	}
	request, ok := decodeSystemControl(w, r)
	if !ok {
		return
	}
	control := request.toDomain()
	control.CreatedBy, control.UpdatedBy = &claims.Subject, &claims.Subject
	id, err := h.repo.CreateSystemControl(r.Context(), control)
	if err != nil {
		writeError(w, err)
		return
	}
	created, err := h.repo.GetSystemControl(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

func (h *Handler) updateSystemControl(w http.ResponseWriter, r *http.Request) {
	claims, ok := ownerClaims(w, r)
	if !ok {
		return
	}
	id, err := pathUint64(r.URL.Path, "/api/admin/system-controls/")
	if err != nil {
		writeBadRequest(w, err.Error())
		return
	}
	if _, err := h.repo.GetSystemControl(r.Context(), id); errors.Is(err, sql.ErrNoRows) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "特殊配置不存在"})
		return
	} else if err != nil {
		writeError(w, err)
		return
	}
	request, ok := decodeSystemControl(w, r)
	if !ok {
		return
	}
	control := request.toDomain()
	control.ID, control.UpdatedBy = id, &claims.Subject
	if err := h.repo.UpdateSystemControl(r.Context(), control); err != nil {
		writeError(w, err)
		return
	}
	updated, err := h.repo.GetSystemControl(r.Context(), id)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, updated)
}

func (h *Handler) deleteSystemControl(w http.ResponseWriter, r *http.Request) {
	if !requireOwner(w, r) {
		return
	}
	id, err := pathUint64(r.URL.Path, "/api/admin/system-controls/")
	if err != nil {
		writeBadRequest(w, err.Error())
		return
	}
	if _, err := h.repo.GetSystemControl(r.Context(), id); errors.Is(err, sql.ErrNoRows) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "特殊配置不存在"})
		return
	} else if err != nil {
		writeError(w, err)
		return
	}
	if err := h.repo.DeleteSystemControl(r.Context(), id); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func decodeSystemControl(w http.ResponseWriter, r *http.Request) (systemControlRequest, bool) {
	var request systemControlRequest
	if err := decodeJSON(r, &request); err != nil {
		writeBadRequest(w, err.Error())
		return request, false
	}
	request.ControlKey = strings.TrimSpace(request.ControlKey)
	request.ControlGroup = strings.TrimSpace(request.ControlGroup)
	request.Name = strings.TrimSpace(request.Name)
	if request.ControlKey == "" || request.ControlGroup == "" || request.Name == "" {
		writeBadRequest(w, "controlKey, controlGroup and name are required")
		return request, false
	}
	if request.EffectiveFrom != nil && request.EffectiveUntil != nil && request.EffectiveFrom.After(*request.EffectiveUntil) {
		writeBadRequest(w, "effectiveFrom must be before effectiveUntil")
		return request, false
	}
	if err := validateControlValue(request); err != nil {
		writeBadRequest(w, err.Error())
		return request, false
	}
	return request, true
}

func validateControlValue(request systemControlRequest) error {
	switch request.ValueType {
	case "json":
		if request.ValueJSON == nil {
			return errors.New("valueJson is required for json type")
		}
	case "bool":
		if request.ValueText == nil || (*request.ValueText != "true" && *request.ValueText != "false") {
			return errors.New("bool value must be true or false")
		}
	case "int":
		if request.ValueText == nil {
			return errors.New("valueText is required")
		}
		if _, err := strconv.ParseInt(*request.ValueText, 10, 64); err != nil {
			return errors.New("invalid integer value")
		}
	case "decimal":
		if request.ValueText == nil {
			return errors.New("valueText is required")
		}
		if _, err := strconv.ParseFloat(*request.ValueText, 64); err != nil {
			return errors.New("invalid decimal value")
		}
	case "string":
		if request.ValueText == nil {
			return errors.New("valueText is required")
		}
	default:
		return errors.New("invalid valueType")
	}
	return nil
}

func (request systemControlRequest) toDomain() domain.SystemControl {
	control := domain.SystemControl{ControlKey: request.ControlKey, ControlGroup: request.ControlGroup,
		Name: request.Name, Description: request.Description, ValueType: request.ValueType,
		Enabled: request.Enabled, IsPublic: request.IsPublic, SortOrder: request.SortOrder,
		EffectiveFrom: request.EffectiveFrom, EffectiveUntil: request.EffectiveUntil}
	if request.ValueType == "json" {
		control.ValueJSON, control.DefaultValueJSON = request.ValueJSON, request.DefaultValueJSON
	} else {
		control.ValueText, control.DefaultValueText = request.ValueText, request.DefaultValueText
	}
	return control
}
