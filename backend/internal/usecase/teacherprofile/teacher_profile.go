package teacherprofile

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/mail"
	"strings"

	"github.com/maksimovyuriy/tutorina/backend/internal/entity"
	"github.com/maksimovyuriy/tutorina/backend/internal/repo"
	"github.com/maksimovyuriy/tutorina/backend/internal/storage/localphotos"
	"github.com/maksimovyuriy/tutorina/backend/internal/usecase"
	"golang.org/x/crypto/bcrypt"
)

type Repository interface {
	List(context.Context, bool) ([]entity.TeacherProfile, error)
	Get(context.Context, int64) (entity.TeacherProfile, error)
	GetByUserID(context.Context, int64) (entity.TeacherProfile, error)
	Create(context.Context, entity.TeacherProfile) (entity.TeacherProfile, error)
	Update(context.Context, entity.TeacherProfile) (entity.TeacherProfile, error)
	ReplacePhoto(context.Context, int64, string) (entity.TeacherProfile, string, error)
	ResetPassword(context.Context, int64, string) error
	CreateAccount(context.Context, int64, string, string) (entity.TeacherProfile, error)
	Archive(context.Context, int64) error
}

type PhotoStorage interface {
	Save(io.Reader) (string, error)
	Delete(string) error
}

type Service struct {
	repository Repository
	photos     PhotoStorage
	log        *slog.Logger
}

func New(repository Repository, photos PhotoStorage, log *slog.Logger) *Service {
	return &Service{repository: repository, photos: photos, log: log}
}

func (s *Service) ListPublic(ctx context.Context) ([]entity.TeacherProfile, error) {
	return s.repository.List(ctx, true)
}

func (s *Service) ListAll(ctx context.Context) ([]entity.TeacherProfile, error) {
	return s.repository.List(ctx, false)
}

func (s *Service) Create(ctx context.Context, profile entity.TeacherProfile) (entity.TeacherProfile, error) {
	profile.ID = 0
	profile.UserID = nil
	profile.PhotoURL = ""
	if err := validate(&profile); err != nil {
		return entity.TeacherProfile{}, err
	}
	return s.repository.Create(ctx, profile)
}

func (s *Service) Update(ctx context.Context, profile entity.TeacherProfile) (entity.TeacherProfile, error) {
	if profile.ID <= 0 {
		return entity.TeacherProfile{}, fmt.Errorf("%w: invalid id", usecase.ErrInvalidInput)
	}
	current, err := s.repository.Get(ctx, profile.ID)
	if err != nil {
		return entity.TeacherProfile{}, err
	}
	profile.UserID = current.UserID
	profile.PhotoURL = current.PhotoURL
	if err := validate(&profile); err != nil {
		return entity.TeacherProfile{}, err
	}
	return s.repository.Update(ctx, profile)
}

func (s *Service) GetMine(ctx context.Context, userID int64) (entity.TeacherProfile, error) {
	if userID <= 0 {
		return entity.TeacherProfile{}, usecase.ErrInvalidInput
	}
	return s.repository.GetByUserID(ctx, userID)
}

func (s *Service) UpdateMine(ctx context.Context, userID int64, profile entity.TeacherProfile) (entity.TeacherProfile, error) {
	current, err := s.GetMine(ctx, userID)
	if err != nil {
		return entity.TeacherProfile{}, err
	}
	profile.ID = current.ID
	profile.UserID = current.UserID
	profile.PhotoURL = current.PhotoURL
	if err := validate(&profile); err != nil {
		return entity.TeacherProfile{}, err
	}
	return s.repository.Update(ctx, profile)
}

func (s *Service) ReplaceMinePhoto(ctx context.Context, userID int64, photo io.Reader) (entity.TeacherProfile, error) {
	profile, err := s.GetMine(ctx, userID)
	if err != nil {
		return entity.TeacherProfile{}, err
	}
	return s.ReplacePhoto(ctx, profile.ID, photo)
}

func (s *Service) RemoveMinePhoto(ctx context.Context, userID int64) (entity.TeacherProfile, error) {
	profile, err := s.GetMine(ctx, userID)
	if err != nil {
		return entity.TeacherProfile{}, err
	}
	return s.RemovePhoto(ctx, profile.ID)
}

func (s *Service) ResetPassword(ctx context.Context, profileID int64, password string) error {
	if profileID <= 0 || len(password) < 12 {
		return fmt.Errorf("%w: password must contain at least 12 characters", usecase.ErrInvalidInput)
	}
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return s.repository.ResetPassword(ctx, profileID, string(passwordHash))
}
func (s *Service) ReplacePhoto(ctx context.Context, profileID int64, photo io.Reader) (entity.TeacherProfile, error) {
	if profileID <= 0 || photo == nil {
		return entity.TeacherProfile{}, fmt.Errorf("%w: invalid photo", usecase.ErrInvalidInput)
	}
	photoURL, err := s.photos.Save(photo)
	if err != nil {
		if errors.Is(err, localphotos.ErrEmpty) || errors.Is(err, localphotos.ErrTooLarge) || errors.Is(err, localphotos.ErrUnsupported) {
			return entity.TeacherProfile{}, fmt.Errorf("%w: %v", usecase.ErrInvalidInput, err)
		}
		return entity.TeacherProfile{}, err
	}
	profile, oldPhotoURL, err := s.repository.ReplacePhoto(ctx, profileID, photoURL)
	if err != nil {
		if cleanupErr := s.photos.Delete(photoURL); cleanupErr != nil {
			s.log.Warn("Failed to clean up rejected teacher photo", slog.Any("error", cleanupErr))
		}
		return entity.TeacherProfile{}, err
	}
	s.deletePhotoBestEffort(oldPhotoURL, profileID)
	return profile, nil
}

func (s *Service) RemovePhoto(ctx context.Context, profileID int64) (entity.TeacherProfile, error) {
	if profileID <= 0 {
		return entity.TeacherProfile{}, fmt.Errorf("%w: invalid id", usecase.ErrInvalidInput)
	}
	profile, oldPhotoURL, err := s.repository.ReplacePhoto(ctx, profileID, "")
	if err != nil {
		return entity.TeacherProfile{}, err
	}
	s.deletePhotoBestEffort(oldPhotoURL, profileID)
	return profile, nil
}

func (s *Service) Archive(ctx context.Context, profileID int64) error {
	if profileID <= 0 {
		return fmt.Errorf("%w: invalid id", usecase.ErrInvalidInput)
	}
	profile, err := s.repository.Get(ctx, profileID)
	if err != nil {
		return err
	}
	if err := s.repository.Archive(ctx, profileID); err != nil {
		return err
	}
	s.deletePhotoBestEffort(profile.PhotoURL, profileID)
	return nil
}

func (s *Service) deletePhotoBestEffort(photoURL string, profileID int64) {
	if err := s.photos.Delete(photoURL); err != nil {
		s.log.Warn("Failed to delete teacher photo", slog.Int64("teacher_profile_id", profileID), slog.Any("error", err))
	}
}

func (s *Service) CreateAccount(ctx context.Context, profileID int64, email, password string) (entity.TeacherProfile, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	address, err := mail.ParseAddress(email)
	if profileID <= 0 || err != nil || address.Address != email {
		return entity.TeacherProfile{}, fmt.Errorf("%w: invalid account data", usecase.ErrInvalidInput)
	}
	if len(password) < 12 {
		return entity.TeacherProfile{}, fmt.Errorf("%w: password must contain at least 12 characters", usecase.ErrInvalidInput)
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return entity.TeacherProfile{}, err
	}
	profile, err := s.repository.CreateAccount(ctx, profileID, email, string(hash))
	if errors.Is(err, repo.ErrConflict) {
		return entity.TeacherProfile{}, usecase.ErrConflict
	}
	return profile, err
}

func validate(profile *entity.TeacherProfile) error {
	profile.DisplayName = strings.TrimSpace(profile.DisplayName)
	profile.Education = strings.TrimSpace(profile.Education)
	profile.Experience = strings.TrimSpace(profile.Experience)
	profile.Approach = strings.TrimSpace(profile.Approach)
	if profile.DisplayName == "" {
		return fmt.Errorf("%w: display name is required", usecase.ErrInvalidInput)
	}
	if len([]rune(profile.DisplayName)) > 160 {
		return fmt.Errorf("%w: display name is too long", usecase.ErrInvalidInput)
	}
	return nil
}
