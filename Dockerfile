FROM golang:1.17

COPY . /src

WORKDIR /src

RUN go build -a -o main 

CMD ["/src/main"]

