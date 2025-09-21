FROM golang:1.23.1 AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o books-api ./cmd/api

FROM gcr.io/distroless/base-debian12
WORKDIR /app
COPY --from=build /app/books-api /app/books-api
ENV PORT=8080
EXPOSE 8080
CMD ["/app/books-api"]
