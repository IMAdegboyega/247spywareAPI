package services

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"blog-backend/internal/models"
	"blog-backend/internal/repository"
)

type PostService struct {
	postRepo       *repository.PostRepository
	subscriberRepo *repository.SubscriberRepository
	userRepo       *repository.UserRepository
	categoryRepo   *repository.CategoryRepository
	emailService   *EmailService
}

func NewPostService(
	postRepo *repository.PostRepository,
	subscriberRepo *repository.SubscriberRepository,
	userRepo *repository.UserRepository,
	categoryRepo *repository.CategoryRepository,
	email *EmailService,
) *PostService {
	return &PostService{
		postRepo:       postRepo,
		subscriberRepo: subscriberRepo,
		userRepo:       userRepo,
		categoryRepo:   categoryRepo,
		emailService:   email,
	}
}

// hydrateEmbedded fills post.Author + post.Category from the related collections
// before the post is written to the database. Called from Create and from
// Update whenever the relevant ids changed.
func (s *PostService) hydrateEmbedded(post *models.Post) {
	if post.AuthorID != 0 {
		if u, err := s.userRepo.FindByID(post.AuthorID); err == nil {
			post.Author = &models.EmbeddedAuthor{
				ID:          u.ID,
				DisplayName: u.DisplayName,
			}
		}
	}
	if post.CategoryID != nil && *post.CategoryID != 0 {
		if c, err := s.categoryRepo.FindByID(*post.CategoryID); err == nil {
			post.Category = &models.EmbeddedCategory{
				ID:   c.ID,
				Name: c.Name,
				Slug: c.Slug,
			}
		}
	} else {
		post.Category = nil
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

	if status == models.StatusScheduled && req.ScheduledFor != nil {
		post.ScheduledFor = req.ScheduledFor
	}

	if status == models.StatusPublished {
		now := time.Now()
		post.PublishedAt = &now
		post.IsLatestNews = true
		post.LatestNewsAt = &now
	}

	s.hydrateEmbedded(post)

	if err := s.postRepo.Create(post); err != nil {
		return nil, err
	}

	if status == models.StatusPublished {
		go s.notifySubscribers(post)
	}

	// Tell the admin a new post just landed (skip if the admin created it
	// themselves — they already know).
	go s.notifyAdminNewPost(post)

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

	if post.Status == models.StatusPublished {
		_ = s.postRepo.IncrementViewCount(post.ID)
		post.ViewCount++
	}

	return post, nil
}

func (s *PostService) GetAllPosts(page, perPage int) (*models.PaginatedResponse, error) {
	posts, total, err := s.postRepo.FindAll(page, perPage)
	if err != nil {
		return nil, err
	}
	return paginated(posts, total, page, perPage), nil
}

func (s *PostService) GetPublishedPosts(page, perPage int) (*models.PaginatedResponse, error) {
	posts, total, err := s.postRepo.FindPublished(page, perPage)
	if err != nil {
		return nil, err
	}
	return paginated(posts, total, page, perPage), nil
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
	return paginated(posts, total, page, perPage), nil
}

func (s *PostService) Update(id uint, req models.UpdatePostRequest, userID uint, isAdmin bool) (*models.Post, error) {
	post, err := s.postRepo.FindByID(id)
	if err != nil {
		return nil, errors.New("post not found")
	}

	if !isAdmin && post.AuthorID != userID {
		return nil, errors.New("unauthorized to edit this post")
	}

	prevStatus := post.Status
	wasPublished := post.Status == models.StatusPublished
	categoryChanged := false

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
		if post.CategoryID == nil || *post.CategoryID != *req.CategoryID {
			categoryChanged = true
		}
		post.CategoryID = req.CategoryID
	}
	if req.BannerImage != "" {
		post.BannerImage = req.BannerImage
	}
	if req.Status != "" {
		post.Status = req.Status

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

		// "Post taken down" — only fire when transitioning from published
		// (the only state where a post is visible publicly) to offline.
		if req.Status == models.StatusOffline && prevStatus == models.StatusPublished {
			go s.notifyPostTakenDown(post)
		}
	}

	if categoryChanged {
		s.hydrateEmbedded(post)
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

	if !isAdmin && post.AuthorID != userID {
		return nil, errors.New("unauthorized to update this post")
	}

	prevStatus := post.Status
	wasPublished := post.Status == models.StatusPublished
	post.Status = status

	if status == models.StatusPublished && !wasPublished {
		now := time.Now()
		post.PublishedAt = &now
		post.IsLatestNews = true
		post.LatestNewsAt = &now
		go s.notifySubscribers(post)
	}

	if status == models.StatusOffline && prevStatus == models.StatusPublished {
		go s.notifyPostTakenDown(post)
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

		p := post // copy so the goroutines capture a stable value
		go s.notifySubscribers(&p)
		go s.notifyScheduledPostLive(&p)
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
	totalAuthors, _ := s.userRepo.CountByRole(models.RoleAuthor)
	totalAdmins, _ := s.userRepo.CountByRole(models.RoleAdmin)

	return &models.DashboardStats{
		TotalPosts:       totalPosts,
		PublishedPosts:   publishedPosts,
		DraftPosts:       draftPosts,
		ScheduledPosts:   scheduledPosts,
		TotalViews:       totalViews,
		TotalSubscribers: totalSubscribers,
		TotalAuthors:     totalAuthors + totalAdmins,
	}, nil
}

func (s *PostService) Search(query string, page, perPage int) (*models.PaginatedResponse, error) {
	posts, total, err := s.postRepo.Search(query, page, perPage)
	if err != nil {
		return nil, err
	}
	return paginated(posts, total, page, perPage), nil
}

func paginated(data interface{}, total int64, page, perPage int) *models.PaginatedResponse {
	totalPages := int(total) / perPage
	if int(total)%perPage > 0 {
		totalPages++
	}
	return &models.PaginatedResponse{
		Data:       data,
		Total:      total,
		Page:       page,
		PerPage:    perPage,
		TotalPages: totalPages,
	}
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
		if err != nil {
			break
		}
		if post != nil && post.ID == excludeID {
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

// notifyAdminNewPost emails the admin whenever any post is created. Skips the
// alert if the admin created the post themselves.
func (s *PostService) notifyAdminNewPost(post *models.Post) {
	if s.emailService == nil || !s.emailService.IsNotificationEnabled() {
		return
	}
	admin, err := s.userRepo.FindAdmin()
	if err != nil || admin == nil || admin.ID == post.AuthorID {
		return
	}
	authorName := "an author"
	if post.Author != nil {
		authorName = post.Author.DisplayName
	}
	headline := authorName + " just created a post"
	body := "<strong>" + post.Title + "</strong> was created with status <em>" + string(post.Status) + "</em>. Hop into the admin panel if you'd like to review it."
	if err := s.emailService.SendAdminAlert(admin.Email, "New post: "+post.Title, headline, body); err != nil {
		log.Printf("notifyAdminNewPost: %v", err)
	}
}

// notifyScheduledPostLive emails the post's author + the admin when a
// scheduled post auto-publishes.
func (s *PostService) notifyScheduledPostLive(post *models.Post) {
	if s.emailService == nil || !s.emailService.IsNotificationEnabled() {
		return
	}
	// Notify the author.
	if post.Author != nil {
		author, err := s.userRepo.FindByID(post.AuthorID)
		if err == nil && author != nil {
			_ = s.emailService.SendAuthorPostStatus(
				author.Email,
				author.DisplayName,
				"Your scheduled post is live",
				"Your scheduled post just went live on the public site.",
				post.Title,
				post.Slug,
			)
		}
	}
	// FYI the admin (if they aren't the author).
	if admin, err := s.userRepo.FindAdmin(); err == nil && admin != nil && admin.ID != post.AuthorID {
		headline := "Scheduled post is live: " + post.Title
		body := "The scheduled post <strong>" + post.Title + "</strong> was automatically published."
		_ = s.emailService.SendAdminAlert(admin.Email, "Scheduled post live: "+post.Title, headline, body)
	}
}

// notifyPostTakenDown fires when a post moves from published -> offline.
func (s *PostService) notifyPostTakenDown(post *models.Post) {
	if s.emailService == nil || !s.emailService.IsNotificationEnabled() {
		return
	}
	if author, err := s.userRepo.FindByID(post.AuthorID); err == nil && author != nil {
		_ = s.emailService.SendAuthorPostStatus(
			author.Email,
			author.DisplayName,
			"Your post has been taken down",
			"Your post is no longer visible on the public site. Reach out to the admin if you think this was a mistake.",
			post.Title,
			post.Slug,
		)
	}
	if admin, err := s.userRepo.FindAdmin(); err == nil && admin != nil && admin.ID != post.AuthorID {
		headline := post.Title + " has been taken down"
		body := "The post <strong>" + post.Title + "</strong> was taken offline."
		_ = s.emailService.SendAdminAlert(admin.Email, "Post taken down: "+post.Title, headline, body)
	}
}

func (s *PostService) notifySubscribers(post *models.Post) {
	if s.emailService == nil || !s.emailService.IsNewsletterEnabled() {
		log.Println("Newsletter email not configured, skipping subscriber notification")
		return
	}

	subscribers, err := s.subscriberRepo.FindAllActive()
	if err != nil {
		log.Printf("Failed to fetch subscribers for notification: %v", err)
		return
	}

	if len(subscribers) == 0 {
		return
	}

	log.Printf("Notifying %d subscribers about new post: %s", len(subscribers), post.Title)

	for _, sub := range subscribers {
		err := s.emailService.SendNewPostNotification(
			sub.Email,
			post.Title,
			post.Slug,
			sub.UnsubscribeToken,
		)
		if err != nil {
			log.Printf("Failed to notify subscriber %s: %v", sub.Email, err)
		}
	}
}
