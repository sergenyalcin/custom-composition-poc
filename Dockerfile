FROM golang:1.16-alpine as builder
ENV CGO_ENABLED=0
WORKDIR /go/src/
COPY . .
# build the custom krm functios
RUN go build -tags netgo -ldflags '-w' -v -o /usr/local/bin/set-annotations ./annotations
RUN go build -tags netgo -ldflags '-w' -v -o /usr/local/bin/set-labels ./labels
# install kpt
ADD https://github.com/GoogleContainerTools/kpt/releases/download/v1.0.0-beta.13/kpt_linux_amd64 /usr/local/bin/kpt
RUN chmod +x /usr/local/bin/kpt
# kpt needs git as a requirement
RUN apk add git
# install kubectl to debug
ADD https://dl.k8s.io/release/v1.23.0/bin/linux/amd64/kubectl /usr/local/bin/kubectl
RUN chmod +x /usr/local/bin/kubectl

FROM alpine:latest
COPY --from=builder /usr/local/bin/set-annotations /set-annotations
COPY --from=builder /usr/local/bin/set-labels /set-labels
COPY --from=builder /usr/local/bin/kpt /usr/local/bin/kpt
COPY --from=builder /usr/bin/git /usr/local/bin/git
COPY --from=builder /usr/local/bin/kubectl /usr/local/bin/kubectl
CMD ["function"]
