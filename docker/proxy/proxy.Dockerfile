# FROM arm32v7/debian
FROM debian:buster

WORKDIR /root

RUN apt-get update && apt-get install net-tools procps iptables curl tcpdump -y

ADD ./ssl/EyeoCA.pem /root/ssl/EyeoCA.pem
ADD ./ssl/EyeoKey.pem /root/ssl/EyeoKey.pem
ADD ./docker/proxy/start.sh /root/start.sh
ADD ./bin/migrate_glinet /root/migrate_glinet
ADD ./bin/filter_glinet /root/filter_glinet
ADD ./tests/gambling_list.json /root/tests/gambling_list.json

ENTRYPOINT [ "/root/start.sh" ]
