FROM debian:buster

WORKDIR /

USER root

RUN apt-get update && apt-get install curl iptables -y

ADD ./docker/client/request /request
ADD ./docker/client/lists /lists
ADD ./docker/client/curl-format.txt /curl-format.txt
ADD ./docker/client/resquest_time.sh /resquest_time.sh
ADD ./docker/client/iptables-rules.sh /iptables-rules.sh
ADD ./ssl/EyeoCA.pem /usr/local/share/ca-certificates/eyeo.crt
RUN update-ca-certificates --fresh

CMD "/iptables-rules.sh"
