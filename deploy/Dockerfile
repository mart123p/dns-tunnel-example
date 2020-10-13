FROM alpine:3.10

RUN echo "===> Installing sudo to emulate normal OS behavior..."  && \
    apk --update add sudo                                         && \
    \
    \
    echo "===> Adding Python runtime..."  && \
    apk --update add python3 py3-pip openssl ca-certificates    && \
    apk --update add --virtual build-dependencies \
                python3-dev libffi-dev openssl-dev build-base  && \
    pip3 install --upgrade pip cffi                            && \
    \
    \
    echo "===> Installing Ansible..."  && \
    pip3 install ansible==2.9.* && \
    \
    \
    echo "===> Installing handy tools (not absolutely required)..."  && \
    pip3 install --upgrade pycrypto pywinrm ansible[azure]==2.9.* && \
    apk --update add sshpass openssh-client rsync curl && \
    \
    \
    python3 -m venv /azcli && \
    source /azcli/bin/activate && \
    cd /azcli/ && \
    pip3 install azure-cli && \
    ln -s /azcli/bin/az /usr/bin/az && \
    deactivate && \
    \
    \
    echo "===> Removing package list..."  && \
    apk del build-dependencies            && \
    rm -rf /var/cache/apk/*               && \
    \
    \
    echo "===> Adding hosts for convenience..."  && \
    mkdir -p /etc/ansible                        && \
    echo 'localhost' > /etc/ansible/hosts

WORKDIR /data
CMD [ "ansible-playbook", "--version" ]