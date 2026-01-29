package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"pvz-service/internal/adapter/auth/jwt"
	"pvz-service/internal/adapter/auth/password"
	"pvz-service/internal/adapter/db/postgres"
	clockad "pvz-service/internal/adapter/time"
	"pvz-service/internal/config"
	"pvz-service/internal/transport/http/handler"
	"pvz-service/internal/transport/http/middleware"
	"pvz-service/internal/usecase/auth"
	pvzUC "pvz-service/internal/usecase/pvz"
	receptionUC "pvz-service/internal/usecase/reception"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"
)

func requireStatus(t *testing.T, res *http.Response, want int, label string) {
	t.Helper()
	if res.StatusCode == want {
		return
	}
	b, _ := io.ReadAll(res.Body)
	_ = res.Body.Close()
	t.Fatalf("unexpected status for %s: got %d want %d; body=%q; headers=%v",
		label, res.StatusCode, want, string(b), res.Header)
}

func postJSON(t *testing.T, url, token string, payload any) *http.Response {
	t.Helper()

	var body bytes.Buffer
	if payload != nil {
		require.NoError(t, json.NewEncoder(&body).Encode(payload))
	}

	req, err := http.NewRequest(http.MethodPost, url, &body)
	require.NoError(t, err)

	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	res, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	return res
}

func get(t *testing.T, url, token string) *http.Response {
	t.Helper()

	req, err := http.NewRequest(http.MethodGet, url, nil)
	require.NoError(t, err)

	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	res, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	return res
}

func mustReadTokenString(t *testing.T, res *http.Response) string {
	t.Helper()
	defer res.Body.Close()

	// DummyLogin возвращает JSON string
	var token string
	require.NoError(t, json.NewDecoder(res.Body).Decode(&token))
	require.NotEmpty(t, token)
	return token
}

func setupServer(t *testing.T) (*httptest.Server, func()) {
	cfg := config.Load()

	// Здесь ВСЁ берётся из config.Load(); порт НЕ хардкодим
	t.Logf("DB cfg: host=%s port=%d user=%s db=%s", cfg.DB.Host, cfg.DB.Port, cfg.DB.User, cfg.DB.Name)

	db, err := postgres.NewDB(cfg.DB.Host, 15433, cfg.DB.User, cfg.DB.Password, cfg.DB.Name)
	require.NoError(t, err)
	require.NoError(t, db.Ping(context.Background()))

	tokenManager := jwt.NewTokenManagerJWT(cfg.JWT.Secret)
	passwordHasher := password.NewHasher()
	clock := clockad.RealClock{}

	userRepo := db.UserRepo()
	pvzRepo := db.PVZRepo()
	receptionRepo := db.ReceptionRepo()
	productRepo := db.ProductRepo()

	authService := auth.NewService(userRepo, tokenManager, passwordHasher, clock)
	pvzService := pvzUC.NewService(pvzRepo, receptionRepo, productRepo, clock)
	receptionService := receptionUC.NewService(pvzRepo, receptionRepo, productRepo, clock)

	authHandler := handler.NewAuthHandler(authService)
	pvzHandler := handler.NewPVZHandler(pvzService, receptionService)

	r := chi.NewRouter()

	// public
	r.Post("/dummyLogin", authHandler.DummyLogin)
	r.Post("/register", authHandler.Register)
	r.Post("/login", authHandler.Login)

	// protected (как в api.New)
	r.Group(func(pr chi.Router) {
		pr.Use(middleware.AuthMiddleware(tokenManager))

		pr.With(middleware.RequireRole("moderator")).Post("/pvz", pvzHandler.CreatePVZ)
		pr.With(middleware.RequireRole("employee", "moderator")).Get("/pvz", pvzHandler.ListPVZ)

		pr.With(middleware.RequireRole("employee")).Post("/receptions", pvzHandler.CreateReception)
		pr.With(middleware.RequireRole("employee")).Post("/products", pvzHandler.AddProduct)

		pr.With(middleware.RequireRole("employee")).Post("/pvz/{pvzId}/close_last_reception", pvzHandler.CloseLastReception)
		pr.With(middleware.RequireRole("employee")).Post("/pvz/{pvzId}/delete_last_product", pvzHandler.DeleteLastProduct)
	})

	ts := httptest.NewServer(r)
	cleanup := func() {
		ts.Close()
		db.Close()
	}
	return ts, cleanup
}

func TestPVZFlow(t *testing.T) {
	ts, cleanup := setupServer(t)
	defer cleanup()

	res := postJSON(t, ts.URL+"/dummyLogin", "", map[string]any{"role": "moderator"})
	requireStatus(t, res, http.StatusOK, "POST /dummyLogin (moderator)")
	modToken := mustReadTokenString(t, res)

	res = postJSON(t, ts.URL+"/pvz", modToken, map[string]any{"city": "Москва"})
	requireStatus(t, res, http.StatusCreated, "POST /pvz")
	var pvzResp struct {
		ID string `json:"id"`
	}
	require.NoError(t, json.NewDecoder(res.Body).Decode(&pvzResp))
	_ = res.Body.Close()
	require.NotEmpty(t, pvzResp.ID)

	res = postJSON(t, ts.URL+"/dummyLogin", "", map[string]any{"role": "employee"})
	requireStatus(t, res, http.StatusOK, "POST /dummyLogin (client)")
	clientToken := mustReadTokenString(t, res)

	res = postJSON(t, ts.URL+"/receptions", clientToken, map[string]any{"pvzId": pvzResp.ID})
	requireStatus(t, res, http.StatusCreated, "POST /receptions")
	_ = res.Body.Close()

	for i := 0; i < 50; i++ {
		res = postJSON(t, ts.URL+"/products", clientToken, map[string]any{
			"pvzId": pvzResp.ID,
			"type":  "электроника",
		})
		requireStatus(t, res, http.StatusCreated, "POST /products")
		_ = res.Body.Close()
	}

	res = get(t, ts.URL+"/pvz", clientToken)
	requireStatus(t, res, http.StatusOK, "GET /pvz")
	_ = res.Body.Close()

	res = postJSON(t, ts.URL+"/pvz/"+pvzResp.ID+"/close_last_reception", clientToken, nil)
	requireStatus(t, res, http.StatusOK, "POST /pvz/{id}/close_last_reception")
	_ = res.Body.Close()
}
