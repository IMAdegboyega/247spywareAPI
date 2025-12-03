package services

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

type UploadService struct {
	uploadPath string
}

func NewUploadService(uploadPath string) *UploadService {
	if uploadPath == "" {
		uploadPath = "./uploads"
	}

	// Create upload directory if it doesn't exist
	os.MkdirAll(uploadPath, os.ModePerm)

	return &UploadService{uploadPath: uploadPath}
}

func (s *UploadService) UploadImage(file *multipart.FileHeader) (string, error) {
	// Validate file type
	if !s.isValidImageType(file.Filename) {
		return "", errors.New("invalid file type. Allowed types: jpg, jpeg, png, gif, webp")
	}

	// Validate file size (max 10MB)
	if file.Size > 10*1024*1024 {
		return "", errors.New("file too large. Maximum size is 10MB")
	}

	// Open the uploaded file
	src, err := file.Open()
	if err != nil {
		return "", err
	}
	defer src.Close()

	// Generate unique filename
	ext := filepath.Ext(file.Filename)
	filename := fmt.Sprintf("%s-%s%s", time.Now().Format("20060102"), uuid.New().String()[:8], ext)

	// Create the destination file
	dstPath := filepath.Join(s.uploadPath, filename)
	dst, err := os.Create(dstPath)
	if err != nil {
		return "", err
	}
	defer dst.Close()

	// Copy the file
	if _, err := io.Copy(dst, src); err != nil {
		return "", err
	}

	// Return the relative path for storage in database
	return "/uploads/" + filename, nil
}

func (s *UploadService) DeleteImage(imagePath string) error {
	if imagePath == "" {
		return nil
	}

	// Remove the /uploads/ prefix to get the filename
	filename := strings.TrimPrefix(imagePath, "/uploads/")
	fullPath := filepath.Join(s.uploadPath, filename)

	// Check if file exists
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return nil
	}

	return os.Remove(fullPath)
}

func (s *UploadService) isValidImageType(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	validExtensions := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
		".webp": true,
	}
	return validExtensions[ext]
}