FROM alpine:3.20.3

# the nonprivileged user to start entrypoint with (will be replaced with a random userid at runtime)
ENV RUNTIMEUSER=1001
ENV TZ Europe/Berlin

ENV wallboxName localhost
ENV wallboxPort 502

EXPOSE 8080

USER root

COPY ./bin/keba-rest-api.linux /keba-rest-api
RUN chmod +x keba-rest-api

USER ${RUNTIMEUSER}

ENTRYPOINT ["/keba-rest-api"]