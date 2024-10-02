FROM curlimages/curl

WORKDIR /

ADD ./docker/client/request /request
ADD ./docker/client/lists /lists
ADD ./docker/client/curl-format.txt /curl-format.txt
ADD ./docker/client/resquest_time.sh /resquest_time.sh

