FROM registry.access.redhat.com/ubi7/ubi-minimal:latest

ENV OPERATOR=/usr/local/bin/custom-pod-autoscaler-operator \
    USER_UID=1001 \
    USER_NAME=custom-pod-autoscaler-operator

# install operator binary
COPY build/_output/bin/custom-pod-autoscaler-operator ${OPERATOR}

COPY build/bin /usr/local/bin
RUN  /usr/local/bin/user_setup

ENTRYPOINT ["/usr/local/bin/entrypoint"]

USER ${USER_UID}
