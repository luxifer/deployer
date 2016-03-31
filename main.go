package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/andybons/hipchat"
	r "github.com/dancannon/gorethink"
	"github.com/docker/engine-api/client"
	"github.com/google/go-github/github"
	"github.com/gorilla/handlers"
	"github.com/julienschmidt/httprouter"
	"golang.org/x/oauth2"
)

var (
	gc *github.Client
	hc hipchat.Client
	dc *client.Client
	rc *r.Session

	version = "0.0.1"
)

func main() {
	hipchatToken := os.Getenv("DEPLOYER_HIPCHAT_TOKEN")

	if hipchatToken == "" {
		log.Fatal("DEPLOYER_HIPCHAT_TOKEN is required")
	}

	dockerHost := os.Getenv("DEPLOYER_DOCKER_HOST")

	if dockerHost == "" {
		dockerHost = "unix:///var/run/docker.sock"
	}

	rethinkHost := os.Getenv("DEPLOYER_RETHINK_HOST")

	if rethinkHost == "" {
		log.Fatal("DEPLOYER_RETHINK_HOST is required")
	}

	githubToken := os.Getenv("DEPLOYER_GITHUB_TOKEN")
	bind := os.Getenv("DEPLOYER_BIND")
	port := os.Getenv("PORT")

	if port == "" {
		port = "4567"
	}

	if githubToken != "" {
		ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: githubToken})
		tc := oauth2.NewClient(oauth2.NoContext, ts)
		gc = github.NewClient(tc)
	} else {
		gc = github.NewClient(nil)
	}

	hc = hipchat.NewClient(hipchatToken)
	defaultHeaders := map[string]string{"User-Agent": fmt.Sprintf("deployer-%s", version)}
	dc, _ = client.NewClient(dockerHost, "v1.22", nil, defaultHeaders)
	rc, _ = r.Connect(r.ConnectOpts{
		Address:  rethinkHost,
		Database: "deployer",
	})

	mux := httprouter.New()
	mux.POST("/event_handler", eventHandler)
	mux.GET("/deployment/:id", deploymentHandler)

	log.Printf("Listening on %s:%s...", bind, port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf("%s:%s", bind, port), handlers.LoggingHandler(os.Stdout, mux)))
}
