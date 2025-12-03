# Blog Backend API

A full-featured blog backend built with Go, Gin, and GORM.

## Features

### Content Management
- **Posts**: Full CRUD with draft, published, scheduled, and offline states
- **Categories**: Organize posts by category
- **Editor's Picks**: Curate featured posts, ordered by selection date
- **Latest News**: Auto-expires after configurable period (default 3 weeks)
- **Auto-save**: Drafts are automatically saved

### User Management
- **Admin**: Full control over all features
- **Authors**: Temporary accounts for guest writers
- **Attribution**: Posts show "Written by [Author Name]"
- **Account Control**: Admin can activate/deactivate authors anytime

### Media & Content
- **Image Upload**: Banner images for posts (supports jpg, png, gif, webp)
- **View Tracking**: Track how many people viewed each post

### Engagement
- **Newsletter**: Email subscription system with unsubscribe tokens
- **Advertisements**: Flexible ad placement with scheduling and analytics

## Project Structure

```
blog-backend/
├── cmd/
│   ├── server/
│   │   └── main.go          # Main application entry point
│   └── seed/
│       └── main.go          # Database seeder
├── internal/
│   ├── config/
│   │   └── database.go      # Database configuration
│   ├── handlers/            # HTTP handlers
│   ├── middleware/          # Auth & CORS middleware
│   ├── models/              # Data models & DTOs
│   ├── repository/          # Database operations
│   └── services/            # Business logic
├── .env.example             # Environment template
├── go.mod                   # Go modules
└── README.md
```

## Quick Start

### 1. Clone and Setup

```bash
cd blog-backend
cp .env.example .env
# Edit .env with your settings
```

### 2. Install Dependencies

```bash
go mod tidy
```

### 3. Seed Database

```bash
go run cmd/seed/main.go
```

This creates:
- Default admin user (admin@blog.com / admin123)
- Default categories

### 4. Run Server

```bash
go run cmd/server/main.go
```

Server starts at `http://localhost:8080`

## API Endpoints

### Public Routes

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/auth/login` | Login |
| GET | `/api/posts` | Get published posts |
| GET | `/api/posts/:slug` | Get post by slug |
| GET | `/api/posts/latest` | Get latest news |
| GET | `/api/posts/category/:slug` | Get posts by category |
| GET | `/api/categories` | Get all categories |
| GET | `/api/editor-picks` | Get editor's picks |
| GET | `/api/adverts/active` | Get active advertisements |
| POST | `/api/subscribe` | Subscribe to newsletter |
| GET | `/api/unsubscribe?token=xxx` | Unsubscribe |

### Protected Routes (Requires Auth)

All routes under `/api/admin` require `Authorization: Bearer <token>` header.

#### User Management (Admin Only)

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/admin/users` | Create user/author |
| GET | `/api/admin/users` | Get all users |
| PUT | `/api/admin/users/:id` | Update user |
| DELETE | `/api/admin/users/:id` | Delete user |
| PUT | `/api/admin/users/:id/toggle-active` | Activate/deactivate user |

#### Category Management (Admin Only)

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/admin/categories` | Create category |
| PUT | `/api/admin/categories/:id` | Update category |
| DELETE | `/api/admin/categories/:id` | Delete category |

#### Post Management

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/admin/posts` | Get all posts (inc. drafts) |
| GET | `/api/admin/posts/:id` | Get post by ID |
| POST | `/api/admin/posts` | Create post |
| PUT | `/api/admin/posts/:id` | Update post |
| DELETE | `/api/admin/posts/:id` | Delete post (Admin only) |
| PUT | `/api/admin/posts/:id/status` | Update post status |
| PUT | `/api/admin/posts/:id/auto-save` | Auto-save draft |
| POST | `/api/admin/posts/:id/upload-image` | Upload banner image |

#### Editor's Picks (Admin Only)

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/admin/editor-picks` | Add editor's pick |
| DELETE | `/api/admin/editor-picks/:id` | Remove editor's pick |
| PUT | `/api/admin/editor-picks/reorder` | Reorder picks |

#### Subscribers (Admin Only)

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/admin/subscribers` | Get all subscribers |
| DELETE | `/api/admin/subscribers/:id` | Delete subscriber |

#### Advertisements (Admin Only)

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/admin/adverts` | Get all adverts |
| POST | `/api/admin/adverts` | Create advert |
| PUT | `/api/admin/adverts/:id` | Update advert |
| DELETE | `/api/admin/adverts/:id` | Delete advert |

#### Dashboard

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/admin/stats` | Get dashboard statistics |

## Request/Response Examples

### Login

```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "admin@blog.com", "password": "admin123"}'
```

Response:
```json
{
  "token": "eyJhbGciOiJIUzI1NiIs...",
  "user": {
    "id": 1,
    "email": "admin@blog.com",
    "display_name": "Admin",
    "role": "admin"
  }
}
```

### Create Post

```bash
curl -X POST http://localhost:8080/api/admin/posts \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "My First Post",
    "content": "This is the content...",
    "excerpt": "A brief summary",
    "category_id": 1,
    "status": "draft"
  }'
```

### Schedule Post

```bash
curl -X POST http://localhost:8080/api/admin/posts \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Scheduled Post",
    "content": "Content here...",
    "status": "scheduled",
    "scheduled_for": "2024-02-01T10:00:00Z"
  }'
```

### Create Temporary Author

```bash
curl -X POST http://localhost:8080/api/admin/users \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "author@example.com",
    "password": "temppass123",
    "display_name": "John Writer",
    "role": "author"
  }'
```

### Upload Banner Image

```bash
curl -X POST http://localhost:8080/api/admin/posts/1/upload-image \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -F "image=@/path/to/image.jpg"
```

### Add Editor's Pick

```bash
curl -X POST http://localhost:8080/api/admin/editor-picks \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"post_id": 1}'
```

## Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | Server port | 8080 |
| `JWT_SECRET` | Secret for JWT tokens | (required) |
| `DB_TYPE` | Database type (sqlite/postgres) | sqlite |
| `DB_PATH` | SQLite database path | ./blog.db |
| `DB_HOST` | PostgreSQL host | localhost |
| `DB_PORT` | PostgreSQL port | 5432 |
| `DB_USER` | PostgreSQL user | postgres |
| `DB_PASSWORD` | PostgreSQL password | - |
| `DB_NAME` | PostgreSQL database name | - |
| `DB_SSLMODE` | PostgreSQL SSL mode | disable |
| `UPLOAD_PATH` | Image upload directory | ./uploads |

## Post Status Flow

```
draft → published
draft → scheduled → published (automatic)
published → offline
offline → published
```

## Latest News Auto-Expiry

Posts marked as "Latest News" automatically expire after 3 weeks (configurable). The scheduler runs hourly to check for expired posts.

## Production Deployment

1. Use PostgreSQL instead of SQLite
2. Set a strong `JWT_SECRET`
3. Configure proper CORS origins
4. Set up a reverse proxy (nginx)
5. Use environment variables for all secrets
6. Consider using S3/Cloudinary for image storage

## License

MIT