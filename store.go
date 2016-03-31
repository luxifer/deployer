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
