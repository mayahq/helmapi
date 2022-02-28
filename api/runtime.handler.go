package api

import (
	"encoding/json"
	"net/http"

	"github.com/dush-t/helmapi/client"
)

func RestartRuntimeHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// data := make(map[string]interface{})
		data := struct {
			RuntimeIds []string `json:"runtimeIds"`
			Concurrent bool     `json:"concurrent"`
			Timeout    string   `json:"timeout"`
		}{}

		err := json.NewDecoder(r.Body).Decode(&data)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		concurrent := data.Concurrent
		runtimeIds := data.RuntimeIds
		timeout := data.Timeout

		result := make(map[string]bool)

		n := len(runtimeIds)
		if !concurrent {
			for i := 0; i < n; i++ {
				err = client.RestartRuntime(runtimeIds[i], timeout)
				if err != nil {
					result[runtimeIds[i]] = false
					continue
				}
				result[runtimeIds[i]] = true
			}

			payload := struct {
				Restarted map[string]bool `json:"restarted"`
			}{Restarted: result}
			json.NewEncoder(w).Encode(payload)
			return
		}
	})
}
