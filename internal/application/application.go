package application

import (
	"context"
	"log"
	"time"

	"github.com/MostajeranMohammad/dekamond-auth-challenge/config"
	_ "github.com/MostajeranMohammad/dekamond-auth-challenge/docs"
	"github.com/MostajeranMohammad/dekamond-auth-challenge/internal/controllers"
	"github.com/MostajeranMohammad/dekamond-auth-challenge/internal/guards"
	"github.com/MostajeranMohammad/dekamond-auth-challenge/internal/repositories"
	"github.com/MostajeranMohammad/dekamond-auth-challenge/internal/routes"
	"github.com/MostajeranMohammad/dekamond-auth-challenge/internal/usecases"
	"github.com/MostajeranMohammad/dekamond-auth-challenge/pkg/logger"
	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/redis/go-redis/v9"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title Dekamond Auth Challenge API
// @version 1.0
// @description OTP-based auth service with users listing
// @securityDefinitions.apikey	BearerAuth
// @in							header
// @name						Authorization
func Run(cfg *config.Config) {
	l, zapLogger := logger.New(cfg.Log.Level)

	db, err := pgxpool.New(context.Background(), cfg.PG.DSN)
	if err != nil {
		l.Fatal(err)
	}
	defer db.Close()

	if err := db.Ping(context.Background()); err != nil {
		l.Fatal("Failed to ping the database:", err)
	}

	if cfg.PG.RunMigrations {
		func() {
			sqlDB := stdlib.OpenDB(*db.Config().ConnConfig)
			defer sqlDB.Close()

			driver, err := postgres.WithInstance(sqlDB, &postgres.Config{})
			if err != nil {
				l.Fatal("Failed to create driver:", err)
			}
			defer driver.Close()

			m, err := migrate.NewWithDatabaseInstance(
				"file://database/migrations",
				"postgres",
				driver,
			)
			if err != nil {
				l.Fatal("Failed to create migrations", err)
			}

			err = m.Up()
			if err != nil {
				if err.Error() != "no change" {
					l.Fatal("Failed to run migrations:", err)
				}
				l.Info("No migrations to run")
			}
		}()
	}

	redisDB := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	status := redisDB.Ping(context.Background())
	if status.Err() != nil {
		log.Fatalln("Failed to connect to redis:", status.Err())
	}

	userRepository := repositories.NewUserRepository(db)

	jwtUsecase := usecases.NewJwtUsecase(cfg.AUTH.JwtSecret)
	otpUsecase := usecases.NewOtpUsecase(redisDB, l)
	authUsecase := usecases.NewAuthUsecase(userRepository, jwtUsecase, cfg, otpUsecase)
	usersService := usecases.NewUsersService(userRepository)

	authController := controllers.NewAuthController(l, authUsecase)
	usersController := controllers.NewUsersController(l, usersService)

	authGuard := guards.NewAuthGuard(authUsecase)

	ginApp := gin.New()
	ginApp.Use(ginzap.Ginzap(zapLogger, time.RFC3339, true))

	v1 := ginApp.Group("/api/v1")

	routes.RegisterAuthV1Router(v1, authController, authGuard)
	routes.RegisterUserV1Router(v1, usersController, authGuard)
	ginApp.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	if err := ginApp.Run(":" + cfg.HTTP.Port); err != nil {
		l.Fatal("ginApp.Run failed", err)
	}

}
