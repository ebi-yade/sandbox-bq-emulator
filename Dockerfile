FROM golang:1.19

WORKDIR /workspace
COPY . /workspace

RUN go install github.com/cespare/reflex@v0.3.1

ENTRYPOINT ["/bin/bash", "entrypoint.sh"]
