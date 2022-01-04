ARG GO_VERSION=1.14
FROM registry.access.redhat.com/ubi8/go-toolset:latest AS builder
WORKDIR /agnostics/

USER root

# custom known hosts
COPY build/known_hosts /ssh/known_hosts
# Support github and gitlab
RUN ssh-keyscan github.com >> /ssh/known_hosts
RUN ssh-keyscan gitlab.com >> /ssh/known_hosts

RUN chown -R ${USER_UID}:0 /agnostics
USER ${USER_UID}

COPY ./ ./
ENV GOOS=linux
RUN go build ./cmd/scheduler

FROM registry.access.redhat.com/ubi8/ubi-minimal:latest AS deploy
RUN microdnf install -y rsync tar
WORKDIR /agnostics/
USER ${USER_UID}
env SSH_KNOWN_HOSTS /ssh/known_hosts
COPY --from=builder /agnostics/scheduler ./
COPY --from=builder /ssh /ssh
COPY ./templates/ ./templates/
CMD ["./scheduler"]
