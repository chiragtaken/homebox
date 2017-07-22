FROM resin/rpi-raspbian

MAINTAINER Chirag Tayal <chiragtayal@gmail.com>

RUN apt-get update && apt-get install -y net-tools && apt-get install -y vim

COPY AfwServer /bin

COPY afwInit.sh /bin/
RUN chmod +x /bin/afwInit.sh

ENTRYPOINT ["/bin/afwInit.sh"]
