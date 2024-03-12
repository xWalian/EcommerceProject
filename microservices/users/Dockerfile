FROM golang:1.21-alpine
WORKDIR /app
COPY . .
RUN go build -o users-server ./main.go
CMD ["./users-server"]