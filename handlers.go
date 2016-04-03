package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/docker/docker/pkg/stdcopy"
	"github.com/dustin/go-humanize"
	"github.com/google/go-github/github"
	"github.com/julienschmidt/httprouter"
	"github.com/manucorporat/sse"
)

var (
	tmpl map[string]*template.Template
)

func init() {
	funcMap := template.FuncMap{
		"lettrine": func(title string) string {
			return strings.Title(title)[:1]
		},
		"ago": func(date time.Time) string {
			return humanize.Time(date)
		},
	}

	tmpl = make(map[string]*template.Template)
	tmpl["list"] = template.Must(template.New("list").Funcs(funcMap).ParseFiles("views/list.html", "views/base.html"))
	tmpl["deployment"] = template.Must(template.New("deployment").Funcs(funcMap).ParseFiles("views/deployment.html", "views/base.html"))
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

func listHandler(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	deployments, err := listDeployment()

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	renderTemplate(w, "list", deployments)
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

	renderTemplate(w, "deployment", &deployment)
}

func cancelHandler(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {
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

	if deployment.Status != statusPending {
		http.Error(w, "Cannot cancel a finished deployment", http.StatusBadRequest)
		return
	}

	cancelDeployment(deployment)
	http.Redirect(w, req, fmt.Sprintf("/deployment/%s", deployment.ID), http.StatusFound)
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
		http.Error(w, "Cannot stream a finished deployment", http.StatusBadRequest)
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

func renderTemplate(w http.ResponseWriter, name string, data interface{}) {
	err := tmpl[name].ExecuteTemplate(w, "base", data)

	if err != nil {
		http.Error(w, fmt.Sprint(err), http.StatusInternalServerError)
	}
}
