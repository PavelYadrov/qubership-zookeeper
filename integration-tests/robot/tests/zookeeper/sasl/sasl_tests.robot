*** Variables ***
${ZOOKEEPER_ACL_NODE_PATH}    /zookeeper_acl
${ZOOKEEPER_TESTS_NODE_PATH}  /zookeeper_acl/tests
${ZOOKEEPER_ADMIN_USERNAME}   %{ZOOKEEPER_ADMIN_USERNAME}
${ZOOKEEPER_ADMIN_PASSWORD}   %{ZOOKEEPER_ADMIN_PASSWORD}
${ZOOKEEPER_CLIENT_USERNAME}  %{ZOOKEEPER_CLIENT_USERNAME}
${ZOOKEEPER_CLIENT_PASSWORD}  %{ZOOKEEPER_CLIENT_PASSWORD}
${ACL_VALUE}                  ACL

*** Settings ***
Library  String
Library  Collections
Resource  ../../shared/keywords.robot
Suite Setup  Setup
Suite Teardown  Cleanup

*** Keywords ***
Setup
    ${access_control_list} =  Create Access Control List  sasl  ${ZOOKEEPER_ADMIN_USERNAME}
    ${admin_zk} =  Connect To Zookeeper  sasl  ${ZOOKEEPER_ADMIN_USERNAME}  ${ZOOKEEPER_ADMIN_PASSWORD}  ${access_control_list}
    Set Suite Variable  ${admin_zk}
    Delete Node  ${admin_zk}  ${ZOOKEEPER_ACL_NODE_PATH}
    Create Node  ${admin_zk}  ${ZOOKEEPER_ACL_NODE_PATH}  ${ACL_VALUE}
    Create Node  ${admin_zk}  ${ZOOKEEPER_TESTS_NODE_PATH}  ${ACL_VALUE}

Add Acl To Node
    [Arguments]  ${access_control_list}
    ${acls} =  Get Acls  ${admin_zk}  ${ZOOKEEPER_TESTS_NODE_PATH}
    Append To List  ${acls}  ${access_control_list}
    Set Access Control Lists  ${admin_zk}  ${ZOOKEEPER_TESTS_NODE_PATH}  ${acls}

Remove Acl From Node
    [Arguments]  ${access_control_list}
    ${acls} =  Get Acls  ${admin_zk}  ${ZOOKEEPER_TESTS_NODE_PATH}
    Remove Values From List  ${acls}  ${access_control_list}
    Set Access Control Lists  ${admin_zk}  ${ZOOKEEPER_TESTS_NODE_PATH}  ${acls}

Cleanup
    Delete Node  ${admin_zk}  ${ZOOKEEPER_ACL_NODE_PATH}
    Disconnect From Zookeeper  ${admin_zk}
    ${admin_zk} =  Set Variable  ${None}

*** Test Cases ***
Test Client With All Grants Can Read Protected Node Data
    [Tags]  zookeeper_sasl  zookeeper_acl  zookeeper
    ${data} =  Get Node Value  ${admin_zk}  ${ZOOKEEPER_TESTS_NODE_PATH}
    Should Be Equal As Strings  ${data}  ${ACL_VALUE}

Test Unauthorized Client Can Not Read Protected Node Data
    [Tags]  zookeeper_sasl  zookeeper_acl  zookeeper
    ${zk} =  Connect To Zookeeper
    Run Keyword And Expect Error  NoAuthError  Get Node Value  ${zk}  ${ZOOKEEPER_TESTS_NODE_PATH}
    [Teardown]  Disconnect From Zookeeper  ${zk}

Test Client Without Read Grant Can Not Read Protected Node Data
    [Tags]  zookeeper_sasl  zookeeper_acl  zookeeper
    ${access_control_list} =  Create Access Control List  sasl  ${ZOOKEEPER_CLIENT_USERNAME}  read=${False}  all=${False}
    Add Acl To Node  ${access_control_list}
    ${zk} =  Connect To Zookeeper  sasl  ${ZOOKEEPER_CLIENT_USERNAME}  ${ZOOKEEPER_CLIENT_PASSWORD}  ${access_control_list}
    Run Keyword And Expect Error  NoAuthError  Get Node Value  ${zk}  ${ZOOKEEPER_TESTS_NODE_PATH}
    Remove Acl From Node  ${access_control_list}
    [Teardown]  Disconnect From Zookeeper  ${zk}

Test Client Without Write Grant Can Not Write To Protected Node
    [Tags]  zookeeper_sasl  zookeeper_acl  zookeeper
    ${access_control_list} =  Create Access Control List  sasl  ${ZOOKEEPER_CLIENT_USERNAME}  write=${False}  all=${False}
    Add Acl To Node  ${access_control_list}
    ${zk} =  Connect To Zookeeper  sasl  ${ZOOKEEPER_CLIENT_USERNAME}  ${ZOOKEEPER_CLIENT_PASSWORD}  ${access_control_list}
    ${data} =  Get Node Value  ${zk}  ${ZOOKEEPER_TESTS_NODE_PATH}
    Should Be Equal As Strings  ${data}  ${ACL_VALUE}
    Run Keyword And Expect Error  NoAuthError  Update Node Value  ${zk}  ${ZOOKEEPER_TESTS_NODE_PATH}  Value
    Remove Acl From Node  ${access_control_list}
    [Teardown]  Disconnect From Zookeeper  ${zk}

Test Client Without Create Grant Can Not Create Node
    [Tags]  zookeeper_sasl  zookeeper_acl  zookeeper
    ${access_control_list} =  Create Access Control List  sasl  ${ZOOKEEPER_CLIENT_USERNAME}  create=${False}  all=${False}
    ${zk} =  Connect To Zookeeper  sasl  ${ZOOKEEPER_CLIENT_USERNAME}  ${ZOOKEEPER_CLIENT_PASSWORD}  ${access_control_list}
    Run Keyword And Expect Error  NoAuthError  Create Node  ${zk}  ${ZOOKEEPER_ACL_NODE_PATH}/uncreated  ${ACL_VALUE}
    Disconnect From Zookeeper  ${zk}
    [Teardown]  Disconnect From Zookeeper  ${zk}

Test Client Without Delete Grant Can Not Delete Protected Node
    [Tags]  zookeeper_sasl  zookeeper_acl  zookeeper
    ${access_control_list} =  Create Access Control List  sasl  ${ZOOKEEPER_CLIENT_USERNAME}  delete=${False}  all=${False}
    ${zk} =  Connect To Zookeeper  sasl  ${ZOOKEEPER_CLIENT_USERNAME}  ${ZOOKEEPER_CLIENT_PASSWORD}  ${access_control_list}
    Run Keyword And Expect Error  NoAuthError  Delete Node  ${zk}  ${ZOOKEEPER_ACL_NODE_PATH}
    [Teardown]  Disconnect From Zookeeper  ${zk}

Test Client Without Admin Grant Can Not Set Permissions To Node
    [Tags]  zookeeper_sasl  zookeeper_acl  zookeeper
    ${access_control_list} =  Create Access Control List  sasl  ${ZOOKEEPER_CLIENT_USERNAME}  admin=${False}  all=${False}
    ${zk} =  Connect To Zookeeper  sasl  ${ZOOKEEPER_CLIENT_USERNAME}  ${ZOOKEEPER_CLIENT_PASSWORD}  ${access_control_list}
    ${node_acls} =  Get Acls  ${admin_zk}  ${ZOOKEEPER_TESTS_NODE_PATH}
    Append To List  ${node_acls}  ${access_control_list}
    Run Keyword And Expect Error  NoAuthError  Set Access Control Lists  ${zk}  ${ZOOKEEPER_TESTS_NODE_PATH}  ${node_acls}
    [Teardown]  Disconnect From Zookeeper  ${zk}

Test Admin Without Grants Can Access To Node
    [Tags]  zookeeper_sasl  zookeeper_acl  zookeeper
    ${access_control_list} =  Create Access Control List  sasl  ${ZOOKEEPER_CLIENT_USERNAME}
    ${acls} =  Create List  ${access_control_list}
    ${zk} =  Connect To Zookeeper  sasl  ${ZOOKEEPER_ADMIN_USERNAME}  ${ZOOKEEPER_ADMIN_PASSWORD}
    Set Access Control Lists  ${zk}  ${ZOOKEEPER_TESTS_NODE_PATH}  ${acls}
    ${data} =  Get Node Value  ${zk}  ${ZOOKEEPER_TESTS_NODE_PATH}
    Should Be Equal As Strings  ${data}  ${ACL_VALUE}
    [Teardown]  Disconnect From Zookeeper  ${zk}
    
Test Connect With Incorrect Password
    [Tags]  zookeeper_sasl  zookeeper
    ${zk} =  Connect To Zookeeper  sasl  ${ZOOKEEPER_CLIENT_USERNAME}  123
    Run Keyword And Expect Error  AuthFailedError  Node Exists  ${zk}  ${ZOOKEEPER_TESTS_NODE_PATH}
    [Teardown]  Disconnect From Zookeeper  ${zk}

Test Connect With Unknown User
    [Tags]  zookeeper_sasl  zookeeper
    ${zk} =  Connect To Zookeeper  sasl  username  password
    Run Keyword And Expect Error  AuthFailedError  Node Exists  ${zk}  ${ZOOKEEPER_TESTS_NODE_PATH}
    [Teardown]  Disconnect From Zookeeper  ${zk}