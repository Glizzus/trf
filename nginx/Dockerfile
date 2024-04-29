FROM nginx:1.25.5-alpine

LABEL org.opencontainers.image.created="${BUILD_TIME}"
LABEL org.opencontainers.image.source="https://github.com/glizzus/trf"
LABEL org.opencontainers.image.title="trf-propogator"
LABEL org.opencontainers.image.description="The web frontend for trf"
LABEL org.opencontainers.image.revision="${GIT_HASH}"

RUN rm /etc/nginx/conf.d/default.conf
COPY nginx.conf.template /etc/nginx/templates/nginx.conf.template

COPY static/ /usr/share/nginx/static
