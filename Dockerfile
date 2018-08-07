FROM alpine
RUN [ ! -e /etc/nsswitch.conf ] && echo 'hosts: files dns' > /etc/nsswitch.conf
COPY bin/gus-proxy /usr/local/bin
EXPOSE 8080
CMD [ "gus-proxy", "run", "-f", "/proxyhosts.txt" ]