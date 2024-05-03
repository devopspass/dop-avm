FROM ubuntu:noble

RUN apt update && apt install -y \
            ca-certificates \
            curl \
            gnupg \
            lsb-release

RUN curl -fsSL https://download.docker.com/linux/ubuntu/gpg | gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg

RUN echo \
      "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/ubuntu \
      $(lsb_release -cs) stable" | tee /etc/apt/sources.list.d/docker.list > /dev/null

RUN apt update && apt install -y python3 python3-pip docker-ce-cli openssh-client

RUN pip install --break-system-packages ansible molecule ansible-lint testinfra
RUN pip install --break-system-packages molecule-plugins[azure] molecule-plugins[containers] molecule-plugins[ec2] molecule-plugins[docker] molecule-plugins[gce] molecule-plugins[openstack]
RUN pip cache purge
RUN rm -Rf /usr/libexec/docker/cli-plugins/docker-* /usr/libexec/gcc/

ENV ANSIBLE_STDOUT_CALLBACK=yaml
WORKDIR /ansible
