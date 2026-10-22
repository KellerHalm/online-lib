FROM golang:1.27-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/server ./cmd/server

FROM alpine:3.20
WORKDIR /app
COPY --from=build /bin/server /bin/server
COPY docs /app/docs
EXPOSE 3000
ENTRYPOINT ["/bin/server"]
