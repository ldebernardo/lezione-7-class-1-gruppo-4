FROM golang:1.22-alpine AS build

WORKDIR /src
COPY go.mod ./
COPY . .

RUN go test ./...
RUN go build -o /out/warehouse ./main.go

FROM alpine:3.20

WORKDIR /app
COPY --from=build /out/warehouse /app/warehouse

EXPOSE 8081
HEALTHCHECK --interval=10s --timeout=3s --retries=10 \
  CMD wget -qO- http://127.0.0.1:8081/health >/dev/null || exit 1

CMD ["/app/warehouse"]
