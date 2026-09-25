FROM golang:1.26-alpine AS build
RUN apk add --no-cache gcc musl-dev
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=1 go build -trimpath -o /basket ./cmd/basket

FROM alpine:3.23
RUN apk add --no-cache ca-certificates tzdata && adduser -D -u 10001 basket
WORKDIR /app
COPY --from=build /basket /app/basket
COPY config.yaml /app/config.yaml
USER basket
EXPOSE 9009
ENTRYPOINT ["/app/basket"]
CMD ["-config", "/app/config.yaml"]
