FROM ubuntu
LABEL org.opencontainers.image.authors="Effective Security <denis@effectivesecurity.pt>" \
      org.opencontainers.image.url="https://github.com/effective-security/protoc-gen-go" \
      org.opencontainers.image.source="https://github.com/effective-security/protoc-gen-go" \
      org.opencontainers.image.documentation="https://github.com/effective-security/protoc-gen-go" \
      org.opencontainers.image.vendor="Effective Security" \
      org.opencontainers.image.description="Protobuf gen tool"

ENV PATH=/tools:/usr/local/protoc/bin:/usr/local/go/bin:$PATH

RUN mkdir -p /tools \
      /tmp \
      /tmp/protoc3 \
      /third_party \
      /usr/local/include/google/protobuf

COPY ./bin/* ./genproto.sh /tools/
COPY ./proto/ /third_party/

RUN apt update && apt install -y curl zip
RUN curl -L https://github.com/google/protobuf/releases/download/v22.3/protoc-22.3-linux-x86_64.zip -o /tmp/protoc.zip
RUN unzip /tmp/protoc.zip -d /tmp/protoc3
RUN mv /tmp/protoc3/bin/* /usr/local/bin/
RUN cp -r /tmp/protoc3/include/ /usr/local/

RUN curl -L https://go.dev/dl/go1.20.3.linux-amd64.tar.gz -o /tmp/go.tar.gz
RUN tar -C /usr/local -xzf /tmp/go.tar.gz

RUN curl -L https://github.com/grpc/grpc-web/releases/download/1.4.2/protoc-gen-grpc-web-1.4.2-linux-x86_64 -o /usr/local/bin/protoc-gen-grpc-web
RUN chmod a+x /usr/local/bin/protoc-gen-grpc-web

#RUN ls -alR /tools/

RUN rm -rf /tmp

VOLUME ["/dirs"]

# Define default command.
ENTRYPOINT ["/tools/genproto.sh"]
