# Stage 1: Download kubectl in a temporary container so we don't keep curl in the final image
FROM alpine:3.20 AS kubectl-downloader

RUN apk add --no-cache curl
RUN KVERSION=$(curl -sSL https://dl.k8s.io/release/stable.txt) && \
    curl -sSL -o /kubectl "https://dl.k8s.io/release/${KVERSION}/bin/linux/amd64/kubectl" && \
    chmod +x /kubectl

# Stage 2: Final minimal image
FROM python:3.12.3-alpine3.20

RUN apk add --no-cache iputils

# Copy kubectl from build stage, not curl
COPY --from=kubectl-downloader /kubectl /usr/local/bin/kubectl

WORKDIR /app
COPY latency_measure.py .

ENTRYPOINT ["python3", "-u", "/app/latency_measure.py"]