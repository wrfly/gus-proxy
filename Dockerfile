FROM wrfly/glide AS build
COPY . /go
RUN glide i &&\
    make test &&\
    make build

FROM alpine
COPY --from=build /go/gus-proxy /usr/local/bin
EXPOSE 8080
CMD [ "gus-proxy", "help" ]