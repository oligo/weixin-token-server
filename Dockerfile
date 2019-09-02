FROM golang:1.12-alpine
WORKDIR /wechat-token-server
COPY . .
RUN GOOS=linux go build -o token-server .

FROM alpine
COPY --from=0 /wechat-token-server/token-server /usr/bin
ENTRYPOINT ["token-server"]
