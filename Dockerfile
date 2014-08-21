FROM       ubuntu
MAINTAINER Johannes 'fish' Ziemke <github@freigeist.org> (@discordianfish)

RUN        apt-get update && apt-get install -yq curl git mercurial gcc
RUN        curl -s https://go.googlecode.com/files/go1.2.linux-amd64.tar.gz | tar -C /usr/local -xzf -
ENV        PATH    /usr/local/go/bin:$PATH
ENV        GOPATH  /go

ADD        . /go/src/github.com/discordianfish/container_exporter
RUN        cd /go/src/github.com/discordianfish/container_exporter && go get -d && go build

ENTRYPOINT [ "/go/src/github.com/discordianfish/container_exporter/container_exporter" ]
EXPOSE     8080
