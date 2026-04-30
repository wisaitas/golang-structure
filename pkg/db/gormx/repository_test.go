package gormx

import (
	"context"
	"errors"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type testUser struct {
	ID   uint
	Name string
}

func setupTestRepo(t *testing.T) BaseRepository[testUser] {
	t.Helper()

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite failed: %v", err)
	}

	if err := db.AutoMigrate(&testUser{}); err != nil {
		t.Fatalf("migrate failed: %v", err)
	}

	return NewBaseRepository[testUser](db)
}

func TestBaseRepositoryCreateSupportsSingleAndSlice(t *testing.T) {
	repo := setupTestRepo(t)
	ctx := context.Background()

	one := &testUser{Name: "alice"}
	if err := repo.Create(ctx, one); err != nil {
		t.Fatalf("create single failed: %v", err)
	}

	many := &[]testUser{
		{Name: "bob"},
		{Name: "charlie"},
	}
	if err := repo.Create(ctx, many); err != nil {
		t.Fatalf("create slice failed: %v", err)
	}

	var users []testUser
	if err := repo.Find(ctx, &users, nil, nil, nil); err != nil {
		t.Fatalf("find failed: %v", err)
	}

	if len(users) != 3 {
		t.Fatalf("expected 3 users, got %d", len(users))
	}
}

func TestBaseRepositoryTransactionCommit(t *testing.T) {
	repo := setupTestRepo(t)
	ctx := context.Background()

	err := repo.Transaction(ctx, func(txRepo BaseRepository[testUser]) error {
		return txRepo.Create(ctx, &testUser{Name: "committed"})
	})
	if err != nil {
		t.Fatalf("transaction commit failed: %v", err)
	}

	var users []testUser
	if err := repo.Find(ctx, &users, nil, &Condition{Query: "name = ?", Args: []interface{}{"committed"}}, nil); err != nil {
		t.Fatalf("find failed: %v", err)
	}
	if len(users) != 1 {
		t.Fatalf("expected committed row, got %d rows", len(users))
	}
}

func TestBaseRepositoryTransactionRollbackOnError(t *testing.T) {
	repo := setupTestRepo(t)
	ctx := context.Background()

	err := repo.Transaction(ctx, func(txRepo BaseRepository[testUser]) error {
		if err := txRepo.Create(ctx, &testUser{Name: "rolled-back"}); err != nil {
			return err
		}
		return errors.New("force rollback")
	})
	if err == nil {
		t.Fatal("expected transaction error, got nil")
	}

	var users []testUser
	if err := repo.Find(ctx, &users, nil, &Condition{Query: "name = ?", Args: []interface{}{"rolled-back"}}, nil); err != nil {
		t.Fatalf("find failed: %v", err)
	}
	if len(users) != 0 {
		t.Fatalf("expected no rows after rollback, got %d", len(users))
	}
}
