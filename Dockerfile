FROM ubuntu
MAINTAINER whyrusleeping (why@ipfs.io)

ADD bin/* /bin/

ADD scripts/start_ipfs.sh /bin/start_ipfs
ADD scripts/addfile /bin/addfile

RUN mkdir -p /home/ipfs
WORKDIR /home/ipfs
ENV IPFS_PATH /home/ipfs/.ipfs

ENTRYPOINT ["/bin/start_ipfs"]
