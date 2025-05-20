#!/bin/bash

# Returns true if specified file ends on new line
#$1 - file to check
newline_at_eof() {
    if [[ -z "$(tail -c 1 "$1")" ]]; then
      echo true
    else
      echo false
    fi
}

# Check that all usernames are different

#$1 - usernames comma separated list
check_usernames_different() {
  usernames_array="${1}"
  OLD_IFS=$IFS # save internal field separator
  IFS=','
  declare -A usernames_map
  for username in ${usernames_array}; do
    usernames_map["$username"]=$(( usernames_map["$username"] + 1 ))
  if [[ "${usernames_map["$username"]}" -gt "1" ]]; then
    echo "Username [${username}] is not unique.
Check ADMIN_USERNAME, CLIENT_USERNAME and ADDITIONAL_USERS parameters.
Usernames should be unique."
    exit 1
  fi
  done
  IFS=${OLD_IFS}
}

# Enables SASL configurations and creates JAAS config for ZooKeeper
enable_sasl_config() {
  export CONF_ZOOKEEPER_authProvider_1=org.apache.zookeeper.server.auth.SASLAuthenticationProvider
  export CONF_ZOOKEEPER_requireClientAuthScheme=sasl
  export SERVER_JVMFLAGS="-Dzookeeper.superUser=${ADMIN_USERNAME} -Dzookeeper.allowSaslFailedClients=false -Djava.security.auth.login.config=${ZOOKEEPER_HOME}/conf/zookeeper_jaas.conf ${SERVER_JVMFLAGS}"

  usernames_array=${ADMIN_USERNAME}
  if [[ -n ${CLIENT_USERNAME} && -n ${CLIENT_PASSWORD} ]]; then
    CLIENT_CREDENTIALS_MACROS="\n    \"user_${CLIENT_USERNAME}\"=\"${CLIENT_PASSWORD}\""
    usernames_array="${usernames_array},${CLIENT_USERNAME}"
  fi
  if [[ -n ${ADDITIONAL_USERS} ]]; then
    OLD_IFS=$IFS # save internal field separator
    IFS=','
    for USER_CREDENTIALS in ${ADDITIONAL_USERS}; do
      USER_CREDENTIALS="$(echo -e "${USER_CREDENTIALS}" | tr -d '[:space:]')"
      IFS=':'
      read -ra USER_CREDENTIALS_ARR <<< "${USER_CREDENTIALS}"
      additional_username=${USER_CREDENTIALS_ARR[0]}
      additional_password=${USER_CREDENTIALS_ARR[1]}
      if [[ -z ${additional_password} ]]; then
        echo "Password is not set for additional user [${additional_username}]. It will be skipped."
      else
        ADDITIONAL_USERS_CREDENTIALS_MACROS="${ADDITIONAL_USERS_CREDENTIALS_MACROS}\n    \"user_${additional_username}\"=\"${additional_password}\""
        usernames_array="${usernames_array},${additional_username}"
      fi
    done
    IFS=${OLD_IFS}
  fi

  check_usernames_different "${usernames_array}"

  echo -en "Server {
    org.apache.zookeeper.server.auth.DigestLoginModule required
    \"user_${ADMIN_USERNAME}\"=\"${ADMIN_PASSWORD}\"${CLIENT_CREDENTIALS_MACROS}${ADDITIONAL_USERS_CREDENTIALS_MACROS};
};"> ${ZOOKEEPER_HOME}/conf/zookeeper_jaas.conf

  if [[ ${QUORUM_AUTH_ENABLED}  == "true" ]]; then
    export CONF_ZOOKEEPER_quorum_auth_enableSasl=true
    export CONF_ZOOKEEPER_quorum_auth_learnerRequireSasl=true
    export CONF_ZOOKEEPER_quorum_auth_serverRequireSasl=true

    cat >> ${ZOOKEEPER_HOME}/conf/zookeeper_jaas.conf << EOL

QuorumServer {
       org.apache.zookeeper.server.auth.DigestLoginModule required
       "user_${ADMIN_USERNAME}"="${ADMIN_PASSWORD}";
};

QuorumLearner {
       org.apache.zookeeper.server.auth.DigestLoginModule required
       username="${ADMIN_USERNAME}"
       password="${ADMIN_PASSWORD}";
};
EOL
  else
    export CONF_ZOOKEEPER_quorum_auth_enableSasl=false
    export CONF_ZOOKEEPER_quorum_auth_learnerRequireSasl=false
    export CONF_ZOOKEEPER_quorum_auth_serverRequireSasl=false
  fi
}

# Disables SASL configurations for ZooKeeper
disable_sasl_config() {
  export CONF_ZOOKEEPER_authProvider_1=
  export CONF_ZOOKEEPER_requireClientAuthScheme=
  export CONF_ZOOKEEPER_quorum_auth_enableSasl=false
  export CONF_ZOOKEEPER_quorum_auth_learnerRequireSasl=false
  export CONF_ZOOKEEPER_quorum_auth_serverRequireSasl=false
}

# Exit immediately if a *pipeline* returns a non-zero status. (Add -x for command tracing)
set -e
if [[ "$DEBUG" == true ]]; then
  set -x
  printenv
fi

# Create missing folders
mkdir -p "${ZOOKEEPER_DATA}/dumps"
mkdir -p "${ZOOKEEPER_DATA}/configs"

if [[ -n "$JMXPORT" ]]; then
  # Docker requires extra JMX-related JVM flags beyond what Zookeeper normally uses
  JMX_EXTRA_FLAGS="-Djava.rmi.server.hostname=${JMXHOST} -Dcom.sun.management.jmxremote.rmi.port=${JMXPORT} -Dcom.sun.management.jmxremote.port=${JMXPORT}"
  if [[ -n "$JVMFLAGS" ]]; then
    export JVMFLAGS="${JMX_EXTRA_FLAGS} ${JVMFLAGS} "
  else
    export JVMFLAGS="${JMX_EXTRA_FLAGS} "
  fi
fi

DIAGNOSTIC_OPTS="-XX:+ExitOnOutOfMemoryError"
: ${NC_DIAGNOSTIC_MODE:="off"}
#DISABLE UNTIL CDC RELEASE
#. ${NC_DIAGNOSTIC_FOLDER}/nc-diagnostic-bootstrap.sh
if [[ -z "$X_JAVA_ARGS" ]]; then
  DIAGNOSTIC_OPTS="${DIAGNOSTIC_OPTS} -XX:+HeapDumpOnOutOfMemoryError -XX:HeapDumpPath=${ZOOKEEPER_DATA}/dumps/"
else
  DIAGNOSTIC_OPTS="${DIAGNOSTIC_OPTS} ${X_JAVA_ARGS}"
fi
export SERVER_JVMFLAGS="${DIAGNOSTIC_OPTS} ${SERVER_JVMFLAGS}"

function apply_esc_configuration() {
  echo 'for sig in $SIGNALS_TO_RETHROW; do trap "rethrow_java_handler $sig" $sig 2>&1 > /dev/null; done' > ${NC_DIAGNOSTIC_FOLDER}/java
  echo /etc/alternatives/java "\${X_JAVA_ARGS}" "\$@" '&' >> ${NC_DIAGNOSTIC_FOLDER}/java
  echo 'java_pid=$!' >> ${NC_DIAGNOSTIC_FOLDER}/java
  echo 'wait "$java_pid"' >> ${NC_DIAGNOSTIC_FOLDER}/java
}

must_send_crash_dump=false
# When java process ends due to signal or System.exit , we can collect crash
# dumps (is such appeared) and send them to remote location for thorough diagnostic.
if [[ "$(type -t send_crash_dump)" = "function" ]]; then
  function after_java() {
    send_crash_dump
  }
  export -f after_java
  must_send_crash_dump=true
fi

pid=0
javaRetCode=0
java_pid=0

# Subcommand, which this entrypoint runs, could be itslef a script in case
# if service needs some custom actions done before java ran.
#
# We need to rethrow signals both to wrapped subcommand and java itself
function rethrow_java_handler() {
  echo "Caught $1 sig in /usr/bin/java"
  if [[ "$java_pid" -ne 0 ]]; then
    echo "Signaling to java"
    kill -"$1" "$java_pid"
    wait "$java_pid"; javaRetCode=$?
  fi
  echo "Java signaled with $1, exit code $javaRetCode"
  if $must_send_crash_dump && [[ "$1" == "SIGTERM" ]]; then
    after_java
  fi
  exit ${javaRetCode}
}
export -f rethrow_java_handler

function rethrow_handler() {
  echo "Caught $1 sig in entrypoint"
  local subRetCode=0
  if [[ "$pid" -ne 0 ]]; then
    if [[ "$1" == "SIGTERM" ]]; then
      /bin/sleep 10
    fi
    echo "Signaling to subcommand"
    kill -"$1" "$pid"
    wait "$pid"; subRetCode=$?
  fi
  echo "Subcommand signaled with $1, exit code $subRetCode"
  exit ${subRetCode}
}

# See full current list in http://man7.org/linux/man-pages/man7/signal.7.html
export SIGNALS_TO_RETHROW="
SIGHUP
SIGINT
SIGQUIT
SIGILL
SIGABRT
SIGFPE
SIGSEGV
SIGPIPE
SIGALRM
SIGTERM
SIGUSR1
SIGUSR2
SIGCONT
SIGSTOP
SIGTSTP
SIGTTIN
SIGTTOU
SIGBUS
SIGPROF
SIGSYS
SIGTRAP
SIGURG
SIGVTALRM
SIGXCPU
SIGXFSZ
SIGSTKFLT
SIGIO
SIGPWR
SIGWINCH
"

if [[ -n "$JOLOKIA_PORT" ]]; then
  export SERVER_JVMFLAGS="${SERVER_JVMFLAGS} -javaagent:/opt/zookeeper/lib/jolokia-jvm-1.7.1.jar=port=$JOLOKIA_PORT,host=0.0.0.0"
  if [[ -n "$CLIENT_USERNAME" && -n "${CLIENT_PASSWORD}" ]]; then
    export SERVER_JVMFLAGS="${SERVER_JVMFLAGS},user=${CLIENT_USERNAME},password=${CLIENT_PASSWORD}"; fi
  unset JOLOKIA_PORT
fi

export SERVER_JVMFLAGS="${SERVER_JVMFLAGS} -javaagent:/opt/zookeeper/lib/jmx_prometheus_javaagent-1.1.0.jar=8080:${ZOOKEEPER_HOME}/conf/jmx-exporter-config.yaml"

if [[ -n "$HEAP_OPTS" ]]; then
  export SERVER_JVMFLAGS="${HEAP_OPTS} ${SERVER_JVMFLAGS}"
  unset HEAP_OPTS
fi

# WA for https://issues.apache.org/jira/browse/ZOOKEEPER-2528
if [[ -d ${ZOOKEEPER_DATA}/version-2 ]]; then
  for empty_log in $(find ${ZOOKEEPER_DATA}/version-2 -type f -empty -name "log.*"); do
    echo "[$(date +'%Y-%m-%dT%H:%M:%S,000')][ERROR] Zookeeper data directory contains empty transaction log $empty_log. Try to remove it."
    rm ${empty_log}
    echo "[$(date +'%Y-%m-%dT%H:%M:%S,001')][INFO] Empty transaction log file $empty_log has been removed."
  done
fi

if [[ -n ${ADMIN_USERNAME} && -n ${ADMIN_PASSWORD} ]]; then
  enable_sasl_config
else
  disable_sasl_config
fi

if [[ "${ENABLE_SSL}" == "true" ]]; then
  echo "Configuring TLS..."
  zookeeper_tls_dir=${ZOOKEEPER_HOME}/tls
  SSL_KEY_LOCATION=${zookeeper_tls_dir}/tls.key
  SSL_CERTIFICATE_LOCATION=${zookeeper_tls_dir}/tls.crt
  SSL_CA_LOCATION=${zookeeper_tls_dir}/ca.crt
  zookeeper_tls_ks_dir=${ZOOKEEPER_HOME}/tls-ks
  SSL_KEYSTORE_LOCATION=${zookeeper_tls_ks_dir}/zookeeper.keystore.jks
  SSL_TRUSTSTORE_LOCATION=${zookeeper_tls_ks_dir}/zookeeper.truststore.jks

  if [[ -f ${SSL_KEY_LOCATION} && -f ${SSL_CERTIFICATE_LOCATION} && -f ${SSL_CA_LOCATION} ]]; then
    mkdir -p ${zookeeper_tls_ks_dir}
    openssl pkcs12 -export -in ${SSL_CERTIFICATE_LOCATION} -inkey ${SSL_KEY_LOCATION} -out ${zookeeper_tls_ks_dir}/zookeeper.keystore.p12 -passout pass:changeit
    keytool -importkeystore -destkeystore ${SSL_KEYSTORE_LOCATION} -deststorepass changeit -srcstoretype PKCS12 -srckeystore ${zookeeper_tls_ks_dir}/zookeeper.keystore.p12 -srcstorepass changeit
    keytool -import -trustcacerts -keystore ${SSL_KEYSTORE_LOCATION} -storepass changeit -noprompt -alias ca-cert -file ${SSL_CA_LOCATION}
    keytool -import -trustcacerts -keystore ${SSL_TRUSTSTORE_LOCATION} -storepass changeit -noprompt -alias ca -file ${SSL_CA_LOCATION}

    export CONF_ZOOKEEPER_serverCnxnFactory=org.apache.zookeeper.server.NettyServerCnxnFactory
    export CONF_ZOOKEEPER_authProvider_x509=org.apache.zookeeper.server.auth.X509AuthenticationProvider
    export CONF_ZOOKEEPER_ssl_keyStore_location=${SSL_KEYSTORE_LOCATION}
    export CONF_ZOOKEEPER_ssl_keyStore_password=changeit
    export CONF_ZOOKEEPER_ssl_trustStore_location=${SSL_TRUSTSTORE_LOCATION}
    export CONF_ZOOKEEPER_ssl_trustStore_password=changeit
    export CONF_ZOOKEEPER_ssl_quorum_keyStore_location=${SSL_KEYSTORE_LOCATION}
    export CONF_ZOOKEEPER_ssl_quorum_keyStore_password=changeit
    export CONF_ZOOKEEPER_ssl_quorum_trustStore_location=${SSL_TRUSTSTORE_LOCATION}
    export CONF_ZOOKEEPER_ssl_quorum_trustStore_password=changeit
    export CONF_ZOOKEEPER_ssl_quorum_hostnameVerification=false
    if [[ "${ENABLE_2WAY_SSL}" != "true" ]]; then
      export CONF_ZOOKEEPER_ssl_clientAuth=none
    fi
    if [[ -n "${SSL_CIPHER_SUITES}" ]]; then
      export CONF_ZOOKEEPER_ssl_ciphersuites=${SSL_CIPHER_SUITES}
    fi
    export CONF_ZOOKEEPER_secureClientPort=2181
    export CONF_ZOOKEEPER_sslQuorum=true
    echo "TLS configuration is applied"
  else
    echo "TLS certificates must be provided."
    exit 1
  fi
fi

function start() {
  #
  # Process the logging-related environment variables. Zookeeper's log configuration allows *some* variables to be
  # set via environment variables, and more via system properties (e.g., "-Dzookeeper.console.threshold=INFO").
  # However, in the interest of keeping things straightforward and in the spirit of the immutable image,
  # we don't use these and instead directly modify the Log4J configuration file (replacing the variables).
  #
  if [[ -z "$LOG_LEVEL" ]]; then
    LOG_LEVEL="INFO"
  fi
  sed -i -r -e "s|value=\"INFO\"|value=\"${LOG_LEVEL}\"|g" ${ZOOKEEPER_HOME}/conf/logback.xml
  sed -i -r -e "s|level=\"INFO\"|level=\"${LOG_LEVEL}\"|g" ${ZOOKEEPER_HOME}/conf/logback.xml
  #
  # Configure cluster settings
  #
  if [[ -z "$SERVER_NAME" ]]; then
    SERVER_NAME="zookeeper"
  fi
  if [[ -z "$SERVER_ID" ]]; then
    SERVER_ID="1"
  fi
  if [[ -z "$SERVER_COUNT" ]]; then
    SERVER_COUNT=1
  fi
  if [[ -z "$RECONFIG_ENABLED" ]]; then
    RECONFIG_ENABLED=false
  fi
  echo "Starting up ${SERVER_ID} of ${SERVER_COUNT}"
  #
  # Generate the dynamic configuration only if it doesn't exist
  #
  if [[ ! -f "$ZOOKEEPER_HOME/conf/zoo.cfg.dynamic" ]]; then
    #
    # Append the server addresses to the configuration file
    #
    export IS_CLOUD_ENABLED=${IS_CLOUD_ENABLED:=true}

    if [[ ${IS_CLOUD_ENABLED} == "true" ]]; then
      if [[ ${DR_JOINT_CLUSTER} == "true" ]]; then
        echo "[$(date +'%Y-%m-%dT%H:%M:%S,000')][INFO] Joint DR ZooKeeper cluster is enabled."
        if [[ "$RECONFIG_ENABLED" == "true" ]]; then
          echo "[$(date +'%Y-%m-%dT%H:%M:%S,001')][WARN] RECONFIG_ENABLED (with storing config on disk) is not compatibly with DR_JOINT_CLUSTER."
          RECONFIG_ENABLED=false
        fi
        export CONF_ZOOKEEPER_reconfigEnabled=true

        ${ZOOKEEPER_HOME}/bin/dr_reconfig.sh config ${DR_ACTIVE_SIDE}
      else
        if [[ -n "$SERVER_DOMAIN" ]]; then
          SERVER_DOMAIN=".$SERVER_DOMAIN"
        fi
        if [[ -n "$SERVER_NAMESPACE" ]]; then
          SERVER_NAMESPACE=".$SERVER_NAMESPACE"
        fi
        CLIENT_PORT=2181
        if [[ "${ENABLE_SSL}" == "true" ]]; then
          CLIENT_PORT=""
          if [[ "${ALLOW_NONENCRYPTED_ACCESS}" == "true" ]]; then
            CLIENT_PORT=2182
          fi
        fi
        echo "# Server List" >> ${ZOOKEEPER_HOME}/conf/zoo.cfg.dynamic
        for (( order_number=1; order_number <= ${SERVER_COUNT}; order_number++ )); do
          echo "server.$order_number=$SERVER_NAME-$order_number$SERVER_DOMAIN$SERVER_NAMESPACE:2888:3888:participant;$CLIENT_PORT" >> ${ZOOKEEPER_HOME}/conf/zoo.cfg.dynamic
        done
      fi
    else
      export QUORUM_SERVERS=${QUORUM_SERVERS:="$SERVER_ID@0.0.0.0:$ZOOKEEPER_FOLLOWERS_PORT:$ZOOKEEPER_ELECTION_PORT:participant"}

      declare -a quorum_servers_array
      IFS=',' read -r -a quorum_servers_array <<< ${QUORUM_SERVERS}

      if [[ ${#quorum_servers_array[@]} != "$SERVER_COUNT" ]]; then
        echo ERROR: "The number of QUORUM_SERVERS must be equal to SERVER_COUNT"
        exit 1
      fi

      for (( i = 1; i <= "$SERVER_COUNT"; i++ )); do
        declare -a quorum_server
        IFS='@' read -r -a quorum_server <<< ${quorum_servers_array[i-1]}

        if [[ ${#quorum_server[@]} != 2 || -z ${quorum_server[0]} ]]; then
            echo ERROR: "Incorrect format of quorum server string (<server_id>@<zookeeper_host>:<follower_port>:<election_post>): ${quorum_servers_array[i-1]}."
            exit 1
        fi

        echo "server.${quorum_server[0]}=${quorum_server[1]}" >> ${ZOOKEEPER_HOME}/conf/zoo.cfg.dynamic
      done
    fi
  fi

  if [[ -f ${ZOOKEEPER_DATA}/myid ]]; then
      echo "ID configuration exists: ${ZOOKEEPER_DATA}/myid. Rewriting it."
  fi

  if [[ ${SCAN_FILE_SYSTEM} == "true" ]]; then
      echo "Reading file system before starting."
      ls -Ra "${ZOOKEEPER_DATA}" > /dev/null 2>&1
  fi

  #
  # Persists the ID of the current instance of Zookeeper in the 'myid' file
  #
  echo ${SERVER_ID} > ${ZOOKEEPER_DATA}/myid

  # Copy backup from shared folder to internal structure
  if [[ -d "$ZOOKEEPER_RECOVERY_DIR" && "$(ls -A "$ZOOKEEPER_RECOVERY_DIR")" ]]; then
    rm -rfv ${ZOOKEEPER_BACKUP_SOURCE_DIR}/*
    # Dot '.' allows to copy all files and folders included hidden ones
    cp -rp ${ZOOKEEPER_RECOVERY_DIR}/. ${ZOOKEEPER_BACKUP_SOURCE_DIR}
    echo "Files are copied from $ZOOKEEPER_RECOVERY_DIR to $ZOOKEEPER_BACKUP_SOURCE_DIR"
  fi

  if [[ "$AUDIT_ENABLED" == true ]]; then
    export CONF_ZOOKEEPER_audit_enable=true
  fi

  # Now start the Zookeeper server
  if [[ "$RECONFIG_ENABLED" == true ]]; then
    export ZOOCFG="${ZOOKEEPER_DATA}/configs/zoo.cfg"
    # Copy the configuration only if it doesn't exist
    if [[ ! -f "$ZOOCFG" ]]; then
      cp ${ZOOKEEPER_HOME}/conf/zoo.cfg ${ZOOCFG}
    fi
    export CONF_ZOOKEEPER_reconfigEnabled=true
  else
    export ZOOCFG="${ZOOKEEPER_HOME}/conf/zoo.cfg"
    rm -fr ${ZOOKEEPER_DATA}/configs/*
  fi

  # Add missing EOF at the end of the config file
  if [[ $(newline_at_eof ${ZOOCFG}) == false ]]; then
    echo "" >> ${ZOOCFG}
  fi

  for VAR in `env | grep ^CONF_ZOOKEEPER_` ; do
    env_var=`echo "$VAR" | sed -r "s/(.*)=.*/\1/g"`
    prop_name=`echo "$VAR" | sed -r "s/^CONF_ZOOKEEPER_(.*)=.*/\1/g" | tr _ .`
    if egrep -q "(^|^#)$prop_name=" ${ZOOCFG}; then
      # Note that no config names or values may contain an '@' char
      sed -r -i "s@(^|^#)($prop_name)=(.*)@\2=${!env_var}@g" ${ZOOCFG}
    else
      #echo "Adding property $prop_name=${!env_var}"
      echo "$prop_name=${!env_var}" >> ${ZOOCFG}
    fi
  done

  echo "zoo.cfg file:"
  cat ${ZOOCFG}
  echo -e
  echo "generated zoo.cfg.dynamic file:"
  cat ${ZOOKEEPER_HOME}/conf/zoo.cfg.dynamic

  exec ${ZOOKEEPER_HOME}/bin/zkServer.sh start-foreground ${ZOOCFG}
}

# Process some known arguments to run Zookeeper
case $1 in
  start)
    if [[ -z "$X_JAVA_ARGS" ]]; then
      start
    else
      # Another distributed profiler (execution statistics collector) also is supported
#      apply_esc_configuration

      # We don't want to mess with shell signal handling in terminal mode.
      # Otherwise we need to rethrow signals to service to terminate it gracefully
      # in case of need, while also executing post-mortem if available.

      for sig in ${SIGNALS_TO_RETHROW}; do trap "rethrow_handler $sig" ${sig} 2>&1 > /dev/null; done
      start &
      pid="$!"
      wait "$pid"; javaRetCode=$?
      echo "Java process ended with return code ${javaRetCode}"
      if $must_send_crash_dump; then
        after_java
      fi
      exit ${javaRetCode}
    fi
    ;;
esac

# Otherwise just run the specified command
exec "$@"
