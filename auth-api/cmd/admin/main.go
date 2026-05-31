// Package main is the auth-api admin CLI: operator-only tasks that connect
// directly to the same Postgres + Redis as the server. Ships in the same
// container image so it can be invoked via `docker exec`.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"github.com/trippier/auth-api/internal/apikey"
	"github.com/trippier/auth-api/internal/config"
	"github.com/trippier/auth-api/internal/db"
)

const usage = `auth-api admin CLI

Usage:
  admin set-quota --email=<email> --limit=<tokens> [--interval=<secs>]
  admin set-quota --user-id=<uuid> --limit=<tokens> [--interval=<secs>]

Examples:
  admin set-quota --email=user@example.com --limit=100000
  admin set-quota --email=user@example.com --limit=50000 --interval=2592000
`

func main() {
	if len(os.Args) < 2 {
		fmt.Fprint(os.Stderr, usage)
		os.Exit(2)
	}

	switch os.Args[1] {
	case "set-quota":
		os.Exit(runSetQuota(os.Args[2:]))
	case "-h", "--help", "help":
		fmt.Print(usage)
		os.Exit(0)
	default:
		fmt.Fprintf(os.Stderr, "unknown subcommand %q\n\n%s", os.Args[1], usage)
		os.Exit(2)
	}
}

func runSetQuota(args []string) int {
	fs := flag.NewFlagSet("set-quota", flag.ExitOnError)
	email := fs.String("email", "", "user email")
	userID := fs.String("user-id", "", "user UUID (alternative to --email)")
	limit := fs.Int("limit", 0, "new token limit (required, >0)")
	interval := fs.Int("interval", 0, "reset interval in seconds (0 = keep current)")
	if err := fs.Parse(args); err != nil {
		return 2
	}

	if *limit <= 0 {
		fmt.Fprintln(os.Stderr, "error: --limit must be > 0")
		return 2
	}
	if (*email == "") == (*userID == "") {
		fmt.Fprintln(os.Stderr, "error: exactly one of --email or --user-id required")
		return 2
	}

	svc, cleanup, err := buildService()
	if err != nil {
		fmt.Fprintf(os.Stderr, "init: %v\n", err)
		return 1
	}
	defer cleanup()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var resolvedID string
	if *userID != "" {
		resolvedID = *userID
		err = svc.SetUserQuota(ctx, resolvedID, *limit, *interval)
	} else {
		resolvedID, err = svc.SetUserQuotaByEmail(ctx, *email, *limit, *interval)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "set-quota: %v\n", err)
		return 1
	}

	fmt.Printf("ok user_id=%s tokens_limit=%d\n", resolvedID, *limit)
	return 0
}

// buildService wires the apikey.Service against the same DB + Redis the server
// uses (config is read from the same AUTH_* env vars).
func buildService() (*apikey.Service, func(), error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, nil, fmt.Errorf("config: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		return nil, nil, fmt.Errorf("postgres: %w", err)
	}

	opt, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		pool.Close()
		return nil, nil, fmt.Errorf("redis url: %w", err)
	}
	rdb := redis.NewClient(opt)

	// Silent logger — CLI output is the result line, not log spam.
	svc := apikey.New(pool, rdb, zap.NewNop())
	cleanup := func() {
		pool.Close()
		_ = rdb.Close()
	}
	return svc, cleanup, nil
}
