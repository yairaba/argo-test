package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"

	// "strings"
	"time"

	"github.com/go-redis/redis/v8"
)

type ServiceData map[string]string
type ServiceDataWithKey struct {
	Repo        string      `json:"repo"`
	Branch      string      `json:"branch"`
	ServiceData ServiceData `json:"serviceData"`
}

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
		var params ServiceDataWithKey

		err := json.NewDecoder(r.Body).Decode(&params)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if params.Repo == "" || params.Branch == "" || len(params.ServiceData) == 0 {
			http.Error(w, "Missing required parameters", http.StatusBadRequest)
			return
		}

		key := fmt.Sprintf("%s:%s", params.Repo, params.Branch)

		id, err := client.Incr(context.Background(), "data:id").Result()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		params.ServiceData["id"] = strconv.FormatInt(id, 10)

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
			Repo   string `json:"repo"`
			Branch string `json:"branch"`
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

		if params.Repo == "" {
			http.Error(w, "Missing required parameter repo", http.StatusBadRequest)
			return
		}

		var keys []string
		if params.Branch == "" {
			redisKeys, err := client.Keys(context.Background(), fmt.Sprintf("%s:*", params.Repo)).Result()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			keys = redisKeys
		} else {
			key := fmt.Sprintf("%s:%s", params.Repo, params.Branch)
			keys = []string{key}
		}

		dataMaps := make([]ServiceDataWithKey, 0)
		for _, key := range keys {
			serviceData, err := client.HGetAll(context.Background(), key).Result()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			if len(serviceData) > 0 {
				splits := strings.Split(key, ":")

				data := ServiceDataWithKey{
					Repo:        splits[0],
					Branch:      splits[1],
					ServiceData: serviceData,
				}

				dataMaps = append(dataMaps, data)
			} else {
				http.NotFound(w, r)
				return
			}
		}

		sort.Slice(dataMaps, func(i, j int) bool {
			idI, _ := strconv.Atoi(dataMaps[i].ServiceData["id"])
			idJ, _ := strconv.Atoi(dataMaps[j].ServiceData["id"])
			return idI < idJ
		})

		jsonData, err := json.Marshal(dataMaps)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonData)

		fmt.Print("Response : \n", params)
		fmt.Printf("%+v\n", dataMaps)
		fmt.Println("------------------------------------------")
	})

	log.Fatal(http.ListenAndServe(":8080", r))
}
