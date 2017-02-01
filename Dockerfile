FROM       golang:latest
MAINTAINER Brian Glogower <bglogower@docker.com>
EXPOSE     9104

WORKDIR /go/src/app

COPY .  /go/src/app
RUN     go-wrapper download && go-wrapper install

ENTRYPOINT ["go-wrapper", "run"]
