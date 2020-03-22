FROM            telegraf:alpine

RUN addgroup -S tools && \
    adduser -S -G tools tools && \
    mkdir /app

COPY ./bin/process.bin /app/process
COPY ./config.yaml /app/config.yaml
RUN chown -R tools:tools /app && chmod 755 /app/process
WORKDIR /app
USER tools

EXPOSE 9280
ENTRYPOINT [ "/app/process" ]
CMD [ "--config", "/app/config.yaml" ]
