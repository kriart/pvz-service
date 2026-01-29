package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"pvz-service/internal/adapter/auth/jwt"
	"pvz-service/internal/adapter/auth/password"
	"pvz-service/internal/adapter/db/postgres"
	"pvz-service/internal/adapter/observability/logging"
	"pvz-service/internal/adapter/observability/metrics"
	clockad "pvz-service/internal/adapter/time"
	"pvz-service/internal/config"
	"pvz-service/internal/transport/http/handler"
	"pvz-service/internal/transport/http/middleware"
	"pvz-service/internal/usecase/auth"
	pvzUC "pvz-service/internal/usecase/pvz"
	recvUC "pvz-service/internal/usecase/reception"
)

type App struct {
	cfg           config.Config
	httpServer    *http.Server
	metricsServer *http.Server
	db            *postgres.PostgresDB
}

func New(cfg config.Config) (*App, error) {
	logging.InitLogger()

	db, err := postgres.NewDB(cfg.DB.Host, cfg.DB.Port, cfg.DB.User, cfg.DB.Password, cfg.DB.Name)
	if err != nil {
		return nil, err
	}

	tokenManager := jwt.NewTokenManagerJWT(cfg.JWT.Secret)
	passwordHasher := password.NewHasher()
	clock := clockad.RealClock{}

	userRepo := db.UserRepo()
	pvzRepo := db.PVZRepo()
	receptionRepo := db.ReceptionRepo()
	productRepo := db.ProductRepo()

	authService := auth.NewService(userRepo, tokenManager, passwordHasher, clock)
	pvzService := pvzUC.NewService(pvzRepo, receptionRepo, productRepo, clock)
	receptionService := recvUC.NewService(pvzRepo, receptionRepo, productRepo, clock)

	authHandler := handler.NewAuthHandler(authService)
	pvzHandler := handler.NewPVZHandler(pvzService, receptionService)

	metricsCollector := metrics.NewPromMetrics()

	r := chi.NewRouter()

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"http://localhost:8081", "http://127.0.0.1:8081"},
		AllowedMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Accept", "Authorization", "Content-Type"},
		MaxAge:         300,
	}))

	r.Use(middleware.LoggingMiddleware)
	r.Use(middleware.MetricsMiddleware(metricsCollector))

	r.Post("/dummyLogin", authHandler.DummyLogin)
	r.Post("/register", authHandler.Register)
	r.Post("/login", authHandler.Login)

	r.Group(func(pr chi.Router) {
		pr.Use(middleware.AuthMiddleware(tokenManager))

		pr.With(middleware.RequireRole("moderator")).Post("/pvz", pvzHandler.CreatePVZ)
		pr.With(middleware.RequireRole("employee", "moderator")).Get("/pvz", pvzHandler.ListPVZ)

		pr.With(middleware.RequireRole("employee")).Post("/receptions", pvzHandler.CreateReception)
		pr.With(middleware.RequireRole("employee")).Post("/products", pvzHandler.AddProduct)

		pr.With(middleware.RequireRole("employee")).Post("/pvz/{pvzId}/close_last_reception", pvzHandler.CloseLastReception)
		pr.With(middleware.RequireRole("employee")).Post("/pvz/{pvzId}/delete_last_product", pvzHandler.DeleteLastProduct)
	})

	httpAddr := fmt.Sprintf(":%d", cfg.Server.HTTPPort)
	httpSrv := &http.Server{
		Addr:              httpAddr,
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
	}

	metricsAddr := fmt.Sprintf(":%d", cfg.Server.MetricsPort)
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	metricsSrv := &http.Server{
		Addr:              metricsAddr,
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	return &App{
		cfg:           cfg,
		httpServer:    httpSrv,
		metricsServer: metricsSrv,
		db:            db,
	}, nil
}

func (a *App) Run(ctx context.Context) error {
	go func() {
		_ = a.metricsServer.ListenAndServe()
	}()
	return a.httpServer.ListenAndServe()
}

func (a *App) Shutdown(ctx context.Context) error {
	defer logging.Logger.Sync()
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	err := a.httpServer.Shutdown(ctx)
	_ = a.metricsServer.Shutdown(ctx)
	if a.db != nil {
		a.db.Close()
	}
	return err
}
