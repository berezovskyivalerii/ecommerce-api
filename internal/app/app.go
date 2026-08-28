// Package app contains build application point
package app

import (
	"context"
	"database/sql"
	"log"
	"net/http"

	"github.com/berezovskyivalerii/ecommerce-api/internal/config"
	httprest "github.com/berezovskyivalerii/ecommerce-api/internal/delivery/http"
	"github.com/berezovskyivalerii/ecommerce-api/internal/delivery/http/middleware"
	hasher "github.com/berezovskyivalerii/ecommerce-api/internal/pkg/hash"
	jwtmanager "github.com/berezovskyivalerii/ecommerce-api/internal/pkg/jwt"
	"github.com/berezovskyivalerii/ecommerce-api/internal/repository/postgres"
	"github.com/berezovskyivalerii/ecommerce-api/internal/service"
	"github.com/gin-gonic/gin"

	_ "github.com/lib/pq"
)

func setupRouter(
	userHandler *httprest.UserHandlers,
	refreshHandler *httprest.RefreshTokensHandlers,
	categoriesHandler *httprest.CategoriesHandlers,
	productsHandler *httprest.ProductsHandlers,
) *gin.Engine {
	r := gin.Default()

	r.POST("/api/register", userHandler.Register)
	r.POST("/api/login", userHandler.Login)

	r.POST("/api/refresh", refreshHandler.RefreshToken)
	r.POST("/api/revoke", refreshHandler.RevokeToken)

	r.GET("/api/categories", categoriesHandler.GetCategories)
	r.GET("/api/categories/:category_id/products", productsHandler.GetProductsByCategory)

	r.GET("/api/products", productsHandler.GetProducts)
	r.GET("/api/products/search", productsHandler.SearchProducts)

	api := r.Group("/api")
	api.Use(middleware.AuthMiddleware("secret-123"))
	{
		adminRoutes := api.Group("/admin")
		adminRoutes.Use(middleware.RequireRole("admin"))
		{
			adminRoutes.GET("/users", userHandler.GetUsers)
			adminRoutes.PUT("/users/:id", userHandler.UpdateUser)
			adminRoutes.DELETE("/users/:id", userHandler.DeleteUser)

			adminRoutes.POST("/categories", categoriesHandler.CreateCategory)
			adminRoutes.PUT("/categories/:id", categoriesHandler.UpdateCategory)
			adminRoutes.DELETE("/categories/:id", categoriesHandler.DeleteCategory)

			adminRoutes.POST("/products", productsHandler.CreateProduct)
			adminRoutes.PUT("/products/:id", productsHandler.UpdateProduct)
			adminRoutes.DELETE("/products/:id", productsHandler.DeleteProduct)
		}
	}

	return r
}

type App struct {
	config *config.Config
	server *http.Server
	db     *sql.DB
}

func New(cfg *config.Config) *App {
	db, err := sql.Open("postgres", cfg.DBURL)
	if err != nil {
		log.Fatalf("error opening database connection: %v", err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatalf("error connecting to database: %v", err)
	}

	// 1. Utilities
	passwordHasher := hasher.New()
	jwtManager := jwtmanager.New()

	// 3. Repository
	store := postgres.NewStore(db)

	userRepo := postgres.NewUserRepository(store)
	refreshTokensRepo := postgres.NewRefreshTokensRepository(store)
	categoriesRepo := postgres.NewCategoryRepository(store)
	productsRepo := postgres.NewProductsRepository(store)

	service.SeedAdmin(context.Background(), userRepo, passwordHasher, cfg.AdminEmail, cfg.AdminPassword)

	// 4. Service
	userService := service.NewUserService(userRepo, passwordHasher, jwtManager, refreshTokensRepo)
	refreshService := service.NewRefreshTokensService(refreshTokensRepo, userRepo, jwtManager)
	categoriesService := service.NewCategoriesService(categoriesRepo)
	productsService := service.NewProductsService(productsRepo, categoriesRepo)

	// 5. Handler
	userHandler := httprest.NewUserHandlers(userService)
	refreshHandler := httprest.NewRefreshTokensHandlers(refreshService)
	categoriesHandler := httprest.NewCategoriesHandlers(categoriesService)
	productsHandler := httprest.NewProductsHandlers(productsService)

	// 6. Routing
	router := setupRouter(userHandler, refreshHandler, categoriesHandler, productsHandler)

	// 7. Build
	return &App{
		config: cfg,
		db:     db,
		server: &http.Server{
			Addr:    cfg.ServerPort,
			Handler: router,
		},
	}
}

func (a *App) Run() error {
	log.Printf("Server is running on port %s", a.config.ServerPort)
	if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

func (a *App) Close() {
	if a.db != nil {
		a.db.Close()
	}
}
