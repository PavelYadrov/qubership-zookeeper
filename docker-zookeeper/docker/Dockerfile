FROM --platform=$BUILDPLATFORM golang:1.22.4-alpine3.20 as builder
ARG BUILDPLATFORM
ARG TARGETOS
ARG TARGETARCH

WORKDIR /workspace
# Copy the Go Modules manifests
COPY go.mod go.mod
COPY go.sum go.sum
# cache deps before building and copying source so that we don't need to re-download as much
# and so that source changes don't invalidate our downloaded layer
RUN go mod download

# Copy the go source
COPY main.go main.go

# Tests
#RUN CGO_ENABLED=0 go test -v ./...
# Build
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} GO111MODULE=on go build -a -o zookeeper-assistant main.go

FROM eclipse-temurin:22.0.1_8-jdk-alpine

ENV ZOOKEEPER_HOME=/opt/zookeeper \
    ZOOKEEPER_OPT=/var/opt/zookeeper \
    ZOOKEEPER_DATA=/var/opt/zookeeper/data \
    ZOOKEEPER_BACKUP_SOURCE_DIR=/var/opt/zookeeper/data/version-2 \
    ZOOKEEPER_BACKUP_DESTINATION_DIR=/opt/zookeeper/backup-storage/tmp \
    ZOOKEEPER_RECOVERY_DIR=/opt/zookeeper/backup-storage/recover \
    ZOOKEEPER_CLIENT_PORT=2181 \
    ZOOKEEPER_FOLLOWERS_PORT=2888 \
    ZOOKEEPER_ELECTION_PORT=3888 \
    ZOOKEEPER_BACKUP_PORT=8081

WORKDIR /
COPY --from=builder --chown=${USER_UID} /workspace/zookeeper-assistant .

# Install misc tools
RUN set -x \
    && apk add --update --no-cache \
        bash \
        shadow \
        netcat-openbsd \
        jq \
        tini \
        rsync \
        curl \
        grep \
        ca-certificates \
        wget \
        jattach \
        iputils \
        findutils \
        util-linux \
        apk-tools \
        ncdu \
        coreutils \
        openssl \
        procps \
        zip \
    && rm -rf /var/cache/apk/*

# Add unpriviliged user
RUN set -x \
    && groupadd -r zookeeper --gid=1000 \
    && useradd -s /bin/bash -r -g zookeeper --uid=1000 zookeeper \
    && usermod -a -G 0 zookeeper

# Reminder: we don't upgrade it to 3.9.x until Clickhouse support
ARG DISTRO_NAME=apache-zookeeper-3.8.4-bin
ARG DISTRO_VERSION_NAME=zookeeper-3.8.4
# Download Apache Zookeeper and install
RUN set -x \
    && export DISTR_DIR="$(mktemp -d)" \
	  && wget \
        --no-check-certificate \
        -nv \
        -O ${DISTR_DIR}/${DISTRO_NAME}.tar.gz \
        "https://downloads.apache.org/zookeeper/${DISTRO_VERSION_NAME}/${DISTRO_NAME}.tar.gz" \
    && tar -xf ${DISTR_DIR}/${DISTRO_NAME}.tar.gz -C ${DISTR_DIR} \
    && mkdir -p ${ZOOKEEPER_HOME} \
    && mv ${DISTR_DIR}/${DISTRO_NAME}/* ${ZOOKEEPER_HOME} \
    && rm -rf ${DISTR_DIR} \
    && mkdir -p ${ZOOKEEPER_OPT} \
    && mkdir -p ${ZOOKEEPER_DATA}

ARG JOLOKIA_VERSION="1.7.1"
# Download Jolokia
RUN set -x \
    && wget \
        --no-check-certificate \
        -nv \
        -O ${ZOOKEEPER_HOME}/lib/jolokia-jvm-${JOLOKIA_VERSION}.jar \
        "https://repo1.maven.org/maven2/org/jolokia/jolokia-jvm/${JOLOKIA_VERSION}/jolokia-jvm-${JOLOKIA_VERSION}.jar"

ARG PROMETHEUS_JMX_EXPORTER_VERSION="1.1.0"
# Download jmx_prometheus_javaagent
RUN set -x \
    && wget \
        --no-check-certificate \
        -nv \
        -O ${ZOOKEEPER_HOME}/lib/jmx_prometheus_javaagent-${PROMETHEUS_JMX_EXPORTER_VERSION}.jar \
        "https://github.com/prometheus/jmx_exporter/releases/download/${PROMETHEUS_JMX_EXPORTER_VERSION}/jmx_prometheus_javaagent-${PROMETHEUS_JMX_EXPORTER_VERSION}.jar"

ARG TMP_DIR="/tmp"
## Download and install async-profiler
RUN set -x \
    && wget \
        --no-check-certificate \
        -nv \
        -O ${TMP_DIR}/async-profiler-3.0-linux-x64.tar.gz \
        "https://github.com/async-profiler/async-profiler/releases/download/v3.0/async-profiler-3.0-linux-x64.tar.gz" \
    && tar -xzf /tmp/async-profiler-3.0-linux-x64.tar.gz -C /opt/ \
    && ln -s /opt/async-profiler-3.0-linux-x64/profiler.sh /usr/local/bin/profiler \
    && rm -f ${TMP_DIR}/async-profiler-3.0-linux-x64.tar.gz

# Copy config and script files
COPY docker/config/logback.xml ${ZOOKEEPER_HOME}/conf/logback.xml
COPY docker/config/zoo.cfg ${ZOOKEEPER_HOME}/conf/zoo.cfg
COPY docker/config/jmx-exporter-config.yaml ${ZOOKEEPER_HOME}/conf/jmx-exporter-config.yaml
COPY docker/docker-entrypoint.sh /
COPY docker/zkHealth.sh ${ZOOKEEPER_HOME}/bin
COPY docker/dr_reconfig.sh ${ZOOKEEPER_HOME}/bin

#ARG CDT_VERSION="9.3.2.78.2"
#ENV NC_DIAGNOSTIC_FOLDER=/app/ncdiag
#RUN mkdir -p ${NC_DIAGNOSTIC_FOLDER} \
#    && wget https://artifactorycn.netcracker.com/pd.saas.mvn/com/netcracker/execution-statistics-collector/installer-cloud/${CDT_VERSION}/installer-cloud-${CDT_VERSION}.zip \
#    && unzip -oq installer-cloud-${CDT_VERSION}.zip -d ${NC_DIAGNOSTIC_FOLDER} \
#    && rm -f installer-cloud-${CDT_VERSION}.zip \
#    && chmod a+rwx -R ${NC_DIAGNOSTIC_FOLDER}

# Adapt grants
RUN chgrp -R 0 ${ZOOKEEPER_HOME} \
    && chgrp -R 0 ${ZOOKEEPER_OPT} \
    && chmod -R g+rw ${ZOOKEEPER_HOME} \
    && chmod -R g+rw ${ZOOKEEPER_OPT} \
    && find ${ZOOKEEPER_HOME} -type d -exec chmod g+x {} + \
    && find ${ZOOKEEPER_OPT} -type d -exec chmod g+x {} + \
    && chmod +x ${ZOOKEEPER_HOME}/bin/*.sh \
    && chmod +x /docker-entrypoint.sh

# Upgrade all tools to avoid vulnerabilities
RUN set -x && apk upgrade --no-cache --available

USER 1000:0
WORKDIR ${ZOOKEEPER_HOME}

# Expose the ports and set up volumes for the data directory
EXPOSE ${ZOOKEEPER_CLIENT_PORT} ${ZOOKEEPER_FOLLOWERS_PORT} ${ZOOKEEPER_ELECTION_PORT} ${ZOOKEEPER_BACKUP_PORT}
VOLUME ["${ZOOKEEPER_DATA}"]

ENTRYPOINT ["/sbin/tini" ,"--", "/docker-entrypoint.sh"]
CMD ["start"]
