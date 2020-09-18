FROM golang:1.14.5-alpine as build

WORKDIR /app
COPY . .
RUN go build cmd/web_server/web_server.go

# production stage
FROM alpine as production
WORKDIR /app
COPY --from=build /app/web_server /app
COPY ./.env /app

EXPOSE 80

ENTRYPOINT ["/app/web_server"]

