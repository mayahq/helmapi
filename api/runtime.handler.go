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
		} else {
			resultChan := make(chan []string)
			defer close(resultChan)

			for _, rId := range runtimeIds {
				runtimeId := rId
				go func() {
					err := client.RestartRuntime(runtimeId, timeout)
					if err != nil {
						resultChan <- []string{runtimeId, "FAIL"}
						return
					}

					resultChan <- []string{runtimeId, "SUCCESS"}
				}()
			}

			i := 0
			for i < len(runtimeIds) {
				d := <-resultChan
				if d[1] == "SUCCESS" {
					result[d[0]] = true
				} else {
					result[d[0]] = false
				}
				i += 1
			}
		}

		payload := struct {
			Restarted map[string]bool `json:"restarted"`
		}{Restarted: result}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(payload)
	})
}

// DeleteReleaseHandler serves requests at /delete
func DeleteRuntimeHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data := struct {
			RuntimeIds []string `json:"runtimeIds"`
			Concurrent bool     `json:"concurrent"`
			Timeout    string   `json:"timeout"`
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
				deleteErr := dr.Execute(data.Timeout)
				if deleteErr != nil {
					log.Println(deleteErr)
					result[rId] = false
					continue
				}
				result[rId] = true
			}
		} else {
			resultChan := make(chan []string)
			defer close(resultChan)

			for _, rId := range data.RuntimeIds {
				runtimeId := rId
				go func() {
					dr := client.DeleteRequest{
						ReleaseName: "rt-" + runtimeId,
					}
					err := dr.Execute(data.Timeout)
					if err != nil {
						log.Println("Delete error", runtimeId, err)
						resultChan <- []string{runtimeId, "FAIL"}
						return
					}

					resultChan <- []string{runtimeId, "SUCCESS"}
				}()
			}

			i := 0
			for i < len(data.RuntimeIds) {
				d := <-resultChan
				if d[1] == "SUCCESS" {
					result[d[0]] = true
				} else {
					result[d[0]] = false
				}
				i += 1
			}
		}

		payload := struct {
			Stopped map[string]bool `json:"stopped"`
		}{
			Stopped: result,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(payload)
	})
}
