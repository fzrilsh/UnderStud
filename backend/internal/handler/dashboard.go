package handler

import (
	"net/http"

	"backend/internal/service"
)

type DashboardHandlers struct {
	dashboard *service.Dashboard
}

func NewDashboardHandlers(d *service.Dashboard) *DashboardHandlers {
	return &DashboardHandlers{dashboard: d}
}

// BK handles GET /api/v1/dashboard/bk.
func (h *DashboardHandlers) BK(w http.ResponseWriter, r *http.Request) {
	actor, err := actorFrom(r)
	if err != nil {
		Error(w, err)
		return
	}
	data, err := h.dashboard.BK(r.Context(), actor)
	if err != nil {
		Error(w, err)
		return
	}
	OK(w, data)
}

// WK handles GET /api/v1/dashboard/wk.
func (h *DashboardHandlers) WK(w http.ResponseWriter, r *http.Request) {
	actor, err := actorFrom(r)
	if err != nil {
		Error(w, err)
		return
	}
	data, err := h.dashboard.WK(r.Context(), actor)
	if err != nil {
		Error(w, err)
		return
	}
	OK(w, data)
}

// Classes handles GET /api/v1/classes.
func (h *DashboardHandlers) Classes(w http.ResponseWriter, r *http.Request) {
	actor, err := actorFrom(r)
	if err != nil {
		Error(w, err)
		return
	}
	data, err := h.dashboard.Classes(r.Context(), actor)
	if err != nil {
		Error(w, err)
		return
	}
	OK(w, data)
}
