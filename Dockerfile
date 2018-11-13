FROM golang:1.10 AS BUILD

# doing dependency build separated from source build optimizes time for developer, but is not required
# install external dependencies first
# ADD go-plugins-helpers/Gopkg.toml $GOPATH/src/go-plugins-helpers/
WORKDIR $GOPATH/src/schelly-influxdb

ADD /main.go $GOPATH/src/schelly-influxdb/main.go
RUN go get -v schelly-influxdb

# now build source code
ADD schelly-influxdb $GOPATH/src/schelly-influxdb
RUN go get -v schelly-influxdb

FROM influxdb:1.6

EXPOSE 7070

# ENV RESTIC_PASSWORD ''
ENV LISTEN_PORT 7070
ENV LISTEN_IP '0.0.0.0'
ENV LOG_LEVEL 'debug'


COPY --from=BUILD /go/bin/* /bin/
ADD startup.sh /

CMD [ "/startup.sh" ]
