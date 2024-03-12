FROM golang:1.21-alpine
WORKDIR /app
COPY . .
RUN go build -o products-server ./main.go
CMD ["./products-server"]