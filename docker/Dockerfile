FROM debian:jessie

RUN apt-get -y update && apt-get -y install git \
                openssh-client \
                locales \
                python-dev \
                python-pip \
                python-paramiko \
                python-yaml \
                python-jinja2 \
                python-httplib2 \
                python-six \
        --no-install-recommends && rm -r /var/lib/apt/lists/*

RUN pip install ansible

ADD ./deploy /bin/deploy
ADD ./config /root/.ssh/config

ENV LANG C.UTF-8
ENV LANGUAGE C.UTF-8
ENV LC_ALL C.UTF-8

ENTRYPOINT ["/bin/deploy"]
