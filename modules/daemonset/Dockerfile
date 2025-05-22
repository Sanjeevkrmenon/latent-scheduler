FROM alpine:3.20

RUN apk add --no-cache python3 py3-pip iputils curl
RUN curl -sSL -o /tmp/kubectl_version https://dl.k8s.io/release/stable.txt && \
    curl -sSL -o /usr/local/bin/kubectl https://dl.k8s.io/release/$(cat /tmp/kubectl_version)/bin/linux/amd64/kubectl && \
    chmod +x /usr/local/bin/kubectl && \
    rm /tmp/kubectl_version

WORKDIR /app
COPY latency_measure.py .

ENTRYPOINT ["python3", "-u", "/app/latency_measure.py"]