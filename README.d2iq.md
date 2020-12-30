
This is a D2iQ specific fork that adds additional patches on top of `dex` vanilla `v2.27.0` version.

This repo contains:

- D2iQ specific theme in `./web/themes/d2iq` directory

## Releasing a new version

D2iQ uses docker container that is stored under D2iQ's Docker Hub repository - https://hub.docker.com/r/mesosphere/dex

To publish a new version you must have valid credentials and permissions to write to the Docker repository.

1. Tag a new version

```sh
git tag "$(git describe --match v2.27.0 --abbrev=4)-d2iq"
```

2. Build a container for Mesosphere repo

```sh
make docker-image DOCKER_REPO=mesosphere/dex
```

An example of successful build:

```
....
Successfully built 60569050bd8b
Successfully tagged mesosphere/dex:v2.27.0-2-gb52a-d2iq
```

The container tag must match the name of newly created tag.


3. Push the container to the mesosphere repository:

```sh
docker push mesosphere/dex:$(git describe --tags)
```
