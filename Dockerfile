FROM golang:1.17.0-alpine3.13 AS build-env

ENV CGO_ENABLED 0
ENV GOPATH=/go
ENV PATH=$PATH:$GOPATH/bin

RUN apk add --no-cache git
ADD . /app
WORKDIR /app

RUN go build -o main

FROM debian:bullseye

RUN apt update && apt install netcat curl wget procps -y

COPY config.yaml /config.yaml
COPY form.template /form.template
COPY --from=build-env /app/main /main

RUN chmod +x /main

EXPOSE 9000

ENTRYPOINT ["/bin/bash"]
CMD ["-c", "/main"]