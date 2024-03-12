FROM golang:1.21-alpine
WORKDIR /app
COPY . .
RUN go build -o payments-server ./main.go
CMD ["./payments-server"]