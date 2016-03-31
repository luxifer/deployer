package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	r "github.com/dancannon/gorethink"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/google/go-github/github"
	"github.com/julienschmidt/httprouter"
)

func eventHandler(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var payload github.DeploymentEvent
	body, _ := ioutil.ReadAll(r.Body)
	json.Unmarshal(body, &payload)
	eventType := r.Header.Get("X-GitHub-Event")

	switch eventType {
	case "deployment":
		go processDeploy(&payload)
	}

	w.WriteHeader(http.StatusCreated)
}

func deploymentHandler(w http.ResponseWriter, req *http.Request, ps httprouter.Params) {
	id := ps.ByName("id")
	res, err := r.Table("deployment").Get(id).Run(rc)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	defer res.Close()

	if res.IsNil() {
		fmt.Fprintf(w, "Deployment \"%s\" not found", id)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	var deployment Deployment
	err = res.One(&deployment)

	if err != nil {
		fmt.Fprint(w, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	stdcopy.StdCopy(w, w, bytes.NewReader(deployment.Logs))
}
