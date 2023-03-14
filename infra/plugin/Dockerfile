FROM golang:alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o main .

FROM gcr.io/distroless/base
COPY --from=builder /app/main /app/
ENTRYPOINT ["/app/main"]