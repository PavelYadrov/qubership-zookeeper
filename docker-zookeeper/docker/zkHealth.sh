#!/bin/bash

need_to_run_backup_server() {
  local is_process_present=$(ps aux | grep "/zookeeper-assistant -c backup" | grep ${ZOOKEEPER_BACKUP_SOURCE_DIR})
  echo "$is_process_present"
  if [[ -z "${is_process_present}" ]]; then
    return 0
  fi
  return 1
}

run_backup_server() {
  nohup /zookeeper-assistant -c backup -s ${ZOOKEEPER_BACKUP_SOURCE_DIR} -d ${ZOOKEEPER_BACKUP_DESTINATION_DIR} &> output_backup_server.log &
  while true; do
    curl -s -X GET "http://localhost:8081/" 2>server_errors.log
    cat server_errors.log | grep -o "Failed to connect to localhost port 8081"
    if [[ $? -eq 0 ]]; then
      sleep 1s
    else
      echo "" > server_errors.log
      break
    fi
  done
}

liveness() {
  if [[ ! -d ${ZOOKEEPER_DATA}/logs ]];then
    mkdir ${ZOOKEEPER_DATA}/logs
  fi

  local output="Check health for localhost. time: $(date)"

  local success=false
  for attempt in {1..5}; do
    if [[ "${ENABLE_SSL}" == "true" ]]; then
      if [[ "${ENABLE_2WAY_SSL}" == "true" ]]; then
        local result=$(echo "ruok" | openssl s_client -crlf -quiet -connect localhost:2181 -cert /opt/zookeeper/tls/tls.crt -key /opt/zookeeper/tls/tls.key -CAfile /opt/zookeeper/tls/ca.crt 2>/dev/null)
      else
        local result=$(echo "ruok" | openssl s_client -crlf -quiet -connect localhost:2181 -CAfile /opt/zookeeper/tls/ca.crt 2>/dev/null)
      fi
    else
      local result=$(echo ruok | nc -w 2 -q 1 localhost 2181)
    fi
    output="${output}\n- ruok attempt ${attempt} - time: $(date) - result: $result"
    echo ${result} | grep "imok"
    if [[ $? -eq 0 ]]; then
      success=true
      break
    fi
  done

  if [[ ${success} == true ]]; then
    for attempt in {1..5}; do
      if [[ "${ENABLE_SSL}" == "true" ]]; then
        if [[ "${ENABLE_2WAY_SSL}" == "true" ]]; then
          local result=$(echo "srvr" | openssl s_client -crlf -quiet -connect localhost:2181 -cert /opt/zookeeper/tls/tls.crt -key /opt/zookeeper/tls/tls.key -CAfile /opt/zookeeper/tls/ca.crt 2>/dev/null)
        else
          local result=$(echo "srvr" | openssl s_client -crlf -quiet -connect localhost:2181 -CAfile /opt/zookeeper/tls/ca.crt 2>/dev/null)
        fi
      else
        local result=$(echo srvr | nc -w 2 -q 1 localhost 2181)
      fi
      output="${output}\n- srvr attempt ${attempt} - time: $(date) - result: $result"
      echo ${result} | grep -P "Mode: (follower|leader|standalone)"
      if [[ $? -eq 0 ]]; then
        return 0
      fi
    done
  fi

  output="${output}\n- failed"
  echo -e "${output}" > ${ZOOKEEPER_DATA}/logs/health.log
  return 1
}

readiness() {
  if [[ -n ${ADMIN_USERNAME} && -n ${ADMIN_PASSWORD} ]]; then
    /zookeeper-assistant -c health -u "${ADMIN_USERNAME}:${ADMIN_PASSWORD}"
  else
    /zookeeper-assistant -c health
  fi
  if [[ $? -eq 0 ]]; then
    return 0
  fi
  echo -e  "ZooKeeper connection failed. time: $(date) - result: ZooKeeper Quorum does not work" >> ${ZOOKEEPER_DATA}/logs/health.log
  return 1
}

if need_to_run_backup_server; then
  run_backup_server
fi

case $1 in
health)
    liveness
    exit $?
    ;;
liveness-probe)
    liveness
    exit $?
    ;;
readiness-probe)
    readiness
    exit $?
    ;;
esac