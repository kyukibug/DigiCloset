package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"com.fukubox/database"
	"github.com/go-chi/chi"
)

type Clothes struct {
	Id         int       `json:"id"`
	UserId     int       `json:"user_id"`
	CategoryId int       `json:"category_id"`
	ImageUrl   string    `json:"image_url"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type ClothesName struct {
	CategoryId int    `json:"category_id"`
	ImageUrl   string `json:"image_url"`
}

func GetClothes(w http.ResponseWriter, r *http.Request) {
	dbpool := database.GetDB()
	ctx := r.Context()

	userIdStr := r.Header.Get("userId")
	if userIdStr == "" {
		http.Error(w, "User ID is not set", http.StatusBadRequest)
		return
	}
	userId, err := strconv.Atoi(userIdStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	conn, err := dbpool.Acquire(ctx)
	if err != nil {
		log.Printf("Failed to acquire a database connection: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer conn.Release()

	rows, err := conn.Query(ctx, "SELECT * FROM clothing_items WHERE user_id = $1", userId)
	if err != nil {
		log.Printf("Query failed: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	clothes := []Clothes{}

	for rows.Next() {
		var cloth Clothes
		if err := rows.Scan(&cloth.Id, &cloth.UserId, &cloth.CategoryId, &cloth.ImageUrl, &cloth.CreatedAt, &cloth.UpdatedAt); err != nil {
			log.Printf("Failed to scan row: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		clothes = append(clothes, cloth)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Error after iterating rows: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(clothes); err != nil {
		log.Printf("Failed to encode response as JSON: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func GetClothesById(w http.ResponseWriter, r *http.Request) {
	dbpool := database.GetDB()
	ctx := r.Context()

	userIdStr := r.Header.Get("userId")
	if userIdStr == "" {
		http.Error(w, "User ID is not set", http.StatusBadRequest)
		return
	}
	userId, err := strconv.Atoi(userIdStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	clothIdStr := chi.URLParam(r, "id")
	clothId, err := strconv.Atoi(clothIdStr)
	if err != nil {
		http.Error(w, "Invalid clothing item ID", http.StatusBadRequest)
		return
	}

	conn, err := dbpool.Acquire(ctx)
	if err != nil {
		log.Printf("Failed to acquire a database connection: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer conn.Release()

	var cloth Clothes
	err = dbpool.QueryRow(ctx, "SELECT id, user_id, category_id, image_url, created_at, updated_at FROM clothing_items WHERE id = $1 AND user_id = $2", clothId, userId).
		Scan(&cloth.Id, &cloth.UserId, &cloth.CategoryId, &cloth.ImageUrl, &cloth.CreatedAt, &cloth.UpdatedAt)
	if err != nil {
		log.Printf("Failed to query row: %v", err)
		http.Error(w, "Clothing Item not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(cloth); err != nil {
		log.Printf("Failed to encode response as JSON: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
