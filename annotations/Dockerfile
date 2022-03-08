FROM sergenyalcin10/custom-composition-poc:3.0-alpha as builder
ENV CGO_ENABLED=0
WORKDIR /go/src/
# build the custom krm functios
RUN go build -tags netgo -ldflags '-w' -v -o /usr/local/bin/function ./annotations

FROM alpine:latest
COPY --from=builder /usr/local/bin/function /function
COPY --from=builder /usr/local/bin/kpt /usr/local/bin/kpt
COPY --from=builder /usr/bin/git /usr/local/bin/git
COPY --from=builder /usr/local/bin/kubectl /usr/local/bin/kubectl
