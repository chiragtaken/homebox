FROM resin/rpi-raspbian

MAINTAINER Chirag Tayal <chiragtayal@gmail.com>

RUN apt-get update && apt-get install -y net-tools && apt-get install -y vim

COPY src/HboxServer/HboxServer /bin

COPY hboxInit.sh /bin/
RUN chmod +x /bin/hboxInit.sh

ENTRYPOINT ["/bin/hboxInit.sh"]
