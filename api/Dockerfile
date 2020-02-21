FROM golang:latest

WORKDIR /app
COPY . .

RUN go get -d -v ./...
RUN go build -o photoview .

CMD ["/app/photoview"]