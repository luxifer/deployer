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
		color = "default"
	}

	return color
}

func (d *Deployment) Icon() string {
	var icon string

	switch d.Status {
	case statusPending:
		icon = "spinner"
	case statusSuccess:
		icon = "check"
	case statusError:
		icon = "times"
	default:
		icon = "question"
	}

	return icon
}

func (d *Deployment) LogToString() string {
	return string(d.Logs)
}

func (d *Deployment) Duration() time.Duration {
	return d.Finished.Sub(d.Started)
}
