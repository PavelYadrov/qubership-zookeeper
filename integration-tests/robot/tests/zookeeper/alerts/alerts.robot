*** Variables ***
${ZOOKEEPER_OS_PROJECT}              %{ZOOKEEPER_OS_PROJECT}
${ZOOKEEPER_IS_DEGRADED_ALERT_NAME}  ZooKeeper_Is_Degraded_Alarm
${ZOOKEEPER_IS_DOWN_ALERT_NAME}      ZooKeeper_Is_Down_Alarm
${ALERT_RETRY_TIME}                      5min
${ALERT_RETRY_INTERVAL}                  10s
${SLEEP_TIME}                            30s

*** Settings ***
Library  MonitoringLibrary  host=%{PROMETHEUS_URL}
...                         username=%{PROMETHEUS_USER}
...                         password=%{PROMETHEUS_PASSWORD}
Resource  ../../shared/keywords.robot
Library  PlatformLibrary  managed_by_operator=%{ZOOKEEPER_IS_MANAGED_BY_OPERATOR}

*** Keywords ***
Check That Prometheus Alert Is Active
    [Arguments]  ${alert_name}
    ${status}=  Get Alert Status  ${alert_name}  ${ZOOKEEPER_OS_PROJECT}
    Should Be Equal As Strings  ${status}  pending

Check That Prometheus Alert Is Inactive
    [Arguments]  ${alert_name}
    ${status}=  Get Alert Status  ${alert_name}  ${ZOOKEEPER_OS_PROJECT}
    Should Be Equal As Strings  ${status}  inactive

Check That Zookeeper Broker Is Up
    [Arguments]  ${service_name}
    ${replicas}=  Get Active Deployment Entities Count For Service  ${ZOOKEEPER_OS_PROJECT}  ${service_name}
    Run keyword if  "${replicas}" == "0"
    ...  Scale Up Deployment Entities By Service Name  ${service_name}  ${ZOOKEEPER_OS_PROJECT}  replicas=1  with_check=True
    Sleep  ${SLEEP_TIME}

Check JMX Metrics Exist In Prometheus
    ${data}=  Get Metric Values  zookeeper_Leader
    Should Contain  str(${data})  ${ZOOKEEPER_OS_PROJECT}

*** Test Cases ***
ZooKeeper Is Degraded Alert
    [Tags]  zookeeper  prometheus  zookeeper_prometheus_alert  zookeeper_is_degraded_alert
    Check That Prometheus Alert Is Inactive  ${ZOOKEEPER_IS_DEGRADED_ALERT_NAME}
    ${replicas}=  Get Active Deployment Entities Count For Service  ${ZOOKEEPER_OS_PROJECT}  ${ZOOKEEPER_HOST}
    Pass Execution If  ${replicas} < 3  ZooKeeper cluster has less than 3 brokers
    Scale Down Deployment Entities By Service Name  ${ZOOKEEPER_HOST}-1  ${ZOOKEEPER_OS_PROJECT}
    Wait Until Keyword Succeeds  ${ALERT_RETRY_TIME}  ${ALERT_RETRY_INTERVAL}
    ...  Check That Prometheus Alert Is Active  ${ZOOKEEPER_IS_DEGRADED_ALERT_NAME}
    Scale Up Deployment Entities By Service Name  ${ZOOKEEPER_HOST}-1  ${ZOOKEEPER_OS_PROJECT}  replicas=1
    Wait Until Keyword Succeeds  ${ALERT_RETRY_TIME}  ${ALERT_RETRY_INTERVAL}
    ...  Check That Prometheus Alert Is Inactive  ${ZOOKEEPER_IS_DEGRADED_ALERT_NAME}
    [Teardown]  Check That Zookeeper Broker Is Up  ${ZOOKEEPER_HOST}-1

ZooKeeper Is Down Alert
    [Tags]  zookeeper  prometheus  zookeeper_prometheus_alert  zookeeper_is_down_alert
    Check That Prometheus Alert Is Inactive  ${ZOOKEEPER_IS_DOWN_ALERT_NAME}
    ${replicas}=  Get Active Deployment Entities Count For Service  ${ZOOKEEPER_OS_PROJECT}  ${ZOOKEEPER_HOST}
    Pass Execution If  ${replicas} < 3  ZooKeeper cluster has less than 3 brokers
    Scale Down Deployment Entities By Service Name  ${ZOOKEEPER_HOST}  ${ZOOKEEPER_OS_PROJECT}
    Wait Until Keyword Succeeds  ${ALERT_RETRY_TIME}  ${ALERT_RETRY_INTERVAL}
    ...  Check That Prometheus Alert Is Active  ${ZOOKEEPER_IS_DOWN_ALERT_NAME}
    Scale Up Deployment Entities By Service Name  ${ZOOKEEPER_HOST}  ${ZOOKEEPER_OS_PROJECT}  replicas=1
    Wait Until Keyword Succeeds  ${ALERT_RETRY_TIME}  ${ALERT_RETRY_INTERVAL}
    ...  Check That Prometheus Alert Is Inactive  ${ZOOKEEPER_IS_DOWN_ALERT_NAME}
    [Teardown]  Check That Zookeeper Broker Is Up  ${ZOOKEEPER_HOST}


Check JMX Metrics
    [Tags]  zookeeper  prometheus  jmx_metrics
    Wait Until Keyword Succeeds  ${ALERT_RETRY_TIME}  ${ALERT_RETRY_INTERVAL}
    ...  Check JMX Metrics Exist In Prometheus
