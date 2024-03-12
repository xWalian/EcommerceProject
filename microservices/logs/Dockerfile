FROM golang:1.21-alpine
WORKDIR /app
COPY . .
RUN go build -o logs-server ./main.go
CMD ["./logs-server"]