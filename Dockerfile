#部署golang环境
FROM golang:1.13.3
WORKDIR /go/cache
ADD go.mod .
ADD go.sum .
RUN go mod download
WORKDIR /go/release
ADD . .
RUN GOOS=linux CGO_ENABLED=0 go build -ldflags="-s -w" -installsuffix cgo -o gin_demo main.go
FROM scratch as prod
COPY --from=build /usr/share/zoneinfo/Asia/Shanghai /etc/localtime
COPY --from=build /go/release/gin_demo /
COPY --from=build /go/release/conf ./conf
CMD ["/gin_demo"]

#开放端口
EXPOSE 8081