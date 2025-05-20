*** Variables ***
${ZOOKEEPER_HA_NODE_PATH}              /zookeeper_ha
${ZOOKEEPER_DISK_IS_FILLED_NODE_PATH}  /zookeeper_ha/disk_is_filled_on_one_node
${LIBRARIES}                           /opt/robot/tests/shared/lib
${OPERATION_RETRY_COUNT}               17x
${OPERATION_RETRY_INTERVAL}            4s
${DISK_FILLED_RETRY_COUNT}             30x
${DISK_FILLED_RETRY_INTERVAL}          5s
${SLEEP_TIME}                          20s

*** Settings ***
Library  Collections
Library  OperatingSystem
Library  String
Library  PlatformLibrary  managed_by_operator=%{ZOOKEEPER_IS_MANAGED_BY_OPERATOR}
Resource  ../../shared/keywords.robot

*** Keywords ***
Scale Up Deployment
    [Arguments]  ${deployment}  ${project}
    Scale Up Deployment Entity  ${deployment}  ${project}
    ${status}=  Check Service Is Scaled  ${deployment}  ${project}  direction=up  timeout=300
    Should Be True  ${status}

Scale Down Deployment
    [Arguments]  ${deployment}  ${project}
    Scale Down Deployment Entity  ${deployment}  ${project}
    ${status}=  Check Service Is Scaled  ${deployment}  ${project}  direction=down  timeout=300
    Should Be True  ${status}

Find Leader
    [Arguments]  ${deployments}
    ${command} =  Catenate  srvr
    ${leader} =  Set Variable  ${None}
    FOR  ${deployment}  IN  @{deployments}
      ${zk} =  Connect To Zookeeper Node  ${deployment}
      ${result} =  Execute Command  ${zk}  ${command}
      Disconnect From Zookeeper  ${zk}
      ${zk} =  Set Variable  ${None}
      ${leader} =  Run Keyword If  'Mode: leader' in '''${result}'''  Set Variable  ${deployment}
      Exit For Loop If  'Mode: leader' in '''${result}'''
    END
    Should Not Be Equal  ${leader}  ${None}
    [Return]  ${leader}

Create Zookeeper Node
    ${value} =  Generate Random String  100000  [LETTERS][NUMBERS]
    ${new_index} =  Evaluate  ${index} + 1
    Set Test Variable  ${index}  ${new_index}
    ${result}  ${errors}  Execute Command In Pod  ${pod_name}  %{ZOOKEEPER_OS_PROJECT}
    ...  ./bin/zkCli.sh create ${ZOOKEEPER_DISK_IS_FILLED_NODE_PATH}/${index} ${value}
    Should Contain  ${errors}
    ...  KeeperErrorCode = ConnectionLoss for ${ZOOKEEPER_DISK_IS_FILLED_NODE_PATH}

Check Disk Is Full
    [Arguments]  ${disk_space}
    ${filled_space} =  Get Filled Space
    ${disk_fullness} =  Evaluate  100 * ${filled_space} / ${disk_space}
    Should Be True  ${disk_fullness} > 90

Get Filled Space
    ${disk_information}  ${errors}  Execute Command In Pod  ${pod_name}  %{ZOOKEEPER_OS_PROJECT}
    ...  du -m /var/opt/zookeeper/data -d 0
    ${filled_space_in_mb}  ${disk_folder} =  Split String  ${disk_information}  ${EMPTY}  1
    [Return]  ${filled_space_in_mb}

Get Disk Space
    [Documentation]  There are cases when disk space differs from PV size:
    ...  1) Available disk space is much less than PV size (for example, command 'df -h' shows that
    ...  3G is available, but PV size is 10G).
    ...  2) If classical hostPath is used, command 'df -h' shows free space on the root system, but
    ...  in reality available space depends on PV size (for example, command 'df -h' shows
    ...  that 80G is available, but PV size is 2G).
    ${full_information}  ${errors}   Execute Command In Pod  ${pod_name}  %{ZOOKEEPER_OS_PROJECT}
    ...  df -m /var/opt/zookeeper/data
    ${header}  ${disk_information}  Split String  ${full_information}  \n  1
    ${disk}  ${space_size_in_mb}  ${rest}  Split String  ${disk_information}  ${EMPTY}  2
    ${space_size_int} =  Convert To Integer  ${space_size_in_mb}
    ${pv_size_in_mb} =  Evaluate  %{ZOOKEEPER_VOLUME_SIZE} * 1024
    ${minimum} =  Find Minimum  ${space_size_int}  ${pv_size_in_mb}
    [Return]  ${minimum}

Clean Disk Space
    Execute Command In Pod  ${pod_name}  %{ZOOKEEPER_OS_PROJECT}  rm /var/opt/zookeeper/data/busy_space
    Execute Command In Pod  ${pod_name}  %{ZOOKEEPER_OS_PROJECT}  ./bin/zkCli.sh deleteall ${ZOOKEEPER_HA_NODE_PATH}

*** Test Cases ***
Test Zookeeper Without Leader Node
    [Tags]  zookeeper_ha  zookeeper_ha_without_leader  zookeeper
    ${deployments}=  Get Active Deployment Entities Names For Service  %{ZOOKEEPER_OS_PROJECT}  %{ZOOKEEPER_HOST}
    ${number_of_deployments} =  Get Length  ${deployments}
    Pass Execution If  ${number_of_deployments} < 3
    ...  The test is skipped due to insufficient number of ZooKeeper replicas for this test (minimum value is 3 replicas).
    ${leader} =  Wait Until Keyword Succeeds  ${OPERATION_RETRY_COUNT}  ${OPERATION_RETRY_INTERVAL}
    ...  Find Leader  ${deployments}
    Log  Current ZooKeeper leader is ${leader}  DEBUG

    Scale Down Deployment  ${leader}  %{ZOOKEEPER_OS_PROJECT}

    Remove Values From List  ${deployments}  ${leader}
    ${new_leader} =  Wait Until Keyword Succeeds  ${OPERATION_RETRY_COUNT}  ${OPERATION_RETRY_INTERVAL}
    ...  Find Leader  ${deployments}
    Log  Current ZooKeeper leader is ${new_leader}  DEBUG

    [Teardown]  Scale Up Deployment  ${leader}  %{ZOOKEEPER_OS_PROJECT}

Test Disk Is Filled On One Node
    [Tags]  zookeeper_ha  zookeeper_ha_disk_is_filled
    ${pod_names}=  Get Pod Names By Service Name  %{ZOOKEEPER_HOST}  %{ZOOKEEPER_OS_PROJECT}
    ${pod_name}=  Get From List  ${pod_names}  0
    Set Suite Variable  ${pod_name}
    Execute Command In Pod  ${pod_name}  %{ZOOKEEPER_OS_PROJECT}
    ...  ./bin/zkCli.sh create ${ZOOKEEPER_HA_NODE_PATH} ha_node
    Execute Command In Pod  ${pod_name}  %{ZOOKEEPER_OS_PROJECT}
    ...  ./bin/zkCli.sh create ${ZOOKEEPER_DISK_IS_FILLED_NODE_PATH} disk_is_filled_on_one_node

    ${filled_space_in_mb} =  Get Filled Space
    ${disk_space_in_mb} =  Get Disk Space
    ${10_gigabytes} =  Evaluate  10 * 1024
    Run Keyword If  ${disk_space_in_mb} > ${10_gigabytes}
    ...  Pass Execution  Current test can't be executed due to too large size (${disk_space_in_mb}Mb) of the ZooKeeper storage
    ${free_space_in_mb} =  Evaluate  ${disk_space_in_mb} - ${filled_space_in_mb}
    ${evaluate_amount_of_blocks} =  Evaluate  ${free_space_in_mb} / 50
    ${blocks_count} =  Convert To Integer  ${evaluate_amount_of_blocks}
    Log  Node space is filling  DEBUG
    Run Keyword And Ignore Error  Execute Command In Pod  ${pod_name}  %{ZOOKEEPER_OS_PROJECT}
    ...  dd if=/dev/zero of=/var/opt/zookeeper/data/busy_space bs=50M count=${blocks_count}

    Wait Until Keyword Succeeds  ${DISK_FILLED_RETRY_COUNT}  ${DISK_FILLED_RETRY_INTERVAL}
    ...  Check Disk Is Full  ${disk_space_in_mb}
    Log  Node space is filled with more than 90%  DEBUG

    Set Test Variable  ${index}  -1
    Wait Until Keyword Succeeds  ${OPERATION_RETRY_COUNT}  ${OPERATION_RETRY_INTERVAL}
    ...  Create Zookeeper Node

    [Teardown]  Clean Disk Space