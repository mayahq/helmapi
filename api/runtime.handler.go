package api

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/dush-t/helmapi/client"
	"github.com/dush-t/helmapi/client/k8s"
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

func FetchRuntimePodsHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data := struct {
			Users     []string `json:"users"`
			Namespace string   `json:"namespace"`
			Limit     int64    `json:"limit"`
			Continue  string   `json:"continue"`
		}{}

		err := json.NewDecoder(r.Body).Decode(&data)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		ctx := context.Background()
		var selector string
		if len(data.Users) > 0 {
			selector = "userRuntimeOwner in (" + strings.Join(data.Users, ",") + "),mayaResourceType=userRuntime"
		} else {
			selector = "mayaResourceType=userRuntime"
		}
		pods, perr := k8s.GetPodsBySelector(ctx, data.Namespace, selector, data.Limit, data.Continue)

		if perr != nil {
			log.Println(perr)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(pods)
	})
}

func FetchRuntimePodByNameHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data := struct {
			Name      string `json:"name"`
			Namespace string `json:"namespace"`
		}{}

		err := json.NewDecoder(r.Body).Decode(&data)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		ctx := context.Background()
		podDetails, perr := k8s.GetPodByName(ctx, data.Namespace, data.Name)
		if perr != nil {
			log.Println(perr)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(podDetails)
	})
}
