// Command provisioner runs tenant provisioning jobs on the VPS host.
//
// It is not a container: onboard-tenant.sh needs the host's docker, /etc/nginx
// and certbot. It listens on loopback or the docker bridge address only and
// is reached by auth-service with X-Provisioner-Token.
// See docs/provisioning-job-contract.md.
//
// Environment:
//
//	PROVISIONER_ADDR       listen address, default 127.0.0.1:8790
//	PROVISIONER_TOKEN      shared secret with auth-service (required, >= 32 chars)
//	PROVISIONER_REPO_DIR   repository checkout that holds scripts/ (required)
//	PROVISIONER_DATA_DIR   where jobs and event logs are stored (required)
//	PROVISIONER_BASH       bash binary used to run the scripts (default "bash")
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"provisioner-service/internal/provisioner"
)

const defaultAddr = "127.0.0.1:8790"

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	if err := run(log); err != nil {
		log.Error("provisioner exiting", "error", err)
		os.Exit(1)
	}
}

func run(log *slog.Logger) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}
	store, err := provisioner.NewStore(cfg.dataDir)
	if err != nil {
		return err
	}
	onboard := []string{cfg.bash, filepath.Join(cfg.repoDir, "scripts", "onboard-tenant.sh")}
	generator := []string{cfg.bash, filepath.Join(cfg.repoDir, "scripts", "generate-tenant-stack.sh")}
	mgr, err := provisioner.NewManager(provisioner.ManagerOptions{
		Store:        store,
		Runner:       &provisioner.ExecRunner{Command: onboard, Dir: cfg.repoDir},
		Logger:       log,
		RegistryPath: filepath.Join(cfg.repoDir, "backend", "tenants", "registry.tsv"),
		PortFree:     provisioner.OSPortFree,
	})
	if err != nil {
		return err
	}
	srv := &provisioner.Server{
		Manager: mgr,
		Ports:   provisioner.ExecNextPorts(generator, cfg.repoDir),
		Token:   cfg.token,
		Logger:  log,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	workerDone := make(chan struct{})
	go func() { defer close(workerDone); mgr.Run(ctx) }()

	// No WriteTimeout: the SSE stream stays open for the whole job.
	httpSrv := &http.Server{
		Addr:              cfg.addr,
		Handler:           srv.Handler(),
		ReadHeaderTimeout: 10 * time.Second,
		IdleTimeout:       120 * time.Second,
		BaseContext:       func(net.Listener) context.Context { return ctx },
	}
	errc := make(chan error, 1)
	go func() { errc <- httpSrv.ListenAndServe() }()
	log.Info("provisioner listening", "addr", cfg.addr, "repo_dir", cfg.repoDir, "data_dir", cfg.dataDir)

	select {
	case err := <-errc:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
	case <-ctx.Done():
		log.Info("shutting down")
	}
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = httpSrv.Shutdown(shutdownCtx)
	<-workerDone
	return nil
}

type config struct {
	addr, token, repoDir, dataDir, bash string
}

func loadConfig() (config, error) {
	c := config{
		addr:    envOr("PROVISIONER_ADDR", defaultAddr),
		token:   os.Getenv("PROVISIONER_TOKEN"),
		repoDir: os.Getenv("PROVISIONER_REPO_DIR"),
		dataDir: os.Getenv("PROVISIONER_DATA_DIR"),
		bash:    envOr("PROVISIONER_BASH", "bash"),
	}
	if len(c.token) < 32 {
		return c, errors.New("PROVISIONER_TOKEN must be set to at least 32 characters (e.g. openssl rand -hex 32)")
	}
	if c.repoDir == "" || c.dataDir == "" {
		return c, errors.New("PROVISIONER_REPO_DIR and PROVISIONER_DATA_DIR are required")
	}
	var err error
	if c.repoDir, err = filepath.Abs(c.repoDir); err != nil {
		return c, err
	}
	if c.dataDir, err = filepath.Abs(c.dataDir); err != nil {
		return c, err
	}
	if _, err := os.Stat(filepath.Join(c.repoDir, "scripts", "onboard-tenant.sh")); err != nil {
		return c, fmt.Errorf("PROVISIONER_REPO_DIR does not contain scripts/onboard-tenant.sh: %w", err)
	}
	if err := checkPrivateAddr(c.addr); err != nil {
		return c, err
	}
	return c, nil
}

// checkPrivateAddr refuses wildcard and public binds: the API runs scripts as
// the host user and must only be reachable from loopback or the docker bridge.
func checkPrivateAddr(addr string) error {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return fmt.Errorf("PROVISIONER_ADDR %q: %w", addr, err)
	}
	ip := net.ParseIP(host)
	if host == "localhost" {
		return nil
	}
	if ip == nil || ip.IsUnspecified() {
		return fmt.Errorf("PROVISIONER_ADDR %q must name a loopback or private IP (e.g. 127.0.0.1 or the docker bridge 172.17.0.1), not a wildcard", addr)
	}
	if !ip.IsLoopback() && !ip.IsPrivate() {
		return fmt.Errorf("PROVISIONER_ADDR %q is a public address; bind to loopback or the docker bridge", addr)
	}
	return nil
}

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
