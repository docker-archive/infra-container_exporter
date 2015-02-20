FROM       golang:latest
MAINTAINER Johannes 'fish' Ziemke <github@freigeist.org> (@discordianfish)
ENV APP    /go/src/github.com/docker-infra/container_exporter
ENV GOPATH /go:/$APP/_vendor
WORKDIR    $APP
ADD        . /go/src/github.com/docker-infra/container_exporter
RUN        go get -u -d && go build
ENTRYPOINT [ "./container_exporter" ]
EXPOSE     9104
