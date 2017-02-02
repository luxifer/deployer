package main

import (
	r "gopkg.in/dancannon/gorethink.v1"
)

var listLimit = 25

func migrate() {
	r.DBCreate("deployer").Run(rc)
	r.TableCreate("deployment").Run(rc)
	r.Table("deployment").IndexCreate("Started").Run(rc)
}

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

func countDeployments() (int, error) {
	res, err := r.Table("deployment").Count().Run(rc)

	if err != nil {
		return 0, err
	}

	var total int
	res.One(&total)

	return total, nil
}

func listDeployment(page int) ([]*Deployment, error) {
	startOffset := (page - 1) * listLimit
	endOffset := startOffset + listLimit + 1
	res, err := r.Table("deployment").OrderBy(r.OrderByOpts{Index: r.Desc("Started")}).Slice(startOffset, endOffset).Run(rc)

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

func lastDeployment(owner string, name string) (*Deployment, error) {
	res, err := r.Table("deployment").OrderBy(r.OrderByOpts{Index: r.Desc("Started")}).Filter(map[string]interface{}{
		"Owner":  owner,
		"Name":   name,
		"Status": "success",
	}).Limit(1).Run(rc)

	if err != nil {
		return nil, err
	}

	defer res.Close()

	var deployment Deployment
	err = res.One(&deployment)

	if err != nil {
		return nil, err
	}

	return &deployment, nil
}
