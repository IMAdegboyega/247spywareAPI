package services

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"log"
	"time"

	"blog-backend/internal/models"
	"blog-backend/internal/repository"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userRepo  *repository.UserRepository
	postRepo  *repository.PostRepository
	email     *EmailService
	jwtSecret string
}

func NewAuthService(userRepo *repository.UserRepository, postRepo *repository.PostRepository, email *EmailService, jwtSecret string) *AuthService {
	return &AuthService{
		userRepo:  userRepo,
		postRepo:  postRepo,
		email:     email,
		jwtSecret: jwtSecret,
	}
}

type Claims struct {
	UserID      uint            `json:"user_id"`
	Email       string          `json:"email"`
	Role        models.UserRole `json:"role"`
	DisplayName string          `json:"display_name"`
	jwt.RegisteredClaims
}

func (s *AuthService) Login(email, password string) (*models.LoginResponse, error) {
	user, err := s.userRepo.FindByEmail(email)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	if !user.IsActive {
		return nil, errors.New("account is deactivated")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, errors.New("invalid credentials")
	}

	token, err := s.generateToken(user)
	if err != nil {
		return nil, err
	}

	return &models.LoginResponse{
		Token: token,
		User:  *user,
	}, nil
}

func (s *AuthService) CreateUser(req models.CreateUserRequest) (*models.User, error) {
	existing, _ := s.userRepo.FindByEmail(req.Email)
	if existing != nil {
		return nil, errors.New("email already exists")
	}

	role := req.Role
	if role == "" {
		role = models.RoleAuthor
	}

	// Single-admin rule: there can only ever be one admin in the system.
	if role == models.RoleAdmin {
		count, _ := s.userRepo.CountByRole(models.RoleAdmin)
		if count > 0 {
			return nil, errors.New("an admin already exists; only one admin is allowed")
		}
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		Email:       req.Email,
		Password:    string(hashedPassword),
		DisplayName: req.DisplayName,
		Role:        role,
		IsActive:    true,
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *AuthService) UpdateUser(id uint, req models.UpdateUserRequest) (*models.User, error) {
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		return nil, errors.New("user not found")
	}

	if req.Email != "" {
		existing, _ := s.userRepo.FindByEmail(req.Email)
		if existing != nil && existing.ID != id {
			return nil, errors.New("email already exists")
		}
		user.Email = req.Email
	}

	if req.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, err
		}
		user.Password = string(hashedPassword)
	}

	displayNameChanged := false
	if req.DisplayName != "" && req.DisplayName != user.DisplayName {
		user.DisplayName = req.DisplayName
		displayNameChanged = true
	}

	if req.Role != "" && req.Role != user.Role {
		// Promotion to admin is blocked — there can only be one.
		if req.Role == models.RoleAdmin {
			return nil, errors.New("promoting users to admin is not allowed")
		}
		// Demotion of the (sole) admin would leave the system with zero admins.
		if user.Role == models.RoleAdmin {
			return nil, errors.New("the admin's role cannot be changed")
		}
		user.Role = req.Role
	}

	if req.IsActive != nil {
		// Don't allow the admin to be deactivated — they'd lock themselves out.
		if user.Role == models.RoleAdmin && !*req.IsActive {
			return nil, errors.New("the admin account cannot be deactivated")
		}
		user.IsActive = *req.IsActive
	}

	if err := s.userRepo.Update(user); err != nil {
		return nil, err
	}

	// Keep the embedded author summary on every post in sync.
	if displayNameChanged && s.postRepo != nil {
		_ = s.postRepo.PropagateAuthorRename(user.ID, user.DisplayName)
	}

	return user, nil
}

func (s *AuthService) GetAllUsers() ([]models.User, error) {
	return s.userRepo.FindAll()
}

func (s *AuthService) GetUserByID(id uint) (*models.User, error) {
	return s.userRepo.FindByID(id)
}

func (s *AuthService) DeleteUser(id uint) error {
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		return errors.New("user not found")
	}
	// The admin account cannot be deleted — that would leave the system
	// with zero admins and lock everyone out.
	if user.Role == models.RoleAdmin {
		return errors.New("the admin account cannot be deleted")
	}
	return s.userRepo.Delete(id)
}

func (s *AuthService) ToggleUserActive(id uint) (*models.User, error) {
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		return nil, errors.New("user not found")
	}
	if user.Role == models.RoleAdmin {
		return nil, errors.New("the admin account cannot be deactivated")
	}

	user.IsActive = !user.IsActive

	if err := s.userRepo.Update(user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *AuthService) generateToken(user *models.User) (string, error) {
	claims := &Claims{
		UserID:      user.ID,
		Email:       user.Email,
		Role:        user.Role,
		DisplayName: user.DisplayName,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}

func (s *AuthService) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

func (s *AuthService) CountAuthors() (int64, error) {
	return s.userRepo.CountByRole(models.RoleAuthor)
}

// ===========================================================================
// Forgot / Reset password
// ===========================================================================

const passwordResetTokenTTL = 1 * time.Hour

// RequestPasswordReset is the entry point for the "forgot password" flow.
// It always returns nil — we don't leak whether the email exists. If the
// email matches an active user, we generate a reset token, persist it, and
// fire off the email + the admin FYI in the background.
func (s *AuthService) RequestPasswordReset(email string) error {
	user, err := s.userRepo.FindByEmail(email)
	if err != nil || user == nil || !user.IsActive {
		// Don't surface the lookup result. The handler will still return a
		// success response so this endpoint cannot be used to enumerate accounts.
		return nil
	}

	token, err := generateResetToken()
	if err != nil {
		log.Printf("forgot-password: failed to generate token: %v", err)
		return nil
	}

	expiry := time.Now().UTC().Add(passwordResetTokenTTL)
	user.ResetToken = token
	user.ResetTokenExpiresAt = &expiry

	if err := s.userRepo.Update(user); err != nil {
		log.Printf("forgot-password: failed to persist token: %v", err)
		return nil
	}

	// Fire the emails off the request goroutine — slow SMTP shouldn't make the
	// HTTP response wait.
	go func(target models.User, tok string) {
		if s.email != nil && s.email.IsNotificationEnabled() {
			if err := s.email.SendPasswordReset(target.Email, target.DisplayName, tok); err != nil {
				log.Printf("forgot-password: send to user failed: %v", err)
			}

			// FYI to the admin (skip if the requester IS the admin).
			if admin, err := s.userRepo.FindAdmin(); err == nil && admin.Email != target.Email {
				headline := target.DisplayName + " requested a password reset"
				body := "Just so you know — " + target.DisplayName + " (" + target.Email + ") just used the forgot-password flow. They'll get an email with a reset link."
				if err := s.email.SendAdminAlert(admin.Email, "Password reset requested", headline, body); err != nil {
					log.Printf("forgot-password: admin alert failed: %v", err)
				}
			}
		}
	}(*user, token)

	return nil
}

// ResetPassword validates the token + email pair and, if valid, sets the new
// password and clears the token.
func (s *AuthService) ResetPassword(token, email, newPassword string) error {
	user, err := s.userRepo.FindByEmail(email)
	if err != nil || user == nil {
		return errors.New("invalid or expired reset link")
	}

	if user.ResetToken == "" || user.ResetToken != token {
		return errors.New("invalid or expired reset link")
	}
	if user.ResetTokenExpiresAt == nil || time.Now().UTC().After(*user.ResetTokenExpiresAt) {
		return errors.New("invalid or expired reset link")
	}
	if !user.IsActive {
		return errors.New("account is deactivated")
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.Password = string(hashed)
	user.ResetToken = ""
	user.ResetTokenExpiresAt = nil

	if err := s.userRepo.Update(user); err != nil {
		return err
	}
	return nil
}

func generateResetToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
