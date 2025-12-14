package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/xeodocs/xeodocs-backend/internal/modules/auth"
	"github.com/xeodocs/xeodocs-backend/internal/modules/configurations"
	"github.com/xeodocs/xeodocs-backend/internal/modules/files"
	"github.com/xeodocs/xeodocs-backend/internal/modules/languages"
	"github.com/xeodocs/xeodocs-backend/internal/modules/paths"
	"github.com/xeodocs/xeodocs-backend/internal/modules/projects"
	"github.com/xeodocs/xeodocs-backend/internal/modules/system"
	"github.com/xeodocs/xeodocs-backend/internal/modules/user_preferences"
	"github.com/xeodocs/xeodocs-backend/internal/modules/users"
	"github.com/xeodocs/xeodocs-backend/internal/modules/workflow"
	"github.com/xeodocs/xeodocs-backend/internal/shared/config"
	"github.com/xeodocs/xeodocs-backend/internal/shared/database"
	customMiddleware "github.com/xeodocs/xeodocs-backend/internal/shared/middleware"
)

func main() {
	// Load Configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Connect to Database
	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize Router
	r := chi.NewRouter()
	r.Use(customMiddleware.RequestResponseLogger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.URLFormat)
	r.Use(middleware.AllowContentType("application/json"))

	// CORS configuration based on environment
	var allowedOrigins []string
	switch cfg.Environment {
	case "production":
		allowedOrigins = []string{"https://admin.xeodocs.com"}
	case "staging":
		allowedOrigins = []string{"https://staging-admin.xeodocs.com"}
	default: // development
		allowedOrigins = []string{"http://dev-admin.xeodocs.localhost:*"}
	}

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "X-API-Key"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))

	// Mount V1 API
	r.Route("/v1", func(r chi.Router) {
		// Initialize Modules
		authService := auth.NewService(db, cfg)
		authHandler := auth.NewHandler(authService, cfg)

		usersRepo := users.NewRepository(db)
		usersHandler := users.NewHandler(usersRepo)

		projectsRepo := projects.NewRepository(db)
		projectsHandler := projects.NewHandler(projectsRepo)

		languagesRepo := languages.NewRepository(db)
		languagesHandler := languages.NewHandler(languagesRepo)

		filesRepo := files.NewRepository(db)
		filesHandler := files.NewHandler(filesRepo)

		workflowService := workflow.NewService(db)
		workflowHandler := workflow.NewHandler(workflowService)

		configRepo := configurations.NewRepository(db)
		configHandler := configurations.NewHandler(configRepo)

		prefsRepo := user_preferences.NewRepository(db)
		prefsHandler := user_preferences.NewHandler(prefsRepo)

		pathsRepo := paths.NewRepository(db)
		pathsHandler := paths.NewHandler(pathsRepo)

		systemHandler := system.NewHandler()

		// Public Routes
		r.Group(func(r chi.Router) {
			systemHandler.RegisterRoutes(r)
			authHandler.RegisterRoutes(r)
		})

		// Protected Routes
		r.Group(func(r chi.Router) {
			r.Use(authHandler.AuthMiddleware)

			usersHandler.RegisterRoutes(r)
			projectsHandler.RegisterRoutes(r)
			languagesHandler.RegisterRoutes(r)
			filesHandler.RegisterRoutes(r)
			workflowHandler.RegisterRoutes(r)
			configHandler.RegisterRoutes(r)
			prefsHandler.RegisterRoutes(r)
			pathsHandler.RegisterRoutes(r)
		})
	})

	log.Printf("Server starting on port %s (with Custom Logger). Env: %s", cfg.Port, cfg.Environment)
	if err := http.ListenAndServe(":"+cfg.Port, r); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
