package controllers

import (
	"context"
	"fmt"
	"os"
	"sync"
	"testing"

	"auth-service/models"
	"auth-service/pkg/logger"
	"auth-service/repositories"
	userService "auth-service/services/user"

	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// TestSeatConcurrentCreatesPostgres runs the concurrent-create race against a
// real Postgres with a multi-connection pool, where only the
// pg_advisory_xact_lock in WithSeatLock keeps two creates from both taking the
// last seat. Opt-in: set AUTH_SEAT_TEST_PG_DSN to an empty, throwaway
// database (the test creates and drops its tables). Never point it at a real
// tenant database.
func TestSeatConcurrentCreatesPostgres(t *testing.T) {
	dsn := os.Getenv("AUTH_SEAT_TEST_PG_DSN")
	if dsn == "" {
		t.Skip("AUTH_SEAT_TEST_PG_DSN not set")
	}
	logger.Log = zap.NewNop()
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: gormlogger.Discard})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(16)
	t.Cleanup(func() { _ = sqlDB.Close() })

	drop := func() {
		_ = db.Migrator().DropTable("user_roles", "role_permissions", &models.TrustedDevice{}, &models.MFASetup{},
			&models.RefreshToken{}, &models.ConfidentialityAgreement{}, &models.SysAlertLog{}, &models.User{}, &models.Role{}, &models.Permission{})
	}
	drop()
	t.Cleanup(drop)
	if err := db.AutoMigrate(&models.Permission{}, &models.Role{}, &models.User{}, &models.RefreshToken{},
		&models.MFASetup{}, &models.TrustedDevice{}, &models.ConfidentialityAgreement{}, &models.SysAlertLog{}); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"ADMIN", "AUDITOR", "VIEWER"} {
		if err := db.Create(&models.Role{Name: name}).Error; err != nil {
			t.Fatal(err)
		}
	}

	repo := repositories.NewUserRepository(db)
	for i := 0; i < 49; i++ {
		u := &models.User{Username: fmt.Sprintf("seed%d", i), Email: fmt.Sprintf("seed%d@x.id", i), PasswordHash: "x", IsActive: true}
		if err := db.Create(u).Error; err != nil {
			t.Fatal(err)
		}
		if err := repo.AssignRoles(u.ID, []string{"AUDITOR"}); err != nil {
			t.Fatal(err)
		}
	}

	svc := userService.NewUserService(repo, nil)
	const workers = 16
	var wg sync.WaitGroup
	start := make(chan struct{})
	results := make([]*models.CreateUserResponse, workers)
	errs := make([]error, workers)
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			<-start
			results[i], errs[i] = svc.CreateUser(context.Background(), &models.CreateUserRequest{
				Username: fmt.Sprintf("racer%d", i), Email: fmt.Sprintf("racer%d@x.id", i),
				Password: "secret1", Roles: []string{"AUDITOR"},
			})
		}(i)
	}
	close(start)
	wg.Wait()

	full := 0
	for i := range results {
		if errs[i] != nil {
			t.Fatalf("create %d: %v", i, errs[i])
		}
		if !results[i].SeatDowngraded {
			full++
		}
	}
	used, err := repo.CountSeats()
	if err != nil {
		t.Fatal(err)
	}
	if full != 1 || used != 50 {
		t.Fatalf("full creates = %d, used = %d; want 1 and 50", full, used)
	}
}
