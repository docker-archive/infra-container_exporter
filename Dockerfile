FROM       alpine:latest
MAINTAINER Brian Glogower <bglogower@docker.com>
EXPOSE     9104

ENV  GOPATH /go
ENV APPPATH $GOPATH/src/github.com/docker-infra/container-exporter
COPY . $APPPATH
RUN apk add --update -t build-deps go git mercurial libc-dev gcc libgcc \
    && cd $APPPATH && go get -d && go build -o /bin/container-exporter \
    && apk del --purge build-deps && rm -rf $GOPATH

ENTRYPOINT [ "/bin/container-exporter" ]
