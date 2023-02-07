# Dependency Builder
FROM golang:1.19 AS build_deps

RUN apt-get update -y
RUN apt-get install -y git bzr

WORKDIR /workspace
ENV GO111MODULE=on

COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Golang Builder
FROM build_deps AS build

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o alertmanager-webhook -ldflags '-w -extldflags "-static"' .

# Use scratch base layer
FROM gcr.io/distroless/base-debian11:nonroot

COPY --from=build /workspace/alertmanager-webhook /usr/local/bin/alertmanager-webhook

ENV GIN_MODE=release
ENV PORT=8080

ENTRYPOINT ["alertmanager-webhook"]
