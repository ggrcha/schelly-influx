FROM golang:1.10 AS BUILD

# doing dependency build separated from source build optimizes time for developer, but is not required
# install external dependencies first
# ADD go-plugins-helpers/Gopkg.toml $GOPATH/src/go-plugins-helpers/
WORKDIR $GOPATH/src/schelly-influx

ADD /main.go $GOPATH/src/schelly-influx/main.go
RUN go get -v schelly-influx

# now build source code
ADD schelly-influx $GOPATH/src/schelly-influx
RUN go get -v schelly-influx

FROM influxdb:1.6

EXPOSE 7070

# ENV RESTIC_PASSWORD ''
ENV LISTEN_PORT 7070
ENV LISTEN_IP '0.0.0.0'
ENV LOG_LEVEL 'debug'

ENV PRE_POST_TIMEOUT '7200'
ENV PRE_BACKUP_COMMAND ''
ENV POST_BACKUP_COMMAND ''

COPY --from=BUILD /go/bin/* /bin/
ADD startup.sh /

CMD [ "/startup.sh" ]
