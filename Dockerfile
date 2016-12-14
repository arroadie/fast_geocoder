FROM alpine:3.3
ENV GEOCODER_VERSION=fast_geocoder-r0.0.1
RUN apk add --update curl && \
    rm -rf /var/cache/apk/*
RUN curl -O https://codeload.github.com/arroadie/fast_geocoder/tar.gz/$GEOCODER_VERSION
RUN tar -xzvf $GEOCODER_VERSION
ENTRYPOINT /fast_geocoder-$GEOCODER_VERSION/bin/linux/fast_geocoder --server
EXPOSE 8080
