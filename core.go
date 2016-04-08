package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/andybons/hipchat"
	"github.com/docker/engine-api/types"
	"github.com/docker/engine-api/types/container"
	"github.com/docker/engine-api/types/network"
	"github.com/google/go-github/github"
	"golang.org/x/net/context"
)

var (
	host       string
	room       string
	sshKeyPath string

	githubStatusPending = "pending"
	githubStatusSuccess = "success"
	githubStatusError   = "failure"

	defaultImage = "xotelia/deployer-ansible"
)

func init() {
	host = os.Getenv("DEPLOYER_HOST")

	if host == "" {
		log.Fatal("DEPLOYER_HOST is required")
		os.Exit(1)
	}

	room = os.Getenv("DEPLOYER_HIPCHAT_ROOM")

	if room == "" {
		log.Fatal("DEPLOYER_HIPCHAT_ROOM is required")
		os.Exit(1)
	}

	sshKeyPath = os.Getenv("DEPLOYER_SSHKEY_PATH")

	if sshKeyPath == "" {
		log.Fatal("DEPLOYER_SSHKEY_PATH is required")
		os.Exit(1)
	}
}

func processDeploy(payload *github.DeploymentEvent) {
	deployment := Deployment{
		Owner:   *payload.Repo.Owner.Login,
		Name:    *payload.Repo.Name,
		JobID:   *payload.Deployment.ID,
		SSHURL:  *payload.Repo.SSHURL,
		HTTPURL: *payload.Repo.HTMLURL,
		Task:    *payload.Deployment.Task,
		Env:     *payload.Deployment.Environment,
		Ref:     *payload.Deployment.Ref,
		SHA:     *payload.Deployment.SHA,
		Author:  *payload.Sender.Login,
		Started: time.Now(),
		Status:  statusPending,
	}

	updateDeployment(&deployment)

	log.Printf("Deploy #%d started at %s", deployment.JobID, deployment.Started.Format(time.UnixDate))

	createDeploymentStatus(&deployment, githubStatusPending)
	notifyDeploymentStatus(&deployment, githubStatusPending, hipchat.ColorYellow)
	err := launchDeployment(&deployment)

	log.Printf("Deploy #%d finished at %s", deployment.JobID, deployment.Finished.Format(time.UnixDate))

	if err != nil {
		log.Print(err)
		deployment.Status = statusError
		createDeploymentStatus(&deployment, githubStatusError)
		notifyDeploymentStatus(&deployment, githubStatusError, hipchat.ColorRed)
	} else {
		deployment.Status = statusSuccess
		createDeploymentStatus(&deployment, githubStatusSuccess)
		notifyDeploymentStatus(&deployment, githubStatusSuccess, hipchat.ColorGreen)
	}

	updateDeployment(&deployment)

	log.Printf("Deploy #%d last %s", deployment.JobID, deployment.Finished.Sub(deployment.Started))
}

func createDeploymentStatus(d *Deployment, state string) {
	targetURL := fmt.Sprintf("%s/deployment/%s", host, d.ID)
	status := github.DeploymentStatusRequest{
		State:     &state,
		TargetURL: &targetURL,
	}
	gc.Repositories.CreateDeploymentStatus(d.Owner, d.Name, d.JobID, &status)
}

func notifyDeploymentStatus(d *Deployment, state string, color string) {
	message := fmt.Sprintf("%s: deployment <a href=\"%s\">#%d</a> in <a href=\"%s\">%s/%s</a> (%s)",
		strings.Title(state),
		fmt.Sprintf("%s/deployment/%s", host, d.ID),
		d.JobID,
		d.HTTPURL,
		d.Owner,
		d.Name,
		d.Ref)
	req := hipchat.MessageRequest{
		RoomId:        room,
		From:          "Deployer",
		Message:       message,
		Color:         color,
		MessageFormat: hipchat.FormatHTML,
		Notify:        true,
	}
	hc.PostMessage(req)
}

func streamDeployment(d *Deployment) (io.ReadCloser, error) {
	ctx := context.Background()
	name := fmt.Sprintf("deployer_%d", d.JobID)

	logOpts := types.ContainerLogsOptions{
		ContainerID: name,
		ShowStdout:  true,
		ShowStderr:  true,
		Follow:      true,
	}

	return dc.ContainerLogs(ctx, logOpts)
}

func cancelDeployment(d *Deployment) error {
	ctx := context.Background()
	name := fmt.Sprintf("deployer_%d", d.JobID)

	return dc.ContainerStop(ctx, name, 0)
}

func launchDeployment(d *Deployment) error {
	ctx := context.Background()
	name := fmt.Sprintf("deployer_%d", d.JobID)

	_, err := dc.ImagePull(ctx, types.ImagePullOptions{ImageID: defaultImage}, nil)

	if err != nil {
		return err
	}

	config := container.Config{
		Image:     defaultImage,
		Tty:       true,
		OpenStdin: true,
		Env: []string{
			fmt.Sprintf("DEPLOYER_ID=%d", d.JobID),
			fmt.Sprintf("DEPLOYER_REPO=%s", d.SSHURL),
			fmt.Sprintf("DEPLOYER_TASK=%s", d.Task),
			fmt.Sprintf("DEPLOYER_ENV=%s", d.Env),
			fmt.Sprintf("DEPLOYER_REF=%s", d.Ref),
			fmt.Sprintf("DEPLOYER_SHA=%s", d.SHA),
		},
	}
	hostConfig := container.HostConfig{
		Binds: []string{fmt.Sprintf("%s:/root/.ssh/id_rsa", sshKeyPath)},
	}
	c, err := dc.ContainerCreate(ctx, &config, &hostConfig, &network.NetworkingConfig{}, name)

	defer func() {
		d.Finished = time.Now()
		dc.ContainerKill(ctx, c.ID, "9")
		removeOptions := types.ContainerRemoveOptions{
			ContainerID:   c.ID,
			RemoveLinks:   true,
			RemoveVolumes: true,
			Force:         false,
		}
		dc.ContainerRemove(ctx, removeOptions)
	}()

	if err != nil {
		return err
	}

	err = dc.ContainerStart(ctx, c.ID)

	if err != nil {
		return err
	}

	exitCode, err := dc.ContainerWait(ctx, c.ID)

	if err != nil {
		return err
	}

	d.ExitCode = exitCode

	logOpts := types.ContainerLogsOptions{
		ContainerID: c.ID,
		ShowStdout:  true,
		ShowStderr:  true,
		Follow:      true,
	}

	reader, err := dc.ContainerLogs(ctx, logOpts)

	if err != nil {
		return err
	} else {
		var b bytes.Buffer
		w := bufio.NewWriter(&b)
		io.Copy(w, reader)
		w.Flush()
		d.Logs = b.Bytes()
	}

	if exitCode != 0 {
		return ExitError{ExitCode: exitCode, ID: d.JobID}
	}

	return nil
}
