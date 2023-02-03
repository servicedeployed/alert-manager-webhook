# Dependency Builder
FROM golang:1.19 AS build_deps

RUN apt-get update -y
RUN apt-get install -y git bzr

WORKDIR /workspace
ENV GO111MODULE=on

COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Builder
FROM build_deps AS build

COPY . .

RUN CGO_ENABLED=0 go build -o webhook -ldflags '-w -extldflags "-static"' .

# Grab the debian slim root runner runner
FROM gcr.io/service-deployed-beta/developer/runners/debian_slim_root

COPY --from=build /workspace/alertmanager-webhook /usr/local/bin/alertmanager-webhook

ENTRYPOINT ["alertmanager-webhook"]
