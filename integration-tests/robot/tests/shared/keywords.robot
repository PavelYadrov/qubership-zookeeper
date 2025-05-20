*** Variables ***
${ZOOKEEPER_HOST}  %{ZOOKEEPER_HOST}
${ZOOKEEPER_PORT}  %{ZOOKEEPER_PORT}


*** Settings ***
Library  ./lib/ZookeeperLibrary.py  zookeeper_os_project=%{ZOOKEEPER_OS_PROJECT}
...                                 zookeeper_host=${ZOOKEEPER_HOST}
...                                 zookeeper_port=${ZOOKEEPER_PORT}
...                                 zookeeper_enable_ssl=%{ZOOKEEPER_ENABLE_TLS}