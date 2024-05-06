FROM golang:1.22 AS builder

COPY ./ ./

WORKDIR /gp/src/app

COPY . . 

RUN go build -o main .

EXPOSE 8080

CMD ["./main"]