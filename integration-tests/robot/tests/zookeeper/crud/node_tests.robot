*** Variables ***
${ZOOKEEPER_CRUD_NODE_PATH}   /zookeeper_crud
${ZOOKEEPER_TESTS_NODE_PATH}  /zookeeper_crud/tests
${CREATION_DATA}              Creation data
${MODIFICATION_DATA}          Modification data

*** Settings ***
Library  String
Resource  ../../shared/keywords.robot
Suite Setup  Setup
Suite Teardown  Cleanup

*** Keywords ***
Setup
    ${zk} =  Connect To Zookeeper
    Set Suite Variable  ${zk}
    Delete Node  ${zk}  ${ZOOKEEPER_CRUD_NODE_PATH}

Check Existence Of Node
    ${node} =  Node Exists  ${zk}  ${ZOOKEEPER_TESTS_NODE_PATH}
    Should Be True  ${node}

Check Absence Of Node
    ${node} =  Node Exists  ${zk}  ${ZOOKEEPER_TESTS_NODE_PATH}
    Should Not Be True  ${node}

Cleanup
    Delete Node  ${zk}  ${ZOOKEEPER_CRUD_NODE_PATH}
    Disconnect From Zookeeper  ${zk}
    ${zk} =  Set Variable  ${None}

*** Test Cases ***
Test Node Creation
    [Tags]  zookeeper_crud  zookeeper
    Check Absence Of Node
    Create Node  ${zk}  ${ZOOKEEPER_CRUD_NODE_PATH}  ${CREATION_DATA}
    Create Node  ${zk}  ${ZOOKEEPER_TESTS_NODE_PATH}  ${CREATION_DATA}
    Check Existence Of Node

Test Reading Data
    [Tags]  zookeeper_crud  zookeeper
    Check Existence Of Node
    ${data} =  Get Node Value  ${zk}  ${ZOOKEEPER_TESTS_NODE_PATH}
    Should Be Equal As Strings  ${data}  ${CREATION_DATA}

Test Updating Data
    [Tags]  zookeeper_crud  zookeeper
    Check Existence Of Node
    Update Node Value  ${zk}  ${ZOOKEEPER_TESTS_NODE_PATH}  ${MODIFICATION_DATA}
    ${data} =  Get Node Value  ${zk}  ${ZOOKEEPER_TESTS_NODE_PATH}
    Should Be Equal As Strings  ${data}  ${MODIFICATION_DATA}

Test Node Deletion
    [Tags]  zookeeper_crud  zookeeper
    Check Existence Of Node
    Delete Node  ${zk}  ${ZOOKEEPER_CRUD_NODE_PATH}
    Check Absence Of Node