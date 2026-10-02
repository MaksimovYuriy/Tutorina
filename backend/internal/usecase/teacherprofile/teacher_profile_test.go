package teacherprofile

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"strings"
	"testing"

	"github.com/maksimovyuriy/tutorina/backend/internal/entity"
)

func TestReplacePhotoDeletesPreviousPhoto(t *testing.T) {
	repository := &repositoryStub{
		replaced: entity.TeacherProfile{ID: 7, PhotoURL: "/api/v1/media/teacher-photos/new.png"},
		oldURL:   "/api/v1/media/teacher-photos/old.png",
	}
	photos := &photoStorageStub{savedURL: repository.replaced.PhotoURL}
	service := New(repository, photos, slog.Default())

	profile, err := service.ReplacePhoto(context.Background(), 7, strings.NewReader("image"))
	if err != nil {
		t.Fatal(err)
	}
	if profile.PhotoURL != repository.replaced.PhotoURL {
		t.Fatalf("photo URL = %q", profile.PhotoURL)
	}
	if len(photos.deleted) != 1 || photos.deleted[0] != repository.oldURL {
		t.Fatalf("deleted photos = %#v", photos.deleted)
	}
}

func TestReplacePhotoCleansNewPhotoWhenDatabaseUpdateFails(t *testing.T) {
	databaseError := errors.New("database unavailable")
	repository := &repositoryStub{replaceError: databaseError}
	photos := &photoStorageStub{savedURL: "/api/v1/media/teacher-photos/new.png"}
	service := New(repository, photos, slog.Default())

	_, err := service.ReplacePhoto(context.Background(), 7, strings.NewReader("image"))
	if !errors.Is(err, databaseError) {
		t.Fatalf("error = %v", err)
	}
	if len(photos.deleted) != 1 || photos.deleted[0] != photos.savedURL {
		t.Fatalf("deleted photos = %#v", photos.deleted)
	}
}

type repositoryStub struct {
	replaced     entity.TeacherProfile
	oldURL       string
	replaceError error
}

func (stub *repositoryStub) List(context.Context, bool) ([]entity.TeacherProfile, error) {
	return nil, nil
}
func (stub *repositoryStub) GetByUserID(context.Context, int64) (entity.TeacherProfile, error) {
	return entity.TeacherProfile{}, nil
}

func (stub *repositoryStub) Get(context.Context, int64) (entity.TeacherProfile, error) {
	return entity.TeacherProfile{}, nil
}
func (stub *repositoryStub) Create(context.Context, entity.TeacherProfile) (entity.TeacherProfile, error) {
	return entity.TeacherProfile{}, nil
}
func (stub *repositoryStub) Update(context.Context, entity.TeacherProfile) (entity.TeacherProfile, error) {
	return entity.TeacherProfile{}, nil
}
func (stub *repositoryStub) ReplacePhoto(context.Context, int64, string) (entity.TeacherProfile, string, error) {
	return stub.replaced, stub.oldURL, stub.replaceError
}
func (stub *repositoryStub) CreateAccount(context.Context, int64, string, string) (entity.TeacherProfile, error) {
	return entity.TeacherProfile{}, nil
}
func (stub *repositoryStub) ResetPassword(context.Context, int64, string) error { return nil }

func (stub *repositoryStub) Archive(context.Context, int64) error { return nil }

type photoStorageStub struct {
	savedURL string
	saveErr  error
	deleted  []string
}

func (stub *photoStorageStub) Save(io.Reader) (string, error) {
	return stub.savedURL, stub.saveErr
}
func (stub *photoStorageStub) Delete(url string) error {
	stub.deleted = append(stub.deleted, url)
	return nil
}
