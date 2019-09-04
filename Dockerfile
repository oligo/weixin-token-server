FROM golang:1.12-alpine
RUN apk add --no-cache wget curl git build-base
WORKDIR /wechat-token-server
COPY . .
RUN GOOS=linux go build -o token-server .

FROM alpine
RUN apk add --no-cache ca-certificates
COPY --from=0 /wechat-token-server/token-server /usr/bin
EXPOSE 8080/tcp
ENTRYPOINT ["token-server"]
