FROM sergenyalcin10/custom-composition-poc:4.1-alpha as builder
ENV CGO_ENABLED=0
WORKDIR /go/src/
# build the custom krm functios
RUN go build -tags netgo -ldflags '-w' -v -o /usr/local/bin/function ./chroot/label

FROM alpine:latest
COPY --from=builder /usr/local/bin/function /usr/local/bin/function

ENTRYPOINT ["/usr/local/bin/function"]