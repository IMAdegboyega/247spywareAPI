package services

import (
	"log"
	"time"
)

// StartScheduler starts background jobs for scheduled posts and latest news expiry
func StartScheduler(postService *PostService) {
	// Check for scheduled posts every minute
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()

		for range ticker.C {
			if err := postService.PublishScheduledPosts(); err != nil {
				log.Printf("Error publishing scheduled posts: %v", err)
			}
		}
	}()

	// Check for expired latest news every hour
	// Default expiry is 3 weeks (21 days)
	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()

		latestNewsDuration := 21 * 24 * time.Hour // 3 weeks

		for range ticker.C {
			if err := postService.ExpireLatestNews(latestNewsDuration); err != nil {
				log.Printf("Error expiring latest news: %v", err)
			}
		}
	}()

	log.Println("Scheduler started")
}