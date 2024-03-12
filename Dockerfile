ARG GO_VERSION=1
FROM golang:${GO_VERSION}-alpine as builder

WORKDIR /usr/src/app
COPY ./backend/go.mod ./backend/go.sum ./
RUN go mod download && go mod verify
COPY backend .
RUN go build -v -o /run-app ./cmd/


FROM alpine:latest

COPY --from=builder /run-app /usr/local/bin/
CMD ["run-app"]
