# Building Production Grade Container Images

When building and running images locally for development purposes, many best practices are neglected. However, eventually the image gets deployed. For a real deployment we want to take additional steps.

## Motivation

In this scenario I *want* to deploy a simple "echo-server" written in go.

This is a good example because it is a compiled language, and it spawns a long-running process with which we can interact via HTTP. That is usually the case when deploying application in container.

The project structure looks like this

```shell
.
├── Dockerfile
├── .dockerignore
├── go.mod
├── main.go
└── README.md
```

<details>
<summary>Go Code </summary>

```go
package main

import (
 "encoding/json"
 "fmt"
 "log"
 "net/http"
)

func main() {
 http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
  reqHeadersBytes, _ := json.Marshal(r.Header)
  text := fmt.Sprintf("RemoteAddr: %s\n", r.RemoteAddr)
  text += fmt.Sprintf("Method: %s\n", r.Method)
  text += fmt.Sprintf("RequestURI: %s\n", r.RequestURI)
  text += fmt.Sprintf("Proto: %s\n", r.Proto)
  text += fmt.Sprintf("ContentLength: %x\n", r.ContentLength)
  text += fmt.Sprintf("Headers: %s\n", string(reqHeadersBytes))
  fmt.Fprint(w, text)
 })
 log.Fatal(http.ListenAndServe(":80", nil))
}
```

</details>

You can find the full [project](https://github.com/bluebrown/echo-server) on github.

## Dockerignore

With the `.dockerignore` file we can exclude content from build directory. So what's listed in this file will not get send to the daemon as build part of the  context.

```shell
$ docker build --tag bluebrown/echo-server .
sending build context to Docker daemon  6.656kB
...
```

Its a good idea to list the `Dockerfile` and the `.dockerignore` itself here in order to avoid cache invalidation of previous steps when working on these 2 files. Additionally list everything that isn't strictly required for building the image.

```shell
.git
.dockerignore
Dockerfile
README.md
```

## Multi Staging

With multi-staging it is possible to perform work in one image and then copy only what is required to a second slim image.

```dockerfile
FROM golang as builder

WORKDIR /src/
COPY . .
RUN go vet
RUN go test
RUN go build \
  -ldflags '-linkmode external -w -extldflags "-static"' \
  -o echo-server

# ---
FROM alpine as runner

CMD ["/usr/code/echo-server"]
COPY --from=builder /src/echo-server /usr/code/
```

```shell
docker build --tag bluebrown/echo-server .
docker image ls
```

If we inspect the output we can see that the final image is much smaller then the builder image. Instead of running eventually a container with the source code and the binary and a size of 868MB, we are going to run only a slim container containing the compiled binary of size 11.7MB

```shell
REPOSITORY              TAG       IMAGE ID       CREATED          SIZE
bluebrown/echo-server   latest    ee63052b3b15   9 seconds ago    11.7MB
<none>                  <none>    5eb556bfbc0f   12 seconds ago   868MB
golang                  latest    ee23292e2826   4 days ago       862MB
alpine                  latest    d4ff818577bc   12 days ago      5.6MB
```

It is not always required or useful to use multi-stage build. For example, you can compile the binary outside the image and copy in the final alpine image. However, compiling on build ensures that it is always compiled in the same environment with the same flags.

## Healthcheck

Healthchecks are a way to determine of the container is running ok. By default, the process ID (PID) is checked. So if a container has a PID it is considered healthy.

It is possible to customize the healthcheck per image. For example using curl to make a HTTP request to see if the application is responding ok.

```dockerfile
FROM alpine as runner
...
HEALTHCHECK \
  --interval=30s \
  --timeout=30s \
  --start-period=5s \
  --retries=3 \
  CMD curl --head --fail localhost || exit 1
...
```

When running starting the container now, we can also see the health status via CLI.

```shell
$ docker run --rm  --detach --name echo-server -p 80:80  bluebrown/echo-server
$ docker ps -a --format {% raw %}'{{.Names}} - {{.Status}}'{% endraw %}
echo-server - Up 5 seconds (health: starting)
```

However, if you wait 2 minutes and check again, you notice that the container is marked as unhealthy.

```shell
echo-server - Up 2 minutes (unhealthy)
```

You may also notice that even though the container is considered unhealthy, docker doesn't stop it or do anything about it. It is up to the operator or orchestration framework to handle the situation according to the health status. Docker swarm for example would restart the container now.

But why was the container *unhealthy* in the first place? If you try to curl on the published port of the the container you get actually a response.

<details>
<summary>Curl Output</summary>

```shell
$ curl localhost
RemoteAddr: 172.17.0.1:50152
Method: GET
RequestURI: /
Proto: HTTP/1.1
ContentLength: 0
Headers: {"Accept":["*/*"],"User-Agent":["curl/7.68.0"]}
```

</details>

The reason is, that the health check command is executed inside the container, and in our case we are trying to use curl even though its not installed in the container. We can see this by inspecting the status logs in the docker inspect output.

```shell
docker inspect echo-server --format \
 {%raw%}  '{{range .State.Health.Log}}{{.End}} | Exit Code: {{.ExitCode}} | {{.Output}}{{end}} {%endraw%}
```

```shell
2021-06-30 10:06:05.795671501 +0000 UTC | Exit Code: 1 | /bin/sh: curl: not found
2021-06-30 10:06:35.888445198 +0000 UTC | Exit Code: 1 | /bin/sh: curl: not found
2021-06-30 10:07:05.959345369 +0000 UTC | Exit Code: 1 | /bin/sh: curl: not found
```

We could simply install curl on build in order to fix this.

```dockerfile
FROM alpine as runner
...
RUN apk add --update curl && rm -rf /var/cache/apk/*
...
```

```shell
echo-server - Up 33 seconds (healthy)
```

## Build Arguments

Build arguments are a great way to customize the build behavior without having to modify the `Dockerfile`.

```dockerfile
FROM golang as builder
...
ARG VET_FLAGS=""
RUN go vet "$VET_FLAGS"

ARG TEST_FLAGS=""
RUN go test "$TEST_FLAGS"

ARG LD_FLAGS='-linkmode external -w -extldflags "-static"'
ARG BUILD_FLAGS=""
RUN go build -ldflags "$LD_FLAGS" -o echo-server "$BUILD_FLAGS"
...
```

That way we can pass additional flags on build. For example being extra verbose for debugging purposes.

```shell
docker build --tag bluebrown/echo-server \
  --build-arg VET_FLAGS="-x" \
  --build-arg BUILD_FLAGS="-x" \
  .
```

## Using Unprivileged User

By default, Docker gives root permission to the process that runs a container. That's no good. It's commonly solved by created an unprivileged user inside the container and run the final command as this user.

```dockerfile
FROM alpine as runner
...
ARG UID=8080
ARG USER="docker-app"
RUN adduser \
    --disabled-password \
    --gecos "" \
    --home /usr/code \
    --no-create-home \
    --uid "$UID" \
    "$USER"
...
USER $USER
```

Note, when working with local volumes you have to ensure the permission on the volume matches the `UID` and `GUID` of that user.

```shell
sudo chown -R 8080:8080 ./my-volume
docker run --volume $PWD/my-volume:/usr/data
```

## Labeling the Image

Label systems are a common way to work with dynamic configuration these days. They are a way to attach key value pairs to resources which can be used by other tools in order to operate given resource.

The Open Container Initiative has label suggestions which are commonly known and accepted. [OCI Annotations](https://github.com/opencontainers/image-spec/blob/master/annotations.md). The older deprecated version of the spec has better explanations in my opinion and since many labels were basically just renamed its can be useful too check out [label-schema.org documentation](http://label-schema.org/rc1/) as well.

The format of the oci labels is `org.opencontainers.image.<label>` where label has to be chosen from a fixed list of labels provided by OCI. If you have custom labels, you should **not** prefix them with `org.opencontainers.image` but with your own prefix e.g. `com.myorg.env="production"`.

```dockerfile
FROM alpine as runner
...
ARG VERSION="0.1.0"
ARG ENVIRONMENT="dev"
ARG BRANCH="main"
ARG COMMIT_HASH="unknown"
ARG CREATED_DATE="unknown"

LABEL org.opencontainers.image.created="${CREATED_DATE}" \
    org.opencontainers.image.url="https://github.com/my-repo"  \
    org.opencontainers.image.source="https://github.com/my-repo/Dockerfile" \
    org.opencontainers.image.version="${VERSION}-${ENVIRONMENT}" \
    org.opencontainers.image.revision="${COMMIT_HASH}" \
    org.opencontainers.image.vendor="rainbowstack" \
    org.opencontainers.image.title="echo-server" \
    org.opencontainers.image.description="go echo server" \
    org.opencontainers.image.documentation="https://github.com/my-repo/README.md" \
    org.opencontainers.image.authors="nico braun" \
    org.opencontainers.image.licenses="(BSD-1-Clause)" \
    org.opencontainers.image.ref.name="${BRANCH}" \
    dev.rainbowstack.environment="${ENVIRONMENT}"
...
```

If you now inspect the image you can find the labels.

```shell
docker inspect bluebrown/echo-server --format \
{%raw%}  '{{range $key, $val := .ContainerConfig.Labels}}{{printf "%s = %s\n" $key $val }}{{end}}'{%endraw%}
```

<details>
<summary>Output</summary>

```shell
dev.rainbowstack.environment = dev
org.opencontainers.image.authors = nico braun
org.opencontainers.image.created = unknown
org.opencontainers.image.description = go echo server
org.opencontainers.image.documentation = https://github.com/my-repo/README.md
org.opencontainers.image.licenses = (BSD-1-Clause)
org.opencontainers.image.ref.name = main
org.opencontainers.image.revision = unknown
org.opencontainers.image.source = https://github.com/my-repo/Dockerfile
org.opencontainers.image.title = echo-server
org.opencontainers.image.url = https://github.com/my-repo
org.opencontainers.image.vendor = rainbowstack
org.opencontainers.image.version = 0.1.0-dev
```

</details>

## The Final Dockerfile

The complete Dockerfile looks now like this. We are using `.dockerignore` and `multi-staging` to reduce the final image size drastically. An `HTTP Healthcheck` is implemented to see if the deployed server is actually functioning. `Arguments` and `Labels` improve the build customization and allow users and programs to get meta data about the image.

```dockerfile
FROM golang as builder

WORKDIR /src/
COPY . .

ARG VET_FLAGS=""
RUN go vet "$VET_FLAGS"

ARG TEST_FLAGS=""
RUN go test "$TEST_FLAGS"

ARG LD_FLAGS='-linkmode external -w -extldflags "-static"'
ARG BUILD_FLAGS=""
RUN go build -ldflags "$LD_FLAGS" -o echo-server "$BUILD_FLAGS"


# ---
FROM alpine as runner

CMD ["/usr/code/echo-server"]

RUN apk add --update curl && rm -rf /var/cache/apk/*

HEALTHCHECK \
  --interval=30s \
  --timeout=30s \
  --start-period=5s \
  --retries=3 \
  CMD curl --head --fail localhost || exit 1

ARG UID=8080
ARG USER="docker-app"
RUN adduser \
    --disabled-password \
    --gecos "" \
    --home /usr/code \
    --no-create-home \
    --uid "$UID" \
    "$USER"

ARG VERSION="0.1.0"
ARG ENVIRONMENT="dev"
ARG BRANCH="main"
ARG COMMIT_HASH="unknown"
ARG CREATED_DATE="unknown"

LABEL org.opencontainers.image.created="${CREATED_DATE}" \
    org.opencontainers.image.url="https://github.com/my-repo"  \
    org.opencontainers.image.source="https://github.com/my-repo/Dockerfile" \
    org.opencontainers.image.version="${VERSION}-${ENVIRONMENT}" \
    org.opencontainers.image.revision="${COMMIT_HASH}" \
    org.opencontainers.image.vendor="rainbowstack" \
    org.opencontainers.image.title="echo-server" \
    org.opencontainers.image.description="go echo server" \
    org.opencontainers.image.documentation="https://github.com/my-repo/README.md" \
    org.opencontainers.image.authors="nico braun" \
    org.opencontainers.image.licenses="(BSD-1-Clause)" \
    org.opencontainers.image.ref.name="${BRANCH}" \
    dev.rainbowstack.environment="${ENVIRONMENT}"

COPY --from=builder /src/echo-server /usr/code/
USER $USER
```

## Bonus: Content Trust

If you are using a private registry, consider opting into [content trust](https://docs.docker.com/engine/security/trust/).

Since version 1.8 docker supports code signage mechanism for published images. It is not enabled by default but can be enabled via environment flag. When enabled docker will automatically sign published images and verify on pull.

Consider running this in your current shell or adding it to your `~/.bashrc`.

```shell
export DOCKER_CONTENT_TRUST=1
```

Note, if you are planning to push images you need to take additional steps to [create your private signage key](https://docs.docker.com/engine/security/trust/#signing-images-with-docker-content-trust).
