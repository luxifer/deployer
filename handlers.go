package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/docker/docker/pkg/stdcopy"
	"github.com/google/go-github/github"
	"github.com/julienschmidt/httprouter"
	"github.com/manucorporat/sse"
)

var (
	funcMap template.FuncMap
)

func init() {
	funcMap = template.FuncMap{
		"date_format": func(date time.Time) string {
			return date.Format(time.Kitchen)
		},
	}
}

func eventHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	eventType := r.Header.Get("X-GitHub-Event")
	body, _ := ioutil.ReadAll(r.Body)

	switch eventType {
	case "deployment":
		var payload github.DeploymentEvent
		json.Unmarshal(body, &payload)
		go processDeploy(&payload)
	}

	w.WriteHeader(http.StatusCreated)
}

func deploymentHandler(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {
	id := ps.ByName("id")
	deployment, err := getDeployment(id)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if deployment == nil {
		fmt.Fprintf(w, "Deployment \"%s\" not found", id)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	tmpl := template.Must(template.New("deployment.html").Funcs(funcMap).ParseFiles("views/deployment.html"))
	err = tmpl.Execute(w, &deployment)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func streamHandler(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {
	id := ps.ByName("id")
	deployment, err := getDeployment(id)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if deployment == nil {
		http.Error(w, fmt.Sprintf("Deployment \"%s\" not found", id), http.StatusNotFound)
		return
	}

	if deployment.Status != statusPending {
		http.Error(w, "Cannot stream a finished event", http.StatusBadRequest)
		return
	}

	reader, err := streamDeployment(deployment)
	defer reader.Close()

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", sse.ContentType)

	rw := &StreamWriter{w, 0}
	stdcopy.StdCopy(rw, rw, reader)
}

type StreamWriter struct {
	writer http.ResponseWriter
	count  int
}

func (w *StreamWriter) Write(data []byte) (int, error) {
	var err = sse.Encode(w.writer, sse.Event{
		Id:    strconv.Itoa(w.count),
		Event: "message",
		Data:  string(data),
	})
	w.writer.(http.Flusher).Flush()
	w.count += len(data)
	return len(data), err
}
