ARG GO_VERSION=1.14
FROM golang:${GO_VERSION}-alpine AS builder
RUN apk update
WORKDIR /scheduler/
COPY scheduler/ ./
ENV GOOS=linux
RUN go build

FROM alpine:latest AS deploy
WORKDIR /scheduler/
USER ${USER_UID}
COPY github_known_hosts /ssh/known_hosts
env SSH_KNOWN_HOSTS /ssh/known_hosts
COPY --from=builder /scheduler/scheduler ./
CMD ["./scheduler"]
