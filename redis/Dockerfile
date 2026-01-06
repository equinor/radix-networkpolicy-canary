FROM alpine:3.23
RUN apk add redis --no-cache
USER 1000
ENTRYPOINT ["/bin/sh", "-c"]
CMD ["redis-server --save \"\" --appendonly no --port ${REDIS_PORT:-6379} --requirepass \"${REDIS_PASSWORD}\""]