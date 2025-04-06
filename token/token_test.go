package token_test

import (
	"testing"
	"time"

	"github.com/3elDU/rss-reader-backend/database"
	"github.com/3elDU/rss-reader-backend/token"
	"github.com/google/go-cmp/cmp"
	"github.com/jmoiron/sqlx"

	_ "modernc.org/sqlite"
)

func TestExpiration(t *testing.T) {
	tests := []struct {
		name         string
		validFor     time.Duration
		shouldExpire bool
	}{
		{"does not expire", 0, false},
		{"expires instantly", time.Nanosecond, true},
		{"expires in a long time", time.Hour, false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			t.Log(test.name)

			var v *time.Duration
			// Handle zero duration as nil
			if test.validFor != 0 {
				v = &test.validFor
			}
			tok := token.New(v)

			time.Sleep(time.Millisecond)

			if tok.Expired() != test.shouldExpire {
				t.Errorf("token expiration should be %v, got %v",
					test.shouldExpire, tok.Expired(),
				)
			}
		})
	}
}

func TestDatabase(t *testing.T) {
	godb, err := database.NewWithMigrations(":memory:", "../database/migrations")
	if err != nil {
		t.Fatal(err)
	}
	db := sqlx.NewDb(godb, "sqlite")
	repo := database.NewTokenRepository(db)

	tok := token.New(nil).ToModel()
	if err := repo.Insert(&tok); err != nil {
		t.Fatal(err)
	}

	if tok.ID != 1 {
		t.Errorf("token id should be 1, got %v", tok.ID)
	}

	tok2, err := repo.Find(tok.Token)
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(tok, *tok2); diff != "" {
		t.Errorf("two token instances should be equal (-want +got):\n%v", diff)
	}
}
