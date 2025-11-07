
## Installation

To clone the project, create a folder and use the git clone command.

```
$ mkdir code
$ cd code
$ git clone https://github.com/h2a26/go-firstcup or git@github.com:h2a26/go-firstcup.git
$ cd go-firstcup
$ go mod vendor
```


## Create Your Own Version

If you want to create a version of the project for your own use, use the new gonew command.

```
$ go install golang.org/x/tools/cmd/gonew@latest

$ mkdir code
$ cd code
$ gonew github.com/h2a26/go-firstcup github.com/mydomain/myproject
$ cd myproject
$ go mod vendor
```


## Run Project


```
$ docker-compose -f zarf/compose/docker_compose.yaml up -d --build
$ docker-compose -f zarf/compose/docker_compose.yaml down -v
```
