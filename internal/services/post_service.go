package services

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"blog-backend/internal/models"
	"blog-backend/internal/repository"
)

type PostService struct {
	postRepo       *repository.PostRepository
	subscriberRepo *repository.SubscriberRepository
}

func NewPostService(postRepo *repository.PostRepository, subscriberRepo *repository.SubscriberRepository) *PostService {
	return &PostService{
		postRepo:       postRepo,
		subscriberRepo: subscriberRepo,
	}
}

func (s *PostService) Create(req models.CreatePostRequest, authorID uint) (*models.Post, error) {
	slug := s.generateUniqueSlug(req.Title)

	status := req.Status
	if status == "" {
		status = models.StatusDraft
	}

	post := &models.Post{
		Title:      req.Title,
		Slug:       slug,
		Content:    req.Content,
		Excerpt:    req.Excerpt,
		CategoryID: req.CategoryID,
		AuthorID:   authorID,
		Status:     status,
	}

	// Handle scheduled posts
	if status == models.StatusScheduled && req.ScheduledFor != nil {
		post.ScheduledFor = req.ScheduledFor
	}

	// If publishing immediately
	if status == models.StatusPublished {
		now := time.Now()
		post.PublishedAt = &now
		post.IsLatestNews = true
		post.LatestNewsAt = &now
	}

	if err := s.postRepo.Create(post); err != nil {
		return nil, err
	}

	// If published, notify subscribers
	if status == models.StatusPublished {
		go s.notifySubscribers(post)
	}

	return s.postRepo.FindByID(post.ID)
}

func (s *PostService) GetByID(id uint) (*models.Post, error) {
	return s.postRepo.FindByID(id)
}

func (s *PostService) GetBySlug(slug string) (*models.Post, error) {
	post, err := s.postRepo.FindBySlug(slug)
	if err != nil {
		return nil, err
	}

	// Increment view count
	if post.Status == models.StatusPublished {
		s.postRepo.IncrementViewCount(post.ID)
		post.ViewCount++
	}

	return post, nil
}

func (s *PostService) GetAllPosts(page, perPage int) (*models.PaginatedResponse, error) {
	posts, total, err := s.postRepo.FindAll(page, perPage)
	if err != nil {
		return nil, err
	}

	totalPages := int(total) / perPage
	if int(total)%perPage > 0 {
		totalPages++
	}

	return &models.PaginatedResponse{
		Data:       posts,
		Total:      total,
		Page:       page,
		PerPage:    perPage,
		TotalPages: totalPages,
	}, nil
}

func (s *PostService) GetPublishedPosts(page, perPage int) (*models.PaginatedResponse, error) {
	posts, total, err := s.postRepo.FindPublished(page, perPage)
	if err != nil {
		return nil, err
	}

	totalPages := int(total) / perPage
	if int(total)%perPage > 0 {
		totalPages++
	}

	return &models.PaginatedResponse{
		Data:       posts,
		Total:      total,
		Page:       page,
		PerPage:    perPage,
		TotalPages: totalPages,
	}, nil
}

func (s *PostService) GetLatestNews(limit int) ([]models.Post, error) {
	if limit <= 0 {
		limit = 10
	}
	return s.postRepo.FindLatestNews(limit)
}

func (s *PostService) GetPostsByCategory(categoryID uint, page, perPage int) (*models.PaginatedResponse, error) {
	posts, total, err := s.postRepo.FindByCategory(categoryID, page, perPage)
	if err != nil {
		return nil, err
	}

	totalPages := int(total) / perPage
	if int(total)%perPage > 0 {
		totalPages++
	}

	return &models.PaginatedResponse{
		Data:       posts,
		Total:      total,
		Page:       page,
		PerPage:    perPage,
		TotalPages: totalPages,
	}, nil
}

func (s *PostService) Update(id uint, req models.UpdatePostRequest, userID uint, isAdmin bool) (*models.Post, error) {
	post, err := s.postRepo.FindByID(id)
	if err != nil {
		return nil, errors.New("post not found")
	}

	// Check permission (authors can only edit their own posts)
	if !isAdmin && post.AuthorID != userID {
		return nil, errors.New("unauthorized to edit this post")
	}

	wasPublished := post.Status == models.StatusPublished

	if req.Title != "" {
		post.Title = req.Title
		post.Slug = s.generateUniqueSlugExcluding(req.Title, id)
	}
	if req.Content != "" {
		post.Content = req.Content
	}
	if req.Excerpt != "" {
		post.Excerpt = req.Excerpt
	}
	if req.CategoryID != nil {
		post.CategoryID = req.CategoryID
	}
	if req.BannerImage != "" {
		post.BannerImage = req.BannerImage
	}
	if req.Status != "" {
		post.Status = req.Status

		// Handle status changes
		if req.Status == models.StatusPublished && !wasPublished {
			now := time.Now()
			post.PublishedAt = &now
			post.IsLatestNews = true
			post.LatestNewsAt = &now
			go s.notifySubscribers(post)
		}

		if req.Status == models.StatusScheduled && req.ScheduledFor != nil {
			post.ScheduledFor = req.ScheduledFor
		}
	}

	if err := s.postRepo.Update(post); err != nil {
		return nil, err
	}

	return s.postRepo.FindByID(id)
}

func (s *PostService) UpdateStatus(id uint, status models.PostStatus, userID uint, isAdmin bool) (*models.Post, error) {
	post, err := s.postRepo.FindByID(id)
	if err != nil {
		return nil, errors.New("post not found")
	}

	// Check permission
	if !isAdmin && post.AuthorID != userID {
		return nil, errors.New("unauthorized to update this post")
	}

	wasPublished := post.Status == models.StatusPublished
	post.Status = status

	if status == models.StatusPublished && !wasPublished {
		now := time.Now()
		post.PublishedAt = &now
		post.IsLatestNews = true
		post.LatestNewsAt = &now
		go s.notifySubscribers(post)
	}

	if err := s.postRepo.Update(post); err != nil {
		return nil, err
	}

	return post, nil
}

func (s *PostService) AutoSave(id uint, content string, title string, userID uint, isAdmin bool) (*models.Post, error) {
	post, err := s.postRepo.FindByID(id)
	if err != nil {
		return nil, errors.New("post not found")
	}

	// Check permission
	if !isAdmin && post.AuthorID != userID {
		return nil, errors.New("unauthorized to edit this post")
	}

	// Only auto-save drafts
	if post.Status != models.StatusDraft {
		return post, nil
	}

	if content != "" {
		post.Content = content
	}
	if title != "" {
		post.Title = title
	}

	if err := s.postRepo.Update(post); err != nil {
		return nil, err
	}

	return post, nil
}

func (s *PostService) Delete(id uint) error {
	return s.postRepo.Delete(id)
}

func (s *PostService) PublishScheduledPosts() error {
	posts, err := s.postRepo.FindScheduledDue()
	if err != nil {
		return err
	}

	for _, post := range posts {
		now := time.Now()
		post.Status = models.StatusPublished
		post.PublishedAt = &now
		post.IsLatestNews = true
		post.LatestNewsAt = &now

		if err := s.postRepo.Update(&post); err != nil {
			continue
		}

		go s.notifySubscribers(&post)
	}

	return nil
}

func (s *PostService) ExpireLatestNews(duration time.Duration) error {
	posts, err := s.postRepo.FindExpiredLatestNews(duration)
	if err != nil {
		return err
	}

	for _, post := range posts {
		post.IsLatestNews = false
		if err := s.postRepo.Update(&post); err != nil {
			continue
		}
	}

	return nil
}

func (s *PostService) GetStats() (*models.DashboardStats, error) {
	totalPosts, _ := s.postRepo.CountAll()
	publishedPosts, _ := s.postRepo.CountByStatus(models.StatusPublished)
	draftPosts, _ := s.postRepo.CountByStatus(models.StatusDraft)
	scheduledPosts, _ := s.postRepo.CountByStatus(models.StatusScheduled)
	totalViews, _ := s.postRepo.SumViewCount()
	totalSubscribers, _ := s.subscriberRepo.CountActive()

	return &models.DashboardStats{
		TotalPosts:       totalPosts,
		PublishedPosts:   publishedPosts,
		DraftPosts:       draftPosts,
		ScheduledPosts:   scheduledPosts,
		TotalViews:       totalViews,
		TotalSubscribers: totalSubscribers,
	}, nil
}

func (s *PostService) Search(query string, page, perPage int) (*models.PaginatedResponse, error) {
	posts, total, err := s.postRepo.Search(query, page, perPage)
	if err != nil {
		return nil, err
	}

	totalPages := int(total) / perPage
	if int(total)%perPage > 0 {
		totalPages++
	}

	return &models.PaginatedResponse{
		Data:       posts,
		Total:      total,
		Page:       page,
		PerPage:    perPage,
		TotalPages: totalPages,
	}, nil
}

func (s *PostService) generateUniqueSlug(title string) string {
	base := generatePostSlug(title)
	slug := base
	counter := 1

	for {
		_, err := s.postRepo.FindBySlug(slug)
		if err != nil {
			break
		}
		slug = fmt.Sprintf("%s-%d", base, counter)
		counter++
	}

	return slug
}

func (s *PostService) generateUniqueSlugExcluding(title string, excludeID uint) string {
	base := generatePostSlug(title)
	slug := base
	counter := 1

	for {
		post, err := s.postRepo.FindBySlug(slug)
		if err != nil || post.ID == excludeID {
			break
		}
		slug = fmt.Sprintf("%s-%d", base, counter)
		counter++
	}

	return slug
}

func generatePostSlug(s string) string {
	slug := strings.ToLower(s)
	slug = strings.ReplaceAll(slug, " ", "-")
	reg := regexp.MustCompile("[^a-z0-9-]+")
	slug = reg.ReplaceAllString(slug, "")
	reg = regexp.MustCompile("-+")
	slug = reg.ReplaceAllString(slug, "-")
	slug = strings.Trim(slug, "-")
	return slug
}

func (s *PostService) notifySubscribers(post *models.Post) {
	subscribers, err := s.subscriberRepo.FindAllActive()
	if err != nil {
		return
	}

	// This is where you'd integrate with an email service
	// For now, just log the notification
	for _, sub := range subscribers {
		// TODO: Send email to sub.Email about new post
		_ = sub // placeholder
	}
}