FROM golang:latest

WORKDIR /go/src/github.com/caitong93/kube-extra-exporter

COPY . .

ENV GOPATH /go

ARG CMD_DIR=./cmd

ARG ROOT=github.com/caitong93/kube-extra-exporter

ARG VERSION=

ARG COMMIT=

ARG TARGET=kube-extra-exporter

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64                      \
	go build -i -v -o /tmp/${TARGET}                  \
	-ldflags "-s -w -X ${ROOT}/pkg/version.Version=${VERSION}  \
	-X ${ROOT}/pkg/version.Commit=${COMMIT}                    \
	-X ${ROOT}/pkg/version.Package=${ROOT}"                    \
	${CMD_DIR}/${TARGET};

FROM debian:jessie

ARG TARGET=kube-extra-exporter

COPY --from=0 /tmp/${TARGET} /${TARGET}

RUN ln -s /${TARGET} /entrypoint

ENTRYPOINT ["/entrypoint"]
