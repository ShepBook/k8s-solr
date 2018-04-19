FROM        solr:7.3
MAINTAINER  Jared Koumentis <jared@koumentis.com>

ENV SOLR_HEAP="256m"
ENV SOLR_CORE="core1"

ADD conf /opt/search_api

ADD scripts/heap.sh /docker-entrypoint-initdb.d/set-heap.sh
ADD scripts/core.sh /docker-entrypoint-initdb.d/set-core.sh

# Where we store our persistent data so we can swap it for mounted storage.
RUN mkdir /opt/solr/data
