package main

import (
	"fmt"
	"time"
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
	Logs     []byte
}

func (e ExitError) Error() string {
	return fmt.Sprintf("Deploy #%d failed with exit code \"%d\"", e.ID, e.ExitCode)
}
