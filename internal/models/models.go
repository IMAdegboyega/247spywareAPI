package models

import (
	"time"

	"gorm.io/gorm"
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

// User represents a user in the system (admin or temporary author)
type User struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Email       string         `gorm:"uniqueIndex;not null" json:"email"`
	Password    string         `gorm:"not null" json:"-"`
	DisplayName string         `gorm:"not null" json:"display_name"`
	Role        UserRole       `gorm:"type:varchar(20);default:'author'" json:"role"`
	IsActive    bool           `gorm:"default:true" json:"is_active"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
	Posts       []Post         `gorm:"foreignKey:AuthorID" json:"posts,omitempty"`
}

// Category represents a blog category
type Category struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Name      string         `gorm:"uniqueIndex;not null" json:"name"`
	Slug      string         `gorm:"uniqueIndex;not null" json:"slug"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	Posts     []Post         `gorm:"foreignKey:CategoryID" json:"posts,omitempty"`
}

// Post represents a blog post
type Post struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	Title        string         `gorm:"not null" json:"title"`
	Slug         string         `gorm:"uniqueIndex;not null" json:"slug"`
	Content      string         `gorm:"type:text" json:"content"`
	Excerpt      string         `gorm:"type:text" json:"excerpt"`
	BannerImage  string         `json:"banner_image"`
	Status       PostStatus     `gorm:"type:varchar(20);default:'draft'" json:"status"`
	ViewCount    uint           `gorm:"default:0" json:"view_count"`
	AuthorID     uint           `gorm:"not null" json:"author_id"`
	Author       User           `gorm:"foreignKey:AuthorID" json:"author,omitempty"`
	CategoryID   *uint          `json:"category_id"`
	Category     *Category      `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
	PublishedAt  *time.Time     `json:"published_at"`
	ScheduledFor *time.Time     `json:"scheduled_for"`
	IsLatestNews bool           `gorm:"default:false" json:"is_latest_news"`
	LatestNewsAt *time.Time     `json:"latest_news_at"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

// EditorPick represents a post selected as editor's pick
type EditorPick struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	PostID    uint           `gorm:"uniqueIndex;not null" json:"post_id"`
	Post      Post           `gorm:"foreignKey:PostID" json:"post,omitempty"`
	PickedAt  time.Time      `gorm:"not null" json:"picked_at"`
	Order     int            `gorm:"default:0" json:"order"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// Subscriber represents a newsletter subscriber
type Subscriber struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	Email        string         `gorm:"uniqueIndex;not null" json:"email"`
	IsActive     bool           `gorm:"default:true" json:"is_active"`
	UnsubscribeToken string     `gorm:"uniqueIndex" json:"-"`
	SubscribedAt time.Time      `json:"subscribed_at"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
}

// Advert represents an advertisement
type Advert struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Title       string         `gorm:"not null" json:"title"`
	ImageURL    string         `json:"image_url"`
	LinkURL     string         `json:"link_url"`
	Position    string         `gorm:"type:varchar(50)" json:"position"` // e.g., "sidebar", "banner", "inline"
	IsActive    bool           `gorm:"default:true" json:"is_active"`
	StartDate   *time.Time     `json:"start_date"`
	EndDate     *time.Time     `json:"end_date"`
	ClickCount  uint           `gorm:"default:0" json:"click_count"`
	ViewCount   uint           `gorm:"default:0" json:"view_count"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// DTOs (Data Transfer Objects)

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