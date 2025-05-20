FROM telegraf:1.31.2-alpine

ENV ZOOKEEPER_MONITORING_HOME=/opt/zookeeper-monitoring

RUN mkdir -p ${ZOOKEEPER_MONITORING_HOME}

COPY docker/config/requirements.txt ${ZOOKEEPER_MONITORING_HOME}/requirements.txt
COPY exec-scripts/ ${ZOOKEEPER_MONITORING_HOME}/exec-scripts/
COPY docker/docker-entrypoint.sh /docker-entrypoint.sh

# Install misc. tools
RUN set -x \
    && apk add --upgrade bash curl tini python3 apk-tools openssl

RUN rm /usr/lib/python3.12/EXTERNALLY-MANAGED

RUN  python3 -m ensurepip \
    && rm -r /usr/lib/python*/ensurepip \
    && pip3 install --upgrade pip setuptools==70.0.0 \
    && pip3 install -r ${ZOOKEEPER_MONITORING_HOME}/requirements.txt \
    && chmod -R 777 ${ZOOKEEPER_MONITORING_HOME}/exec-scripts \
    && chmod +x /docker-entrypoint.sh \
    && python3 -m unittest discover -s ${ZOOKEEPER_MONITORING_HOME}/exec-scripts -p "test_*.py" \
    && rm -fr ${ZOOKEEPER_MONITORING_HOME}/exec-scripts/test_*.py \
    && rm -rf /var/cache/apk/*

# Upgrade all tools to avoid vulnerabilities
RUN set -x && apk upgrade --no-cache --available

USER 1000:0
WORKDIR ${ZOOKEEPER_MONITORING_HOME}

ENTRYPOINT ["/docker-entrypoint.sh"]
