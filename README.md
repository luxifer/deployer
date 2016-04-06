# Deployer

Small server that listens to github deployment event and trigger related ansible playbook inside a docker container.

## Configuration

* `DEPLOYER_HIPCHAT_TOKEN` _(required)_: Hipchat API token
* `DEPLOYER_HIPCHAT_ROOM` _(required)_: Hipchat room to notify deployment
* `DEPLOYER_RETHINK_HOST` _(required)_: RethinkDB (`<host>:<port>`)
* `DEPLOYER_HOST` _(required)_: Host (`https?://<domain>`)
* `DEPLOYER_SSHKEY_PATH` _(required)_: SSH key path allowed to clone the repository and access the target deployment hosts
* `DEPLOYER_GITHUB_TOKEN` _(optional)_: Github API token (required for private repos)
* `DEPLOYER_DOCKER_HOST` _(optional)_: Docker host (default to: unix:///var/run/docker.sock)
* `DEPLOYER_BIND` _(optional)_: IP to bind to (default: 0.0.0.0)
* `PORT` _(optional)_: Port to bind to (default: 4567)

## Run

Dependencies:

* Docker
* RethinkDB

### Runner

See https://github.com/Xotelia/deployer-ansible

```bash
$ docker pull xotelia/deployer-ansible
```

### Local

If you want to run the deployer server directly on your host:

```bash
$ go build
$ ./deployer
```

### Docker

If you want to run the deployer server inside a container:

```bash
$ docker build -t xotelia/deployer .
$ docker run -d -v /var/run/docker.sock:/var/run/docker.sock --name deployer [OPTIONS] xotelia/deployer
```

The server does not need to be run as a privileged container because it will not create child container but sibling. That's why we have to share the docker socket (if only the target docker server listen on a socket).

## Target

Create a [webhook](https://developer.github.com/webhooks/creating/) on the github target you want to deploy who points to `DEPLOYER_HOST/event_handler`.
Create a shell script called `deployer` at the root of the repository. In this script you will have the following env var available:

* `DEPLOYER_ID`: Github deployment ID
* `DEPLOYER_REPO`: SSH URL of the repository
* `DEPLOYER_TASK`: Task to run (default: deploy)
* `DEPLOYER_ENV`: Environment to deploy (default: production)
* `DEPLOYER_REF`: Ref to deploy (default: master)

In this shell script you may only call ansible modules and/or playbooks.
