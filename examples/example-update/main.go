package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// User represents the users table
type User struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Username  string    `gorm:"uniqueIndex;not null" json:"username"`
	Email     string    `gorm:"uniqueIndex;not null" json:"email"`
	FullName  string    `gorm:"column:full_name" json:"full_name"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
	Posts     []Post    `gorm:"foreignKey:UserID" json:"posts,omitempty"`
}

// Post represents the posts table
type Post struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Title     string    `gorm:"not null" json:"title"`
	Content   string    `json:"content"`
	UserID    uint      `gorm:"not null" json:"user_id"`
	Status    string    `gorm:"default:draft" json:"status"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
	Tags      []Tag     `gorm:"many2many:post_tags" json:"tags,omitempty"`
}

// Tag represents the tags table
type Tag struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Name      string    `gorm:"uniqueIndex;not null" json:"name"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	Posts     []Post    `gorm:"many2many:post_tags" json:"posts,omitempty"`
}

// GenericUpdate performs a partial update on any model
func GenericUpdate[T any](db *gorm.DB, id uint, updates map[string]interface{}) (*T, error) {
	var model T

	// First check if the record exists
	result := db.First(&model, id)
	if result.Error != nil {
		return nil, fmt.Errorf("record not found: %v", result.Error)
	}

	// Get the field names from the updates map
	var fields []string
	for field := range updates {
		fields = append(fields, field)
	}

	// Perform the update using only the specified fields
	result = db.Model(&model).Select(fields).Updates(updates)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to update: %v", result.Error)
	}

	// Retrieve the updated record
	result = db.First(&model, id)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to retrieve updated record: %v", result.Error)
	}

	return &model, nil
}

type Server struct {
	db *gorm.DB
}

func (s *Server) handleCreateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	result := s.db.Create(&user)
	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func (s *Server) handleUpdateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract user ID from URL path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) != 3 {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}

	userID, err := strconv.ParseUint(pathParts[2], 10, 32)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	updatedUser, err := GenericUpdate[User](s.db, uint(userID), updates)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedUser)
}

func (s *Server) handleGetUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract user ID from URL path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) != 3 {
		http.Error(w, "Invalid URL", http.StatusBadRequest)
		return
	}

	userID, err := strconv.ParseUint(pathParts[2], 10, 32)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	var user User
	result := s.db.Preload("Posts.Tags").First(&user, userID)
	if result.Error != nil {
		http.Error(w, result.Error.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func main() {
	// Set up GORM logger for SQL statements
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold:             time.Second,   // Slow SQL threshold
			LogLevel:                  logger.Info,   // Log level (show all queries)
			IgnoreRecordNotFoundError: false,        // Include not found errors
			Colorful:                  true,         // Enable colors
		},
	)

	// Initialize database connection with logger
	dsn := "host=localhost user=sqld_user password=sqld_password dbname=sqld_example port=5101 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Get the underlying SQL DB
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal("Failed to get underlying *sql.DB:", err)
	}
	defer sqlDB.Close()

	server := &Server{db: db}

	// Set up routes
	http.HandleFunc("/users", server.handleCreateUser)
	http.HandleFunc("/users/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			server.handleGetUser(w, r)
		case http.MethodPatch:
			server.handleUpdateUser(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Start the server
	log.Println("Server starting on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
