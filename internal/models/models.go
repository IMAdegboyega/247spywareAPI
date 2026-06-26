package models

import (
	"time"
)

// UserRole represents the role of a user
type UserRole string

const (
	RoleAdmin  UserRole = "admin"
	RoleAuthor UserRole = "author"
)

// PostStatus represents the status of a post
type PostStatus string

const (
	StatusDraft     PostStatus = "draft"
	StatusPublished PostStatus = "published"
	StatusScheduled PostStatus = "scheduled"
	StatusOffline   PostStatus = "offline"
)

// User represents a user in the system (admin or temporary author).
// The document uses a sequential integer `_id` instead of an ObjectID so the
// frontend's existing `number` IDs keep working.
type User struct {
	ID                  uint       `bson:"_id" json:"id"`
	Email               string     `bson:"email" json:"email"`
	Password            string     `bson:"password" json:"-"`
	DisplayName         string     `bson:"display_name" json:"display_name"`
	Role                UserRole   `bson:"role" json:"role"`
	IsActive            bool       `bson:"is_active" json:"is_active"`
	ResetToken          string     `bson:"reset_token,omitempty" json:"-"`
	ResetTokenExpiresAt *time.Time `bson:"reset_token_expires_at,omitempty" json:"-"`
	CreatedAt           time.Time  `bson:"created_at" json:"created_at"`
	UpdatedAt           time.Time  `bson:"updated_at" json:"updated_at"`
}

// Category represents a blog category.
type Category struct {
	ID        uint      `bson:"_id" json:"id"`
	Name      string    `bson:"name" json:"name"`
	Slug      string    `bson:"slug" json:"slug"`
	CreatedAt time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time `bson:"updated_at" json:"updated_at"`
}

// EmbeddedAuthor is a denormalised snapshot of the user shown on a Post.
// Only the fields the public site actually renders are kept.
type EmbeddedAuthor struct {
	ID          uint   `bson:"id" json:"id"`
	DisplayName string `bson:"display_name" json:"display_name"`
}

// EmbeddedCategory is a denormalised snapshot of the category shown on a Post.
type EmbeddedCategory struct {
	ID   uint   `bson:"id" json:"id"`
	Name string `bson:"name" json:"name"`
	Slug string `bson:"slug" json:"slug"`
}

// Post represents a blog post. Author + category are stored both as an
// indexed scalar id (for queries / propagation) and as an embedded summary
// (for cheap reads on the public site).
type Post struct {
	ID           uint              `bson:"_id" json:"id"`
	Title        string            `bson:"title" json:"title"`
	Slug         string            `bson:"slug" json:"slug"`
	Content      string            `bson:"content" json:"content"`
	Excerpt      string            `bson:"excerpt" json:"excerpt"`
	BannerImage  string            `bson:"banner_image" json:"banner_image"`
	Status       PostStatus        `bson:"status" json:"status"`
	ViewCount    uint              `bson:"view_count" json:"view_count"`
	AuthorID     uint              `bson:"author_id" json:"author_id"`
	Author       *EmbeddedAuthor   `bson:"author,omitempty" json:"author,omitempty"`
	CategoryID   *uint             `bson:"category_id,omitempty" json:"category_id"`
	Category     *EmbeddedCategory `bson:"category,omitempty" json:"category,omitempty"`
	PublishedAt  *time.Time        `bson:"published_at,omitempty" json:"published_at"`
	ScheduledFor *time.Time        `bson:"scheduled_for,omitempty" json:"scheduled_for"`
	IsLatestNews bool              `bson:"is_latest_news" json:"is_latest_news"`
	LatestNewsAt *time.Time        `bson:"latest_news_at,omitempty" json:"latest_news_at"`
	CreatedAt    time.Time         `bson:"created_at" json:"created_at"`
	UpdatedAt    time.Time         `bson:"updated_at" json:"updated_at"`
}

// EditorPick stores a reference to a post. The post itself is fetched via
// $lookup when listing picks so renames / banner-image changes propagate
// automatically.
type EditorPick struct {
	ID        uint      `bson:"_id" json:"id"`
	PostID    uint      `bson:"post_id" json:"post_id"`
	Post      *Post     `bson:"post,omitempty" json:"post,omitempty"`
	PickedAt  time.Time `bson:"picked_at" json:"picked_at"`
	Order     int       `bson:"order" json:"order"`
	CreatedAt time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time `bson:"updated_at" json:"updated_at"`
}

// Subscriber represents a newsletter subscriber.
type Subscriber struct {
	ID               uint      `bson:"_id" json:"id"`
	Email            string    `bson:"email" json:"email"`
	IsActive         bool      `bson:"is_active" json:"is_active"`
	UnsubscribeToken string    `bson:"unsubscribe_token" json:"-"`
	SubscribedAt     time.Time `bson:"subscribed_at" json:"subscribed_at"`
	CreatedAt        time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt        time.Time `bson:"updated_at" json:"updated_at"`
}

// Advert represents an advertisement.
type Advert struct {
	ID         uint       `bson:"_id" json:"id"`
	Title      string     `bson:"title" json:"title"`
	ImageURL   string     `bson:"image_url" json:"image_url"`
	LinkURL    string     `bson:"link_url" json:"link_url"`
	Position   string     `bson:"position" json:"position"`
	IsActive   bool       `bson:"is_active" json:"is_active"`
	StartDate  *time.Time `bson:"start_date,omitempty" json:"start_date"`
	EndDate    *time.Time `bson:"end_date,omitempty" json:"end_date"`
	ClickCount uint       `bson:"click_count" json:"click_count"`
	ViewCount  uint       `bson:"view_count" json:"view_count"`
	CreatedAt  time.Time  `bson:"created_at" json:"created_at"`
	UpdatedAt  time.Time  `bson:"updated_at" json:"updated_at"`
}

// Counter is the document type for the sequential-id generator.
type Counter struct {
	ID  string `bson:"_id"`
	Seq uint   `bson:"seq"`
}

// ---------------------------------------------------------------------------
// DTOs (Data Transfer Objects)
// ---------------------------------------------------------------------------

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

type CreateUserRequest struct {
	Email       string   `json:"email" binding:"required,email"`
	Password    string   `json:"password" binding:"required,min=6"`
	DisplayName string   `json:"display_name" binding:"required"`
	Role        UserRole `json:"role"`
}

type UpdateUserRequest struct {
	Email       string   `json:"email" binding:"omitempty,email"`
	Password    string   `json:"password" binding:"omitempty,min=6"`
	DisplayName string   `json:"display_name"`
	Role        UserRole `json:"role"`
	IsActive    *bool    `json:"is_active"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type ResetPasswordRequest struct {
	Token       string `json:"token" binding:"required"`
	Email       string `json:"email" binding:"required,email"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

type CreatePostRequest struct {
	Title        string     `json:"title" binding:"required"`
	Content      string     `json:"content"`
	Excerpt      string     `json:"excerpt"`
	CategoryID   *uint      `json:"category_id"`
	Status       PostStatus `json:"status"`
	ScheduledFor *time.Time `json:"scheduled_for"`
}

type UpdatePostRequest struct {
	Title        string     `json:"title"`
	Content      string     `json:"content"`
	Excerpt      string     `json:"excerpt"`
	CategoryID   *uint      `json:"category_id"`
	Status       PostStatus `json:"status"`
	ScheduledFor *time.Time `json:"scheduled_for"`
	BannerImage  string     `json:"banner_image"`
}

type UpdatePostStatusRequest struct {
	Status PostStatus `json:"status" binding:"required"`
}

type CreateCategoryRequest struct {
	Name string `json:"name" binding:"required"`
}

type UpdateCategoryRequest struct {
	Name string `json:"name" binding:"required"`
}

type SubscribeRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type CreateAdvertRequest struct {
	Title     string     `json:"title" binding:"required"`
	ImageURL  string     `json:"image_url"`
	LinkURL   string     `json:"link_url"`
	Position  string     `json:"position"`
	IsActive  bool       `json:"is_active"`
	StartDate *time.Time `json:"start_date"`
	EndDate   *time.Time `json:"end_date"`
}

type UpdateAdvertRequest struct {
	Title     string     `json:"title"`
	ImageURL  string     `json:"image_url"`
	LinkURL   string     `json:"link_url"`
	Position  string     `json:"position"`
	IsActive  *bool      `json:"is_active"`
	StartDate *time.Time `json:"start_date"`
	EndDate   *time.Time `json:"end_date"`
}

type AddEditorPickRequest struct {
	PostID uint `json:"post_id" binding:"required"`
}

type ReorderEditorPicksRequest struct {
	Picks []struct {
		ID    uint `json:"id"`
		Order int  `json:"order"`
	} `json:"picks"`
}

type DashboardStats struct {
	TotalPosts       int64 `json:"total_posts"`
	PublishedPosts   int64 `json:"published_posts"`
	DraftPosts       int64 `json:"draft_posts"`
	ScheduledPosts   int64 `json:"scheduled_posts"`
	TotalViews       int64 `json:"total_views"`
	TotalSubscribers int64 `json:"total_subscribers"`
	TotalAuthors     int64 `json:"total_authors"`
}

type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Total      int64       `json:"total"`
	Page       int         `json:"page"`
	PerPage    int         `json:"per_page"`
	TotalPages int         `json:"total_pages"`
}
