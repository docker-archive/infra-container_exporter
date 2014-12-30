FROM       golang:latest
MAINTAINER Johannes 'fish' Ziemke <github@freigeist.org> (@discordianfish)
ENV APP    /gopath/src/github.com/docker-infra/container-exporter
ENV GOPATH /gopath:/$APP/_vendor
WORKDIR    $APP
ADD        . /gopath/src/github.com/docker-infra/container-exporter
RUN        go get -d && go build
ENTRYPOINT [ "./container-exporter" ]
EXPOSE     8080
