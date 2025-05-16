# Copyright 2024-2025 NetCracker Technology Corporation
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

import subprocess
import logging
import os
import re
from logging.handlers import RotatingFileHandler

logger = logging.getLogger(__name__)

m_tls_formatted_command = "echo mntr | openssl s_client -crlf -quiet -connect {} -cert /tls/tls.crt -key /tls/tls.key -CAfile /tls/ca.crt"
non_encrypted_formatted_command = "echo mntr | nc {} 2181"

def _str2bool(v: str) -> bool:
    return v.lower() in ("yes", "true", "t", "1")


def __configure_logging(log):
    log.setLevel(logging.DEBUG)
    formatter = logging.Formatter(fmt='[%(asctime)s,%(msecs)03d][%(levelname)s] %(message)s',
                                  datefmt='%Y-%m-%dT%H:%M:%S')
    log_handler = RotatingFileHandler(filename='/opt/zookeeper-monitoring/exec-scripts/version_info.log',
                                      maxBytes=50 * 1024,
                                      backupCount=5)
    log_handler.setFormatter(formatter)
    log_handler.setLevel(logging.DEBUG if os.getenv('ZOOKEEPER_MONITORING_SCRIPT_DEBUG') else logging.INFO)
    log.addHandler(log_handler)
    err_handler = RotatingFileHandler(filename='/opt/zookeeper-monitoring/exec-scripts/version_info.err',
                                      maxBytes=50 * 1024,
                                      backupCount=5)
    err_handler.setFormatter(formatter)
    err_handler.setLevel(logging.ERROR)
    log.addHandler(err_handler)

def _get_zk_version(server, command):
    try:
        result = subprocess.run(command.format(server), shell=True, capture_output=True, text=True)
        for line in result.stdout.strip().split('\n'):
            if line.startswith("zk_version"):
                match  = re.search(r"(\d+\.\d+\.\d+)", line)
                if match:
                    return match.group(1)
    except subprocess.CalledProcessError as e:
        logger.error(f"Failed to run mntr command on server {server}: {e}")
    return None

def _collect_metrics(servers: str):
    servers_list = servers.split(',')
    logger.debug(f'Servers are {servers_list}')
    formatted_command = m_tls_formatted_command if _str2bool(os.getenv("ZOOKEEPER_ENABLE_SSL", "false")) else non_encrypted_formatted_command
    zookeeper_server = servers_list[0].replace('\'', '').replace(' ', '')
    version = _get_zk_version(zookeeper_server, formatted_command)
    logger.info(f'Current zookeeper version={version}')
    return f'zookeeper,version={version} version=1'

def run():
    logger.info('Start collecting ZooKeeper version info...')
    servers = os.getenv('ZOOKEEPER_HOST')
    metrics = _collect_metrics(servers)
    print(metrics)

if __name__ == "__main__":
    __configure_logging(logger)
    run()