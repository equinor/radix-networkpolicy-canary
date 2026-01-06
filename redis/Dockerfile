FROM alpine:3.23
RUN apk add redis --no-cache
USER 1000
ENTRYPOINT ["/bin/sh", "-c"]
# REDIS_SERVICE_PORT is automatically set by Kubernetes
CMD ["redis-server --save \"\" --appendonly no --port ${REDIS_SERVICE_PORT:-6379} --requirepass \"${REDIS_PASSWORD}\""]