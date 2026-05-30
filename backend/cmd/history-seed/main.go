package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	content_repo "github.com/SIniutin/history-app-backend/internal/modules/content/repo/postgre"
	"github.com/SIniutin/history-app-backend/internal/modules/content/seeder"
	content_usecase "github.com/SIniutin/history-app-backend/internal/modules/content/usecase"
	users_repo "github.com/SIniutin/history-app-backend/internal/modules/users/repo/postgre"
	"github.com/jackc/pgx/v5/pgxpool"
)

const defaultDatabaseURL = "postgres://postgres:postgres@localhost:5432/history-db?sslmode=disable"

func main() {
	var seedPath string
	var actorEmail string
	flag.StringVar(&seedPath, "file", "seeds/history_course_structured.json", "path to structured content seed JSON")
	flag.StringVar(&actorEmail, "actor-email", "", "existing admin/content actor email")
	flag.Parse()

	if err := run(context.Background(), seedPath, resolveActorEmail(actorEmail)); err != nil {
		fmt.Fprintf(os.Stderr, "seed failed: %v\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, seedPath, actorEmail string) error {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = defaultDatabaseURL
	}
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return fmt.Errorf("create db pool: %w", err)
	}
	defer pool.Close()
	if err := pool.Ping(ctx); err != nil {
		return fmt.Errorf("ping db: %w", err)
	}

	seed, err := seeder.LoadFile(seedPath)
	if err != nil {
		return err
	}
	userRepo := users_repo.NewPostgreRepo(pool)
	actorID, err := seeder.ResolveActor(ctx, userRepo, actorEmail)
	if err != nil {
		return err
	}
	contentRepo := content_repo.NewPostgreRepo(pool)
	contentService := content_usecase.NewServiceFromRepository(contentRepo)
	result, err := seeder.NewRunner(contentService).Run(ctx, seed, actorID)
	if err != nil {
		return err
	}
	if result.Skipped {
		fmt.Printf("seed already exists for course %q, skipped\n", seed.Course.Title)
		return nil
	}
	fmt.Printf(
		"seeded content: courses=%d sections=%d units=%d skills=%d challenges=%d\n",
		result.Courses,
		result.Sections,
		result.Units,
		result.Skills,
		result.Challenges,
	)
	return nil
}

func resolveActorEmail(flagValue string) string {
	for _, value := range []string{
		flagValue,
		os.Getenv("SEED_ACTOR_EMAIL"),
		os.Getenv("BOOTSTRAP_ADMIN_EMAIL"),
		"admin@example.com",
	} {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}
