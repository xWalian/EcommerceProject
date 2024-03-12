FROM golang:1.21-alpine
WORKDIR /app
COPY . .
RUN go build -o logging-server ./main.go
CMD ["./logging-server"]