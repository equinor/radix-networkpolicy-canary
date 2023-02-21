FROM alpine
RUN apk add redis --no-cache
USER 1000
CMD redis-server --port ${REDIS_SERVICE_PORT} --requirepass "${REDIS_PASSWORD}"