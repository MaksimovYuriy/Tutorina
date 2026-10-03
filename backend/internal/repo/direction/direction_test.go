package direction

import (
	"errors"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/maksimovyuriy/tutorina/backend/internal/repo"
	"testing"
)

func TestReferencedDirectionAndDuplicateNameAreConflicts(t *testing.T) {
	for _, code := range []string{"23001", "23503", "23505"} {
		if err := mapError(&pgconn.PgError{Code: code}); !errors.Is(err, repo.ErrConflict) {
			t.Fatalf("SQLSTATE %s: %v", code, err)
		}
	}
}
