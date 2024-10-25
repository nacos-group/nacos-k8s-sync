FROM golang:1.22.2 AS build


WORKDIR /src
COPY . .
RUN go env -w GOPROXY=https://goproxy.cn,direct
RUN STATIC=0 GOOS=linux GOARCH=amd64 LDFLAGS='-extldflags -static -s -w' go build -o main ./cmd/nacos-k8s-sync

FROM ubuntu:24.04
WORKDIR /
COPY --from=build /src/main /main
CMD ["./main"]