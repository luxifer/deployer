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
	ExitCode int
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
	ExitCode int
	Status   string
	Logs     []byte
}

func (e ExitError) Error() string {
	return fmt.Sprintf("Deploy #%d failed with exit code \"%d\"", e.ID, e.ExitCode)
}

func (d *Deployment) PanelColor() string {
	var color string

	switch d.Status {
	case statusPending:
		color = "warning"
	case statusSuccess:
		color = "success"
	case statusError:
		color = "danger"
	default:
		color = "info"
	}

	return color
}

func (d *Deployment) Icon() string {
	var icon string

	switch d.Status {
	case statusPending:
		icon = "octicon-sync"
	case statusSuccess:
		icon = "octicon-check"
	case statusError:
		icon = "octicon-x"
	default:
		icon = "octicon-question"
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

func (d *Deployment) FullName() string {
	return fmt.Sprintf("%s/%s", d.Owner, d.Name)
}
