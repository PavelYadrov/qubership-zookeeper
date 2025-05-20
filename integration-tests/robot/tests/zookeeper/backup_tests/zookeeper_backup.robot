*** Variables ***
${ZOOKEEPER_NODE_PATH}                /tests_znode
${ZOOKEEPER_BACKUP_DAEMON_PROTOCOL}   %{ZOOKEEPER_BACKUP_DAEMON_PROTOCOL}
${ZOOKEEPER_BACKUP_DAEMON_HOST}       %{ZOOKEEPER_BACKUP_DAEMON_HOST}
${ZOOKEEPER_BACKUP_DAEMON_PORT}       %{ZOOKEEPER_BACKUP_DAEMON_PORT}
${ZOOKEEPER_BACKUP_DAEMON_USERNAME}   %{ZOOKEEPER_BACKUP_DAEMON_USERNAME}
${ZOOKEEPER_BACKUP_DAEMON_PASSWORD}   %{ZOOKEEPER_BACKUP_DAEMON_PASSWORD}
${CREATE_BACKUP_TIMEOUT}              30min
${CREATE_BACKUP_TIME_INTERVAL}        10s
${DELETE_BACKUP_TIMEOUT}              5min
${DELETE_BACKUP_TIME_INTERVAL}        10s
${RESTORE_BACKUP_TIMEOUT}             5min
${RESTORE_BACKUP_TIME_INTERVAL}       10s

*** Settings ***
Library  String
Library	 RequestsLibrary
Resource  ../../shared/keywords.robot
Suite Setup  Setup
Suite Teardown  Cleanup
Test Setup  Prepare Backup Daemon Session

*** Keywords ***
Setup
    ${zk} =  Connect To Zookeeper
    Set Suite Variable  ${zk}
    Create Node  ${zk}  ${ZOOKEEPER_NODE_PATH}

Check Existence Of Node
    [Arguments]  ${node_path}
    ${node} =  Node Exists  ${zk}  ${node_path}
    Should Be True  ${node}

Check Absence Of Node
    [Arguments]  ${node_path}
    ${node} =  Node Exists  ${zk}  ${node_path}
    Should Not Be True  ${node}

Prepare Backup Daemon Session
    ${auth}=  Create List  ${ZOOKEEPER_BACKUP_DAEMON_USERNAME}  ${ZOOKEEPER_BACKUP_DAEMON_PASSWORD}
    ${verify}=  Set Variable If  '${ZOOKEEPER_BACKUP_DAEMON_PROTOCOL}' == 'https'  /backupTLS/ca.crt  ${False}
    Create Session  backup_daemon  ${ZOOKEEPER_BACKUP_DAEMON_PROTOCOL}://${ZOOKEEPER_BACKUP_DAEMON_HOST}:${ZOOKEEPER_BACKUP_DAEMON_PORT}
    ...  disable_warnings=1  auth=${auth}  verify=${verify}
    ${headers}=  Create Dictionary  Content-Type=application/json  Accept=application/json
    Set Global Variable  ${headers}

Check That Backup Is Presented
    [Arguments]  ${backup_id}
    ${response}=  GET On Session  backup_daemon  /listbackups/${backup_id}
    Should Be Equal As Strings  ${response.json()}[failed]  False
    Should Be Equal As Strings  ${response.json()}[valid]  True

Check That Backup Is Not Presented
    [Arguments]  ${backup_id}
    ${backups}=  GET On Session  backup_daemon  /listbackups
    ${is_contained}=  Evaluate  '${backups.content}'.find("${backup_id}") == 1
    Should Not Be True  ${is_contained}

Check That Restore Succeed
    [Arguments]  ${restore_resp}
    ${response}=  GET On Session  backup_daemon  /jobstatus/${restore_resp.content}
    Should Be Equal As Strings  ${response.json()}[status]  Successful

Create Nodes For Restore Test
    Create Node  ${zk}  ${ZOOKEEPER_NODE_PATH}/tmp1  tmp1
    Create Node  ${zk}  ${ZOOKEEPER_NODE_PATH}/tmp2  tmp2
    Create Node  ${zk}  ${ZOOKEEPER_NODE_PATH}/tmp3  tmp3
    Create Node  ${zk}  ${ZOOKEEPER_NODE_PATH}/tmp4  tmp4

Create Nodes For Restore Advanced
    Create Node  ${zk}  ${ZOOKEEPER_NODE_PATH}/tmp1  123$%_&$/n
    Create Node  ${zk}  ${ZOOKEEPER_NODE_PATH}/tmp2  темп2
    Create Node  ${zk}  ${ZOOKEEPER_NODE_PATH}/tmp3  ""
    Create Node  ${zk}  ${ZOOKEEPER_NODE_PATH}/tmp3/nested  tmp3
    Create Byte Node  ${zk}  ${ZOOKEEPER_NODE_PATH}/tmp4  \x00\x01\xffsd

Create Nodes For Hierarchical Backup Test
    Create Node  ${zk}  ${ZOOKEEPER_NODE_PATH}/tmp1  tmp1
    Create Node  ${zk}  ${ZOOKEEPER_NODE_PATH}/tmp2  tmp2

Create Nodes For Transactional Backup Test
    Create Node  ${zk}  ${ZOOKEEPER_NODE_PATH}/test_b  t1
    Create Node  ${zk}  ${ZOOKEEPER_NODE_PATH}/test_b1  T1
    update node value  ${zk}  ${ZOOKEEPER_NODE_PATH}/test_b1  T2
    update node value  ${zk}  ${ZOOKEEPER_NODE_PATH}/test_b1  T3

Delete Nodes For Restore Test
    Delete Node  ${zk}  ${ZOOKEEPER_NODE_PATH}/tmp1
    Delete Node  ${zk}  ${ZOOKEEPER_NODE_PATH}/tmp2
    Check Absence Of Node  ${ZOOKEEPER_NODE_PATH}/tmp1
    Check Absence Of Node  ${ZOOKEEPER_NODE_PATH}/tmp2

Delete Nodes For Restore Advanced
    Delete Node  ${zk}  ${ZOOKEEPER_NODE_PATH}/tmp1
    Delete Node  ${zk}  ${ZOOKEEPER_NODE_PATH}/tmp2
    Delete Node  ${zk}  ${ZOOKEEPER_NODE_PATH}/tmp3
    Delete Node  ${zk}  ${ZOOKEEPER_NODE_PATH}/tmp4
    Check Absence Of Node  ${ZOOKEEPER_NODE_PATH}/tmp1
    Check Absence Of Node  ${ZOOKEEPER_NODE_PATH}/tmp2
    Check Absence Of Node  ${ZOOKEEPER_NODE_PATH}/tmp3
    Check Absence Of Node  ${ZOOKEEPER_NODE_PATH}/tmp3/nested
    Check Absence Of Node  ${ZOOKEEPER_NODE_PATH}/tmp4

Delete Nodes For Restore High Load
    Delete Node  ${zk}  ${ZOOKEEPER_NODE_PATH}/highload
    Check Absence Of Node  ${ZOOKEEPER_NODE_PATH}/highload

Cleanup Data For Restore Test
    Delete Node  ${zk}  ${ZOOKEEPER_NODE_PATH}/tmp1
    Delete Node  ${zk}  ${ZOOKEEPER_NODE_PATH}/tmp2
    Delete Node  ${zk}  ${ZOOKEEPER_NODE_PATH}/tmp3
    Delete Node  ${zk}  ${ZOOKEEPER_NODE_PATH}/tmp4

Cleanup Data For Hierarchical Backup Test
    Delete Node  ${zk}  ${ZOOKEEPER_NODE_PATH}/tmp1
    Delete Node  ${zk}  ${ZOOKEEPER_NODE_PATH}/tmp2

Cleanup Data For Transactional Backup Test
    Delete Node  ${zk}  ${ZOOKEEPER_NODE_PATH}/test_b
    Delete Node  ${zk}  ${ZOOKEEPER_NODE_PATH}/test_b1

Check Node Value
    [Arguments]  ${path}  ${expected_data}
    ${data} =  Get Node Value  ${zk}  ${path}
    Should Be Equal As Strings  ${data}  ${expected_data}

Check Byte Node Value
    [Arguments]  ${path}  ${expected_data}
    ${data} =  Get Byte Node Value  ${zk}  ${path}
    Should Be Equal As Strings  ${data}  ${expected_data}

Check That Nodes Exist After Restore
    Check Existence Of Node  ${ZOOKEEPER_NODE_PATH}/tmp1
    Check Node Value  ${ZOOKEEPER_NODE_PATH}/tmp1  tmp1
    Check Existence Of Node  ${ZOOKEEPER_NODE_PATH}/tmp2
    Check Node Value  ${ZOOKEEPER_NODE_PATH}/tmp2  tmp2
    Check Existence Of Node  ${ZOOKEEPER_NODE_PATH}/tmp3
    Check Node Value  ${ZOOKEEPER_NODE_PATH}/tmp3  tmp3
    Check Existence Of Node  ${ZOOKEEPER_NODE_PATH}/tmp4
    Check Node Value  ${ZOOKEEPER_NODE_PATH}/tmp4  tmp4

Check That Nodes Exist After Restore Advanced
    Check Existence Of Node  ${ZOOKEEPER_NODE_PATH}/tmp1
    Check Node Value  ${ZOOKEEPER_NODE_PATH}/tmp1  123$%_&$/n
    Check Existence Of Node  ${ZOOKEEPER_NODE_PATH}/tmp2
    Check Node Value  ${ZOOKEEPER_NODE_PATH}/tmp2  темп2
    Check Existence Of Node  ${ZOOKEEPER_NODE_PATH}/tmp3/nested
    Check Node Value  ${ZOOKEEPER_NODE_PATH}/tmp3/nested  tmp3
    Check Existence Of Node  ${ZOOKEEPER_NODE_PATH}/tmp4
    Check Byte Node Value  ${ZOOKEEPER_NODE_PATH}/tmp4  \x00\x01\xffsd

Check That Backup Is Created And Delete It
    [Arguments]  ${resp_backup}
    ${backup_id}=  Set Variable  ${resp_backup.content}
    Wait Until Keyword Succeeds  ${CREATE_BACKUP_TIMEOUT}  ${CREATE_BACKUP_TIME_INTERVAL}
    ...  Check That Backup Is Presented  ${backup_id}
    POST On Session  backup_daemon  /evict/${backup_id}
    Wait Until Keyword Succeeds  ${DELETE_BACKUP_TIMEOUT}  ${DELETE_BACKUP_TIME_INTERVAL}
    ...  Check That Backup Is Not Presented  ${backup_id}

Delete Current Backup
    [Arguments]  ${backup_id}
    Run Keyword If  '${backup_id}'
    ...  POST On Session  backup_daemon  /evict/${backup_id}

Cleanup
    Delete Node  ${zk}  ${ZOOKEEPER_NODE_PATH}
    Disconnect From Zookeeper  ${zk}
    ${zk} =  Set Variable  ${None}

*** Test Cases ***
Restore Hierarchical Backup
    [Tags]  zookeeper  zookeeper_backup_daemon  restore_hierarchical_backup  hierarchical_backup
    Set Test Variable  ${backup_id}  ${EMPTY}
    Create Nodes For Restore Test
    ${resp_backup}=  POST On Session  backup_daemon  /backup
    ${backup_id}=  Set Variable  ${resp_backup.content}
    Wait Until Keyword Succeeds  ${CREATE_BACKUP_TIMEOUT}  ${CREATE_BACKUP_TIME_INTERVAL}
    ...  Check That Backup Is Presented  ${backup_id}
    Delete Nodes For Restore Test
    ${restore_data}=  Set Variable  {"vault":"${backup_id}", "dbs":["tests_znode"]}
    ${restore_resp}=  POST On Session  backup_daemon  /restore  headers=${headers}  data=${restore_data}
    Wait Until Keyword Succeeds  ${RESTORE_BACKUP_TIMEOUT}  ${RESTORE_BACKUP_TIME_INTERVAL}
    ...  Check That Restore Succeed  ${restore_resp}
    Check That Nodes Exist After Restore
    [Teardown]  Run Keywords
                ...  Cleanup Data For Restore Test
                ...  AND
                ...  Delete Current Backup  ${backup_id}

Restore Hierarchical Backup Advanced
    [Tags]  zookeeper  zookeeper_backup_daemon  restore_hierarchical_backup_advanced  hierarchical_backup
    Set Test Variable  ${backup_id}  ${EMPTY}
    Create Nodes For Restore Advanced
    ${resp_backup}=  POST On Session  backup_daemon  /backup
    ${backup_id}=  Set Variable  ${resp_backup.content}
    Wait Until Keyword Succeeds  ${CREATE_BACKUP_TIMEOUT}  ${CREATE_BACKUP_TIME_INTERVAL}
    ...  Check That Backup Is Presented  ${backup_id}
    Delete Nodes For Restore Advanced
    ${restore_data}=  Set Variable  {"vault":"${backup_id}", "dbs":["tests_znode"]}
    ${restore_resp}=  POST On Session  backup_daemon  /restore  headers=${headers}  data=${restore_data}
    Wait Until Keyword Succeeds  ${RESTORE_BACKUP_TIMEOUT}  ${RESTORE_BACKUP_TIME_INTERVAL}
    ...  Check That Restore Succeed  ${restore_resp}
    Check That Nodes Exist After Restore Advanced
    [Teardown]  Run Keywords
                ...  Cleanup Data For Restore Test
                ...  AND
                ...  Delete Current Backup  ${backup_id}

Restore Hierarchical Backup High Load
    [Tags]  zookeeper  zookeeper_backup_daemon  restore_hierarchical_backup_high_load  hierarchical_backup
    Set Test Variable  ${backup_id}  ${EMPTY}
    Create Node With Children  ${zk}  ${ZOOKEEPER_NODE_PATH}/highload  2562  test_data 
    ${resp_backup}=  POST On Session  backup_daemon  /backup
    ${backup_id}=  Set Variable  ${resp_backup.content}
    Wait Until Keyword Succeeds  ${CREATE_BACKUP_TIMEOUT}  ${CREATE_BACKUP_TIME_INTERVAL}
    ...  Check That Backup Is Presented  ${backup_id}
    Delete Nodes For Restore High Load
    ${restore_data}=  Set Variable  {"vault":"${backup_id}", "dbs":["tests_znode"]}
    ${restore_resp}=  POST On Session  backup_daemon  /restore  headers=${headers}  data=${restore_data}
    Wait Until Keyword Succeeds  ${RESTORE_BACKUP_TIMEOUT}  ${RESTORE_BACKUP_TIME_INTERVAL}
    ...  Check That Restore Succeed  ${restore_resp}
    Check Existence Of Node  ${ZOOKEEPER_NODE_PATH}/highload
    [Teardown]  Run Keywords
                ...  Delete Node  ${zk}  ${ZOOKEEPER_NODE_PATH}/highload
                ...  AND
                ...  Delete Current Backup  ${backup_id}

Create And Delete Hierarchical Backup
    [Tags]  zookeeper  zookeeper_backup_daemon  create_and_delete_hierarchical_backup  hierarchical_backup
    Set Test Variable  ${backup_id}  ${EMPTY}
    Create Nodes For Hierarchical Backup Test
    ${resp_backup}=  POST On Session  backup_daemon  /backup
    Check That Backup Is Created And Delete It  ${resp_backup}
    [Teardown]  Run Keywords
                ...  Cleanup Data For Hierarchical Backup Test
                ...  AND
                ...  Delete Current Backup  ${backup_id}

Create And Delete Transactional Backup
    [Tags]  zookeeper  zookeeper_backup_daemon  create_and_delete_transactional_backup  transactional_backup
    Set Test Variable  ${backup_id}  ${EMPTY}
    Create Nodes For Transactional Backup Test
    ${resp_backup}=  POST On Session  backup_daemon  /backup  headers=${headers}  data={"mode":"transactional"}
    Check That Backup Is Created And Delete It  ${resp_backup}
    [Teardown]  Run Keywords
                ...  Cleanup Data For Transactional Backup Test
                ...  AND
                ...  Delete Current Backup  ${backup_id}

Full Eviction Test
    [Tags]  zookeeper  zookeeper_backup_daemon  full_eviction_test
    POST On Session  backup_daemon  /evict

Unauthorized Access
    [Tags]  zookeeper  zookeeper_backup_daemon  unauthorized_access
    Create Session  backup_daemon_unauthorized  ${ZOOKEEPER_BACKUP_DAEMON_PROTOCOL}://${ZOOKEEPER_BACKUP_DAEMON_HOST}:${ZOOKEEPER_BACKUP_DAEMON_PORT}
    ...  disable_warnings=1
    POST On Session  backup_daemon_unauthorized  /backup  expected_status=401