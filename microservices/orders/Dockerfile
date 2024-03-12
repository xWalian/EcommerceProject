FROM golang:1.21-alpine
WORKDIR /app
COPY . .
RUN go build -o orders-server ./main.go
CMD ["./orders-server"]