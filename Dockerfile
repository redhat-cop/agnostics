ARG GO_VERSION=1.14
FROM registry.access.redhat.com/ubi8/go-toolset:latest AS builder
WORKDIR /scheduler/

USER root
RUN chown -R ${USER_UID}:0 /scheduler
USER ${USER_UID}

COPY scheduler/ ./
ENV GOOS=linux
RUN go build

FROM registry.access.redhat.com/ubi8/ubi-minimal:latest AS deploy
WORKDIR /scheduler/
USER ${USER_UID}
COPY github_known_hosts /ssh/known_hosts
env SSH_KNOWN_HOSTS /ssh/known_hosts
COPY --from=builder /scheduler/scheduler ./
CMD ["./scheduler"]
