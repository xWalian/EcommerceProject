FROM golang:1.21-alpine
WORKDIR /app
COPY . .
RUN go build -o auth-server ./main.go
CMD ["./auth-server"]