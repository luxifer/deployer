package main

import (
	"fmt"
	"time"
)

var (
	statusPending = "pending"
	statusSuccess = "success"
	statusError   = "failed"
)

type ExitError struct {
	ExitCode int64
	ID       int
}

type Deployment struct {
	ID       string `gorethink:"id,omitempty"`
	Owner    string
	Name     string
	JobID    int
	SSHURL   string
	HTTPURL  string
	Task     string
	Env      string
	Ref      string
	SHA      string
	Author   string
	Started  time.Time
	Finished time.Time
	ExitCode int64
	Status   string
	Logs     []byte
	User     User
	Commits  []Commit
	Files    []File
}

type User struct {
	Login     string
	AvatarURL string
	HTTPURL   string
}

type Commit struct {
	SHA     string
	HTTPURL string
	Message string
	Author  User
}

type File struct {
	Filename string
	Status   string
}

func (e ExitError) Error() string {
	return fmt.Sprintf("Deploy #%d failed with exit code \"%d\"", e.ID, e.ExitCode)
}

func (d *Deployment) PanelColor() string {
	var color string

	switch d.Status {
	case statusPending:
		color = "yellow"
	case statusSuccess:
		color = "green"
	case statusError:
		color = "red"
	default:
		color = "teal"
	}

	return color
}

func (d *Deployment) Icon() string {
	var icon string

	switch d.Status {
	case statusPending:
		icon = "refresh yellow"
	case statusSuccess:
		icon = "check circle green"
	case statusError:
		icon = "remove circle red"
	default:
		icon = "help circle teal"
	}

	return icon
}

func (f File) Icon() string {
	var icon string

	switch f.Status {
	case "added":
		icon = "plus square outline green"
	case "modified":
		icon = "write yellow"
	case "removed":
		icon = "minus square outline red"
	case "renamed":
		icon = "edit orange"
	}

	return icon
}

func (d *Deployment) LogToString() string {
	return string(d.Logs)
}

func (d *Deployment) Duration() time.Duration {
	if d.Finished.IsZero() {
		return time.Now().Sub(d.Started)
	} else {
		return d.Finished.Sub(d.Started)
	}
}

func (d *Deployment) ShortSHA() string {
	if len(d.SHA) > 7 {
		return d.SHA[:7]
	} else {
		return d.SHA
	}
}

func (c Commit) ShortSHA() string {
	return c.SHA[:7]
}

func (d *Deployment) FullName() string {
	return fmt.Sprintf("%s/%s", d.Owner, d.Name)
}
