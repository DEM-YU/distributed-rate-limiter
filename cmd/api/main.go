package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"rate-limiter/internal/limiter"

	"github.com/redis/go-redis/v9"
)

func main() {
	// Initialize Redis client
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	defer rdb.Close()

	// Check connection
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		fmt.Printf("Failed to connect to Redis: %v\n", err)
		return
	}

	// Initialize Limiter
	// Capacity: 5, Rate: 1 token/second
	capacity := 5
	rate := 1.0
	l := limiter.NewTokenBucketLimiter(rdb, capacity, rate)

	http.HandleFunc("/api/data", func(w http.ResponseWriter, r *http.Request) {
		// Only allow GET method
		if r.Method != http.MethodGet {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		// Extract user_id from query parameters
		userID := r.URL.Query().Get("user_id")
		if userID == "" {
			userID = "anonymous"
		}

		// Check rate limit
		ctx, cancel := context.WithTimeout(r.Context(), 500*time.Millisecond)
		defer cancel()

		allowed, err := l.Allow(ctx, userID)
		if err != nil {
			http.Error(w, fmt.Sprintf("Internal Server Error: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")

		if allowed {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]string{
				"status": "success",
				"data":   "这里是核心业务数据",
			})
		} else {
			w.WriteHeader(http.StatusTooManyRequests)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Too Many Requests, please try again later.",
			})
		}
	})

	fmt.Println("🚀 服务器已启动在 http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Printf("Server failed: %v\n", err)
	}
}
