FROM golang:alpine
WORKDIR /
COPY . .
RUN go mod download
RUN go build -o appointment-service main.go
CMD ["./appointment-service"]
