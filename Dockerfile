FROM wrfly/glide AS build
RUN mkdir -p /go/src/github.com/wrfly/gus-proxy
COPY . /go/src/github.com/wrfly/gus-proxy
RUN cd /go/src/github.com/wrfly/gus-proxy && \
    glide i &&\
    make test &&\
    make build &&\
    mv * /go

FROM alpine
COPY --from=build /etc/nsswitch.conf /etc/nsswitch.conf 
COPY --from=build /go/gus-proxy /usr/local/bin
EXPOSE 8080
CMD [ "gus-proxy", "help" ]