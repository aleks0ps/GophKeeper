# docker build . -t gophkeeper:v1

FROM debian:bookworm-slim

ARG GO_URL=https://go.dev/dl/go1.23.4.linux-amd64.tar.gz

WORKDIR /tmp

RUN DEBIAN_FRONTEND=noninteractive apt-get update \
    && apt-get install -y wget curl ca-certificates make \
    && wget ${GO_URL} \
    && tar -xvzpf $(basename ${GO_URL}) -C /usr/local \
    && useradd -m gopher -s /bin/bash \
    && rm -rvf $(basename ${GO_URL})

WORKDIR /app
COPY . .
RUN chown -vR gopher:gopher /app
USER gopher
RUN echo "export PATH=\"$PATH:/usr/local/go/bin\"" > ~/.bashrc
ENV PATH="$PATH:/usr/local/go/bin"
RUN make
EXPOSE 4443
CMD /app/cmd/gophkeeper/gophkeeper
