package pushNotifications_test

import (
	"testing"
	"verni/internal/db"
	"verni/internal/repositories/pushNotifications"

	"github.com/google/uuid"
)

var (
	_s *pushNotifications.Repository
)

func getRepository(t *testing.T) pushNotifications.Repository {
	if _s != nil {
		return *_s
	}
	db, err := db.Postgres(db.PostgresConfig{
		Host:     "localhost",
		Port:     5432,
		User:     "tester",
		Password: "test_password",
		DbName:   "mydb",
	})
	if err != nil {
		t.Fatalf("failed to init repository err: %v", err)
	}
	repository := pushNotifications.PostgresRepository(db)
	if err != nil {
		t.Fatalf("failed to init repository err: %v", err)
	}
	_s = &repository
	return repository
}

func randomUid() pushNotifications.UserId {
	return pushNotifications.UserId(uuid.New().String())
}

func TestStorePushToken(t *testing.T) {
	s := getRepository(t)
	uid := randomUid()
	token := uuid.New().String()
	transaction := s.StorePushToken(uid, token)
	if err := transaction.Perform(); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	tokenFromDb, err := s.GetPushToken(uid)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if tokenFromDb == nil {
		t.Fatalf("unexpected nil")
	}
	if *tokenFromDb != token {
		t.Fatalf("should be equal")
	}
}
