# Dependency Builder
FROM golang:1.20 AS build_deps

RUN apt-get update -y
RUN apt-get install -y git bzr

WORKDIR /workspace

COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Golang Builder
FROM build_deps AS build

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o alertmanager-webhook -ldflags '-w -extldflags "-static"' .

# Use distroless base layer
FROM gcr.io/distroless/base-debian11:nonroot

COPY --from=build /workspace/alertmanager-webhook /usr/local/bin/alertmanager-webhook

ENV GIN_MODE=release
ENV PORT=8080

# Monitoring API Token Secret Name
ENV API_TOKEN_SECRET_NAME=cluster-token

# Change from default to override. If default, app will try to retrieve its running namespace
ENV API_TOKEN_SECRET_NAMESPACE=default

# Required Secret fields:
# key: Auth Token
# url: Webhook URL

ENV DEBUG=false

ENTRYPOINT ["alertmanager-webhook"]
