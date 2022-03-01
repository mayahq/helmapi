package api

import (
	"encoding/json"
	"log"
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

// DeleteReleaseHandler serves requests at /delete
func DeleteRuntimeHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data := struct {
			RuntimeIds []string `json:"runtimeIds"`
			Concurrent bool     `json:"concurrent"`
			Timeout    int      `json:"timeout"`
		}{}

		err := json.NewDecoder(r.Body).Decode(&data)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			log.Println(err)
			return
		}

		result := make(map[string]bool)
		if !data.Concurrent {
			for _, rId := range data.RuntimeIds {
				dr := client.DeleteRequest{
					ReleaseName: "rt-" + rId,
				}
				deleteErr := dr.Execute()
				if deleteErr != nil {
					log.Println(deleteErr)
					w.WriteHeader(http.StatusInternalServerError)
					result[rId] = false
					continue
				}
				result[rId] = true
			}
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		payload := struct {
			Status string `json:"status"`
		}{Status: "SUCCESS"}
		json.NewEncoder(w).Encode(payload)
	})
}
