FROM golang:1.25-alpine AS deps
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

FROM golang:1.25-alpine AS base
WORKDIR /app
COPY --from=deps /go/pkg /go/pkg
COPY . .

FROM base AS dev
RUN go install github.com/air-verse/air@v1.63.0
EXPOSE 9000
CMD ["air", "-c", ".air.toml"]

FROM base AS prod
RUN go build -o main .
EXPOSE 9000
CMD ["./main"]
