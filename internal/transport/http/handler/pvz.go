package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"pvz-service/internal/usecase/pvz"
	receptionuc "pvz-service/internal/usecase/reception"

	"github.com/go-chi/chi/v5"
)

type PVZHandler struct {
	pvzService       *pvz.Service
	receptionService *receptionuc.Service
}

func NewPVZHandler(pvzService *pvz.Service, receptionService *receptionuc.Service) *PVZHandler {
	return &PVZHandler{pvzService: pvzService, receptionService: receptionService}
}

func productTypeAPIToInternal(t string) (string, bool) {
	switch t {
	case "электроника":
		return "electronics", true
	case "одежда":
		return "clothes", true
	case "обувь":
		return "shoes", true
	default:
		return "", false
	}
}

func productTypeInternalToAPI(t string) string {
	switch t {
	case "electronics":
		return "электроника"
	case "clothes":
		return "одежда"
	case "shoes":
		return "обувь"
	default:
		return t
	}
}

func receptionStatusInternalToAPI(s string) string {
	if s == "closed" {
		return "close"
	}
	return s
}

type apiPVZ struct {
	ID               string    `json:"id"`
	RegistrationDate time.Time `json:"registrationDate"`
	City             string    `json:"city"`
}

type apiReception struct {
	ID       string    `json:"id"`
	DateTime time.Time `json:"dateTime"`
	PVZID    string    `json:"pvzId"`
	Status   string    `json:"status"`
}

type apiProduct struct {
	ID          string    `json:"id"`
	DateTime    time.Time `json:"dateTime"`
	Type        string    `json:"type"`
	ReceptionID string    `json:"receptionId"`
}

type apiPVZListItem struct {
	PVZ        apiPVZ `json:"pvz"`
	Receptions []struct {
		Reception apiReception `json:"reception"`
		Products  []apiProduct `json:"products"`
	} `json:"receptions"`
}

func (h *PVZHandler) CreatePVZ(w http.ResponseWriter, r *http.Request) {
	var req struct {
		City string `json:"city"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.City == "" {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	result, err := h.pvzService.Create(r.Context(), req.City)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp := apiPVZ{
		ID:               result.ID,
		RegistrationDate: result.CreatedAt,
		City:             result.City,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *PVZHandler) ListPVZ(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	var startDate, endDate *time.Time
	if s := q.Get("startDate"); s != "" {
		tm, err := time.Parse(time.RFC3339Nano, s)
		if err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}
		startDate = &tm
	}
	if s := q.Get("endDate"); s != "" {
		tm, err := time.Parse(time.RFC3339Nano, s)
		if err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}
		endDate = &tm
	}

	page := 1
	if p := q.Get("page"); p != "" {
		v, err := strconv.Atoi(p)
		if err != nil || v < 1 {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}
		page = v
	}

	limit := 10
	if l := q.Get("limit"); l != "" {
		v, err := strconv.Atoi(l)
		if err != nil || v < 1 || v > 30 {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}
		limit = v
	}
	offset := (page - 1) * limit

	result, err := h.pvzService.List(r.Context(), startDate, endDate, limit, offset)
	if err != nil {
		http.Error(w, "Internal Error", http.StatusInternalServerError)
		return
	}

	resp := make([]apiPVZListItem, 0, len(result))
	for _, p := range result {
		item := apiPVZListItem{
			PVZ: apiPVZ{
				ID:               p.ID,
				RegistrationDate: p.CreatedAt,
				City:             p.City,
			},
		}

		for _, rcv := range p.Receptions {
			block := struct {
				Reception apiReception `json:"reception"`
				Products  []apiProduct `json:"products"`
			}{
				Reception: apiReception{
					ID:       rcv.ID,
					DateTime: rcv.StartedAt,
					PVZID:    p.ID,
					Status:   receptionStatusInternalToAPI(rcv.Status),
				},
			}

			for _, pr := range rcv.Products {
				block.Products = append(block.Products, apiProduct{
					ID:          pr.ID,
					DateTime:    pr.AddedAt,
					Type:        productTypeInternalToAPI(pr.Type),
					ReceptionID: rcv.ID,
				})
			}

			item.Receptions = append(item.Receptions, block)
		}

		resp = append(resp, item)
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *PVZHandler) CreateReception(w http.ResponseWriter, r *http.Request) {
	var req struct {
		PVZID string `json:"pvzId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.PVZID == "" {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	rec, err := h.receptionService.Open(r.Context(), req.PVZID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp := apiReception{
		ID:       rec.ID,
		DateTime: rec.StartedAt,
		PVZID:    rec.PVZID,
		Status:   "in_progress",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *PVZHandler) AddProduct(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Type  string `json:"type"`
		PVZID string `json:"pvzId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Type == "" || req.PVZID == "" {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	internalType, ok := productTypeAPIToInternal(req.Type)
	if !ok {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	pr, err := h.receptionService.AddProduct(r.Context(), req.PVZID, internalType)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp := apiProduct{
		ID:          pr.ID,
		DateTime:    pr.AddedAt,
		Type:        productTypeInternalToAPI(pr.Type),
		ReceptionID: pr.ReceptionID,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *PVZHandler) CloseLastReception(w http.ResponseWriter, r *http.Request) {
	pvzID := chi.URLParam(r, "pvzId")
	if pvzID == "" {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	rec, err := h.receptionService.Close(r.Context(), pvzID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	resp := apiReception{
		ID:       rec.ID,
		DateTime: rec.StartedAt,
		PVZID:    rec.PVZID,
		Status:   "close",
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func (h *PVZHandler) DeleteLastProduct(w http.ResponseWriter, r *http.Request) {
	pvzID := chi.URLParam(r, "pvzId")
	if pvzID == "" {
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	_, err := h.receptionService.RemoveProduct(r.Context(), pvzID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}
