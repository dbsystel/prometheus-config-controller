FROM alpine:latest

RUN apk update \
    && apk add --no-cache curl \
                          ca-certificates \
                          tzdata \
    && update-ca-certificates

RUN addgroup -S kube-operator && adduser -S -g kube-operator kube-operator
USER kube-operator

COPY prometheus-config-controller /bin/prometheus-config-controller

ENTRYPOINT ["/bin/prometheus-config-controller"]

