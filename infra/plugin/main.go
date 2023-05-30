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

	r.HandleFunc("/api/v1/getparams.execute", func(w http.ResponseWriter, r *http.Request) {

		// authHeader := r.Header.Get("Authorization")
		// expectedToken := "my-secret-token"
		// if !strings.HasPrefix(authHeader, "Bearer ") || authHeader[7:] != expectedToken {
		// 	http.Error(w, "Unauthorized", http.StatusUnauthorized)
		// 	return
		// }

		type ParametersRequest struct {
			Repo   string `json:"repo"`
			Branch string `json:"branch"`
		}

		type PluginRequest struct {
			ApplicationSetName string            `json:"applicationSetName"`
			Parameters         ParametersRequest `json:"inputParameters"`
		}

		type Output struct {
			// Parameters is the list of parameter sets returned by the plugin.
			Parameters []ServiceDataWithKey `json:"parameters"`
		}

		// ServiceResponse is the response object returned by the plugin service.
		type PluginResponse struct {
			// Output is the map of outputs returned by the plugin.
			Output Output `json:"output"`
		}

		var pluginRequest PluginRequest

		err := json.NewDecoder(r.Body).Decode(&pluginRequest)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		fmt.Println("Incoming request to /api/v1/getparams.execute")
		fmt.Println(time.Now().Format("2006-01-02 15:04:05"), "Incoming request to /api/v1/getparams.execute")

		// Print headers for debugging purposes
		for name, values := range r.Header {
			for _, value := range values {
				fmt.Printf("%s: %s\n", name, value)
			}
		}
		fmt.Print("Request : \n", pluginRequest)
		fmt.Printf("%+v\n", pluginRequest)

		if pluginRequest.Parameters.Repo == "" {
			http.Error(w, "Missing required parameter repo", http.StatusBadRequest)
			return
		}

		var keys []string
		if pluginRequest.Parameters.Branch == "" {
			redisKeys, err := client.Keys(context.Background(), fmt.Sprintf("%s:*", pluginRequest.Parameters.Repo)).Result()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			keys = redisKeys
		} else {
			key := fmt.Sprintf("%s:%s", pluginRequest.Parameters.Repo, pluginRequest.Parameters.Branch)
			keys = []string{key}
		}

		// if len(keys) == 0 {
		// 	fmt.Println("Not found keys")
		// 	http.NotFound(w, r)
		// 	return
		// }

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
			}
			// else {
			// 	fmt.Println("Not found serviceData")
			// 	http.NotFound(w, r)
			// 	return
			// }
		}

		sort.Slice(dataMaps, func(i, j int) bool {
			idI, _ := strconv.Atoi(dataMaps[i].ServiceData["id"])
			idJ, _ := strconv.Atoi(dataMaps[j].ServiceData["id"])
			return idI < idJ
		})

		response := PluginResponse{
			Output{
				Parameters: dataMaps,
			},
		}

		jsonData, err := json.Marshal(response)

		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonData)

		fmt.Print("Response : \n", pluginRequest)
		fmt.Printf("%+v\n", response)
		fmt.Println("------------------------------------------")
	})

	log.Fatal(http.ListenAndServe(":8080", r))
}
