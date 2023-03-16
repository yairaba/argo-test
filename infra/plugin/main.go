package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	// "strings"
	"time"

	"github.com/go-redis/redis/v8"
)

type ServiceData map[string]string

func main() {
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	client := redis.NewClient(&redis.Options{
		Addr:        redisAddr,
		MaxRetries:  3,
		DialTimeout: 5 * time.Second,
	})

	r := http.NewServeMux()

	const expiration = 604800

	r.HandleFunc("/update", func(w http.ResponseWriter, r *http.Request) {
		var params struct {
			RepoName    string      `json:"repoName"`
			BranchName  string      `json:"branchName"`
			ServiceData ServiceData `json:"serviceData"`
		}

		err := json.NewDecoder(r.Body).Decode(&params)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if params.RepoName == "" || params.BranchName == "" || len(params.ServiceData) == 0 {
			http.Error(w, "Missing required parameters", http.StatusBadRequest)
			return
		}

		key := fmt.Sprintf("%s:%s", params.RepoName, params.BranchName)

		data := make([]interface{}, 0, len(params.ServiceData)*2)
		for k, v := range params.ServiceData {
			data = append(data, k, v)
		}

		err = client.HMSet(context.Background(), key, data...).Err()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		err = client.Expire(context.Background(), key, expiration*time.Second).Err()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, "Successfully updated Redis with key %s", key)
	})

	r.HandleFunc("/api/v1/template.execute", func(w http.ResponseWriter, r *http.Request) {

		// authHeader := r.Header.Get("Authorization")
		// expectedToken := "my-secret-token"
		// if !strings.HasPrefix(authHeader, "Bearer ") || authHeader[7:] != expectedToken {
		// 	http.Error(w, "Unauthorized", http.StatusUnauthorized)
		// 	return
		// }

		var params struct {
			RepoName   string `json:"repoName"`
			BranchName string `json:"branchName"`
		}
		err := json.NewDecoder(r.Body).Decode(&params)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		fmt.Println("Incoming request to /api/v1/template.execute")
		fmt.Println(time.Now().Format("2006-01-02 15:04:05"), "Incoming request to /api/v1/template.execute")

		// Print headers for debugging purposes
		for name, values := range r.Header {
			for _, value := range values {
				fmt.Printf("%s: %s\n", name, value)
			}
		}
		fmt.Print("Request : \n", params)
		fmt.Printf("%+v\n", params)

		if params.RepoName == "" || params.BranchName == "" {
			http.Error(w, "Missing required parameters", http.StatusBadRequest)
			return
		}

		key := fmt.Sprintf("%s:%s", params.RepoName, params.BranchName)

		data, err := client.HGetAll(context.Background(), key).Result()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// serviceData := ServiceData(data)

		jsonData, err := json.Marshal([]map[string]string{data})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonData)

		fmt.Print("Response : \n", params)
		fmt.Printf("%+v\n", data)
		fmt.Println("------------------------------------------")
	})

	log.Fatal(http.ListenAndServe(":8080", r))
}
