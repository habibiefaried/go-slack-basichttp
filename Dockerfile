FROM debian:bullseye

RUN apt update && apt install netcat curl wget procps -y

COPY config.yaml /config.yaml
COPY form.template /form.template
COPY main /app/main /main

RUN chmod +x /main

EXPOSE 9000

ENTRYPOINT ["/bin/bash"]
CMD ["-c", "/main"]