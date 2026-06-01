package app

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	appmiddleware "github.com/SIniutin/history-app-backend/internal/app/middleware"
	"github.com/SIniutin/history-app-backend/internal/config"
	"github.com/SIniutin/history-app-backend/internal/logger"
	content_infra "github.com/SIniutin/history-app-backend/internal/modules/content/infra"
	content_repo "github.com/SIniutin/history-app-backend/internal/modules/content/repo/postgre"
	content_usecase "github.com/SIniutin/history-app-backend/internal/modules/content/usecase"
	gamification_infra "github.com/SIniutin/history-app-backend/internal/modules/gamification/infra"
	gamification_repo "github.com/SIniutin/history-app-backend/internal/modules/gamification/repository/postgres"
	gamification_usecase "github.com/SIniutin/history-app-backend/internal/modules/gamification/usecase"
	learning_infra "github.com/SIniutin/history-app-backend/internal/modules/learning/infra"
	learning_repo "github.com/SIniutin/history-app-backend/internal/modules/learning/repo/postgre"
	learning_usecase "github.com/SIniutin/history-app-backend/internal/modules/learning/usecase"
	media_infra "github.com/SIniutin/history-app-backend/internal/modules/media/infra"
	progress_infra "github.com/SIniutin/history-app-backend/internal/modules/progress/infra"
	progress_repo "github.com/SIniutin/history-app-backend/internal/modules/progress/repo/postgre"
	progress_usecase "github.com/SIniutin/history-app-backend/internal/modules/progress/usecase"
	recommendation_infra "github.com/SIniutin/history-app-backend/internal/modules/recommendation/infra"
	recommendation_usecase "github.com/SIniutin/history-app-backend/internal/modules/recommendation/usecase"
	users_domain "github.com/SIniutin/history-app-backend/internal/modules/users/domain"
	users_infra "github.com/SIniutin/history-app-backend/internal/modules/users/infra"
	users_repo "github.com/SIniutin/history-app-backend/internal/modules/users/repo/postgre"
	users_security "github.com/SIniutin/history-app-backend/internal/modules/users/security"
	users_usecase "github.com/SIniutin/history-app-backend/internal/modules/users/usecase"
	platformdb "github.com/SIniutin/history-app-backend/internal/platform/db"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

type App struct {
	userHandler           *users_infra.Handler
	contentHandler        *content_infra.Handler
	learningHandler       *learning_infra.Handler
	mediaHandler          *media_infra.Handler
	gamificationHandler   *gamification_infra.Handler
	progressHandler       *progress_infra.Handler
	recommendationHandler *recommendation_infra.Handler
	cfg                   config.Config
	logger                *zap.Logger
	db                    *pgxpool.Pool
}

func New(ctx context.Context) (*App, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}
	log, err := logger.New(cfg.App.Env, cfg.Log.Level)
	if err != nil {
		return nil, err
	}

	pool, err := platformdb.NewPool(ctx, cfg.DB)
	if err != nil {
		_ = log.Sync()
		return nil, err
	}

	authMiddleware := users_infra.NewAuthMiddleware(
		[]byte(cfg.Auth.AccessSecret),
		cfg.Auth.Issuer,
		cfg.Auth.Audience,
	)
	userRepo := users_repo.NewPostgreRepo(pool)
	contentRepo := content_repo.NewPostgreRepo(pool)
	contentService := content_usecase.NewServiceFromRepository(contentRepo)
	progressRepo := progress_repo.NewPostgreRepo(pool)
	progressService := progress_usecase.NewService(progress_usecase.Dependencies{
		Repository: progressRepo,
		Content:    contentRepo,
	})
	gamificationRepo := gamification_repo.NewRepository(pool)
	gamificationService := gamification_usecase.NewService(gamificationRepo)
	recommendationService := recommendation_usecase.NewService(contentRepo, progressRepo)
	learningRepo := learning_repo.NewPostgreRepo(pool)
	learningService := learning_usecase.NewService(learning_usecase.Dependencies{
		Sessions:     learningRepo,
		Queue:        learningRepo,
		Attempts:     learningRepo,
		Content:      contentRepo,
		Picker:       recommendationService,
		Progress:     progress_usecase.NewLearningRecorder(progressService),
		Gamification: gamification_usecase.NewLearningRecorder(gamificationService),
	})
	mediaHandler, err := media_infra.NewHandler(media_infra.Dependencies{
		Config: cfg.S3,
		Auth:   authMiddleware,
	})
	if err != nil {
		pool.Close()
		_ = log.Sync()
		return nil, err
	}
	hasher := users_security.NewPasswordHasher()
	if err := ensureBootstrapAdmin(ctx, cfg.BootstrapAdmin, userRepo, hasher, log); err != nil {
		pool.Close()
		_ = log.Sync()
		return nil, err
	}
	tokens := users_security.NewTokenService(users_security.TokenConfig{
		AccessSecret:  cfg.Auth.AccessSecret,
		RefreshSecret: cfg.Auth.RefreshSecret,
		AccessTTL:     cfg.Auth.AccessTTL,
		RefreshTTL:    cfg.Auth.RefreshTTL,
		Issuer:        cfg.Auth.Issuer,
		Audience:      cfg.Auth.Audience,
	})

	return &App{
		userHandler: users_infra.NewHandler(users_infra.Dependencies{
			Register:   users_usecase.NewRegisterUsecase(userRepo, hasher),
			Login:      users_usecase.NewLoginUsecase(userRepo, userRepo, hasher, tokens),
			Refresh:    users_usecase.NewRefreshUsecase(userRepo, userRepo, tokens),
			GetMe:      users_usecase.NewGetMeUsecase(userRepo),
			Logout:     users_usecase.NewLogoutUsecase(userRepo, tokens),
			ChangeRole: users_usecase.NewChangeUserRoleUsecase(userRepo),
			FindUser:   users_usecase.NewFindUserUsecase(userRepo),
			Auth:       authMiddleware,
		}),
		contentHandler: content_infra.NewHandler(content_infra.Dependencies{
			Courses:    contentService,
			Sections:   contentService,
			Units:      contentService,
			Skills:     contentService,
			Challenges: contentService,
			Auth:       authMiddleware,
		}),
		learningHandler: learning_infra.NewHandler(learning_infra.Dependencies{
			Sessions: learningService,
			Auth:     authMiddleware,
		}),
		mediaHandler: mediaHandler,
		gamificationHandler: gamification_infra.NewHandler(gamification_infra.Dependencies{
			Service: gamificationService,
			Auth:    authMiddleware,
		}),
		progressHandler: progress_infra.NewHandler(progress_infra.Dependencies{
			Service: progressService,
			Auth:    authMiddleware,
		}),
		recommendationHandler: recommendation_infra.NewHandler(recommendation_infra.Dependencies{
			Service: recommendation_usecase.NewAPIService(recommendationService),
			Auth:    authMiddleware,
		}),
		cfg:    cfg,
		logger: log,
		db:     pool,
	}, nil
}

type bootstrapUserRepository interface {
	Create(ctx context.Context, params users_domain.CreateUserParams) (users_domain.User, error)
	GetByEmail(ctx context.Context, email users_domain.Email) (users_domain.User, error)
	GetByLogin(ctx context.Context, login users_domain.Login) (users_domain.User, error)
	ChangeRole(ctx context.Context, id users_domain.UserID, role users_domain.Role) (users_domain.User, error)
}

type bootstrapPasswordHasher interface {
	Hash(password string) (users_domain.PasswordHash, error)
}

func ensureBootstrapAdmin(
	ctx context.Context,
	cfg config.BootstrapAdminConfig,
	users bootstrapUserRepository,
	hasher bootstrapPasswordHasher,
	log *zap.Logger,
) error {
	if strings.TrimSpace(cfg.Email) == "" && strings.TrimSpace(cfg.Login) == "" && cfg.Password == "" {
		return nil
	}
	if strings.TrimSpace(cfg.Email) == "" || strings.TrimSpace(cfg.Login) == "" || cfg.Password == "" {
		return fmt.Errorf("BOOTSTRAP_ADMIN_EMAIL, BOOTSTRAP_ADMIN_LOGIN and BOOTSTRAP_ADMIN_PASSWORD must be set together")
	}

	email, err := users_domain.NewEmail(cfg.Email)
	if err != nil {
		return fmt.Errorf("invalid BOOTSTRAP_ADMIN_EMAIL: %w", err)
	}
	login, err := users_domain.NewLogin(cfg.Login)
	if err != nil {
		return fmt.Errorf("invalid BOOTSTRAP_ADMIN_LOGIN: %w", err)
	}
	if err := users_domain.ValidatePassword(cfg.Password); err != nil {
		return fmt.Errorf("invalid BOOTSTRAP_ADMIN_PASSWORD: %w", err)
	}

	existing, err := users.GetByEmail(ctx, *email)
	if err != nil && !errors.Is(err, users_domain.ErrUserNotFound) {
		return fmt.Errorf("find bootstrap admin by email: %w", err)
	}
	if errors.Is(err, users_domain.ErrUserNotFound) {
		existing, err = users.GetByLogin(ctx, *login)
		if err != nil && !errors.Is(err, users_domain.ErrUserNotFound) {
			return fmt.Errorf("find bootstrap admin by login: %w", err)
		}
	}

	if err == nil {
		if existing.Role != users_domain.RoleAdmin {
			if _, err := users.ChangeRole(ctx, existing.ID, users_domain.RoleAdmin); err != nil {
				return fmt.Errorf("promote bootstrap admin: %w", err)
			}
			log.Info("bootstrap admin promoted", zap.String("login", string(existing.Login)))
		}
		return nil
	}

	passwordHash, err := hasher.Hash(cfg.Password)
	if err != nil {
		return fmt.Errorf("hash bootstrap admin password: %w", err)
	}
	now := time.Now().UTC()
	_, err = users.Create(ctx, users_domain.CreateUserParams{
		ID:                users_domain.UserID(uuid.New()),
		Login:             *login,
		Email:             *email,
		Role:              users_domain.RoleAdmin,
		PasswordHash:      passwordHash,
		PasswordChangedAt: now,
		CreatedAt:         now,
		UpdatedAt:         now,
	})
	if err != nil {
		return fmt.Errorf("create bootstrap admin: %w", err)
	}
	log.Info("bootstrap admin created", zap.String("login", string(*login)))
	return nil
}

func (a *App) Port() string {
	return a.cfg.App.Port
}

func (a *App) Logger() *zap.Logger {
	return a.logger
}

func (a *App) Close() {
	if a.db != nil {
		a.db.Close()
	}
	if a.logger != nil {
		_ = a.logger.Sync()
	}
}

func (a *App) Router() http.Handler {
	r := chi.NewRouter()

	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(appmiddleware.CORS(a.cfg.CORS.AllowedOrigins))
	r.Use(appmiddleware.Logging(a.logger))
	r.Use(chimiddleware.Recoverer)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	a.userHandler.RegisterRoutes(r)
	a.contentHandler.RegisterRoutes(r)
	a.learningHandler.RegisterRoutes(r)
	a.mediaHandler.RegisterRoutes(r)
	a.gamificationHandler.RegisterRoutes(r)
	a.progressHandler.RegisterRoutes(r)
	a.recommendationHandler.RegisterRoutes(r)

	return r
}
