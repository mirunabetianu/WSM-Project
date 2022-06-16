FROM golang:1.18

WORKDIR /usr/src/app

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .

RUN go build .

EXPOSE 3002

CMD ["go", "run", "main.go"]