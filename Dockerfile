FROM alpine:latest

RUN addgroup -S kube-operator && adduser -S -g kube-operator kube-operator

USER kube-operator

COPY ./bin/prometheus-config-controller . 

ENTRYPOINT ["./prometheus-config-controller"]
