ARG DOCKER_BUILDER_IMAGE=golang:1.19
ARG DOCKER_BASE_IMAGE=alpine:3.13

FROM ${DOCKER_BUILDER_IMAGE} AS builder

WORKDIR /backup-restore-tool

# Download dependencies
COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .

RUN make build

# Production environment
FROM ${DOCKER_BASE_IMAGE}

RUN apk --update add postgresql-client

ENV BRT_BIN=/backup-restore/backup-restore-tool \
  USER_UID=10001 \
  USER_NAME=backup-restore

WORKDIR /backup-restore

COPY --from=builder /backup-restore-tool/build/_output/bin/backup-restore-tool /backup-restore
COPY --from=builder /backup-restore-tool/build/bin /usr/local/bin

RUN  /usr/local/bin/user_setup

ENTRYPOINT ["/usr/local/bin/entrypoint"]

USER ${USER_UID}
