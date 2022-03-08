FROM golang:1.16-alpine as builder
ENV CGO_ENABLED=0
WORKDIR /go/src/
COPY . .
# install kpt
ADD https://github.com/GoogleContainerTools/kpt/releases/download/v1.0.0-beta.13/kpt_linux_amd64 /usr/local/bin/kpt
RUN chmod +x /usr/local/bin/kpt
# kpt needs git as a requirement
RUN apk add git
# install kubectl to debug
ADD https://dl.k8s.io/release/v1.23.0/bin/linux/amd64/kubectl /usr/local/bin/kubectl
RUN chmod +x /usr/local/bin/kubectl
