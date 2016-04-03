package main

import (
	r "github.com/dancannon/gorethink"
)

func updateDeployment(d *Deployment) {
	res, err := r.Table("deployment").Get(d.ID).Run(rc)

	if err != nil {
		return
	}

	if res.IsNil() {
		w, _ := r.Table("deployment").Insert(d).RunWrite(rc)
		d.ID = w.GeneratedKeys[0]
	} else {
		r.Table("deployment").Get(d.ID).Update(d).RunWrite(rc)
	}
}

func getDeployment(id string) (*Deployment, error) {
	res, err := r.Table("deployment").Get(id).Run(rc)

	if err != nil {
		return nil, err
	}

	defer res.Close()

	if res.IsNil() {
		return nil, nil
	}

	var deployment Deployment
	err = res.One(&deployment)

	if err != nil {
		return nil, err
	}

	return &deployment, nil
}

func listDeployment() ([]*Deployment, error) {
	res, err := r.Table("deployment").Run(rc)

	if err != nil {
		return nil, err
	}

	defer res.Close()

	var deployments []*Deployment

	err = res.All(&deployments)

	if err != nil {
		return nil, err
	}

	return deployments, nil
}
