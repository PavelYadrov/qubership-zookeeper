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

import logging
import os
import shlex
import subprocess
from logging.handlers import RotatingFileHandler

logger = logging.getLogger(__name__)

m_tls_formatted_command = "openssl s_client -crlf -quiet -connect {} -cert /tls/tls.crt -key /tls/tls.key -CAfile /tls/ca.crt"
non_encrypted_formatted_command = "nc {}"

def __configure_logging(log):
    log.setLevel(logging.DEBUG)
    formatter = logging.Formatter(fmt='[%(asctime)s,%(msecs)03d][%(levelname)s] %(message)s',
                                  datefmt='%Y-%m-%dT%H:%M:%S')
    log_handler = RotatingFileHandler(filename='/opt/zookeeper-monitoring/exec-scripts/health_metric.log',
                                      maxBytes=50 * 1024,
                                      backupCount=5)
    log_handler.setFormatter(formatter)
    log_handler.setLevel(logging.DEBUG if os.getenv('ZOOKEEPER_MONITORING_SCRIPT_DEBUG') else logging.INFO)
    log.addHandler(log_handler)
    err_handler = RotatingFileHandler(filename='/opt/zookeeper-monitoring/exec-scripts/health_metric.err',
                                      maxBytes=50 * 1024,
                                      backupCount=5)
    err_handler.setFormatter(formatter)
    err_handler.setLevel(logging.ERROR)
    log.addHandler(err_handler)


def _get_number_of_alive_nodes(servers: list, formatted_command: str) -> int:
    """
    Receives number of alive nodes in ZooKeeper cluster.
    :param servers: list of
    :return: number of alive nodes
    """
    nodes_number = 0
    for server in servers:
        zookeeper_server = server.replace('\'', '').replace(' ', '')
        try:
            # sends request through 'subprocess', because Kazoo doesn't work with unavailable ZooKeeper servers
            args = shlex.split(formatted_command.format(zookeeper_server))
            logger.debug(f'Arguments for request are {args}')
            process = subprocess.Popen(args, stdin=subprocess.PIPE, stdout=subprocess.PIPE, stderr=subprocess.DEVNULL)
            response, error = process.communicate(b'ruok', 4)
            logger.debug(f'Response from request is {response}')
            if response == b'imok':
                nodes_number += 1
        except Exception:
            logger.exception(f'There are problems connecting to the server {zookeeper_server}')
    return nodes_number


def _get_status_code(alive_nodes: int, total_nodes: int):
    """
    Receives status of ZooKeeper cluster.
    :param alive_nodes: number of alive nodes
    :param total_nodes: total number of nodes
    :return: 0 - if all nodes are alive; 5 - if some nodes failed; 10 - if all nodes failed
    """
    status_code = 10
    if alive_nodes == 0:
        status_code = 10
    elif alive_nodes < total_nodes:
        status_code = 5
    elif alive_nodes == total_nodes:
        status_code = 0
    return status_code


def _str2bool(v: str) -> bool:
    return v.lower() in ("yes", "true", "t", "1")


ZK_SERVER_STATE_NON_TLS_COMMAND = "echo mntr | nc -w 5 {} | grep zk_server_state"
ZK_SERVER_STATE_TLS_COMMAND = "echo mntr | openssl s_client -crlf -quiet -connect {} -cert /tls/tls.crt -key /tls/tls.key -CAfile /tls/ca.crt | grep zk_server_state"

def get_server_state(server_address):
    try:
        use_tls = _str2bool(os.getenv("ZOOKEEPER_ENABLE_SSL", "false"))
        formatted_command = ZK_SERVER_STATE_TLS_COMMAND if use_tls else ZK_SERVER_STATE_NON_TLS_COMMAND
        
        zookeeper_server = server_address.replace('\'', '').replace(' ', '')
        command = formatted_command.format(zookeeper_server)
        logger.info(f"Executing command: {command}")
        
        process = subprocess.Popen(command, stdout=subprocess.PIPE, stderr=subprocess.PIPE, text=True, shell=True)
        stdout, stderr = process.communicate(timeout=10)
        
        logger.debug(f"Command return code: {process.returncode}")
        logger.debug(f"Command stdout: {stdout}")
        logger.debug(f"Command stderr: {stderr}")
        
        if process.returncode == 0 and stdout:
            # Extract only the state part (leader, follower, etc.)
            parts = stdout.split()
            if len(parts) >= 2:
                state = parts[-1].strip()
                logger.info(f"Server {zookeeper_server} state: {state}")
                return state
            else:
                logger.warning(f"Unexpected output format from {zookeeper_server}: {stdout}")
        else:
            logger.warning(f"Failed to retrieve state information from {zookeeper_server} (return code: {process.returncode})")
    except subprocess.TimeoutExpired:
        logger.warning(f"Timeout while connecting to {zookeeper_server}")
    except Exception as e:
        logger.exception(f"Error retrieving state information from {zookeeper_server}: {e}")

    logger.warning(f"Returning 'NA' for server {zookeeper_server}")
    return "NA"

def get_leader_node(zookeeper_hosts):
    logger.info(f"Checking leader info for hosts: {zookeeper_hosts}")
    results = []
    for host in zookeeper_hosts:
        logger.info(f"Checking host: {host}")
        hostname = host.strip("'").split(':')[0]  # Extract hostname without port
        server_info = get_server_state(host)
        results.append((hostname, server_info))
        logger.info(f"Server info for {hostname}: {server_info}")
    
    leaders = [host for host, state in results if state == "leader"]
    followers = [host for host, state in results if state == "follower"]
    
    logger.info(f"Leaders: {leaders}")
    logger.info(f"Followers: {followers}")
    
    if len(leaders) == 1:
        logger.info(f"Unique leader found: {leaders[0]}")
        return leaders[0]  # Return just the hostname without port
    elif len(leaders) > 1:
        logger.warning(f"Multiple leaders found: {leaders}. This is unexpected.")
    else:
        logger.warning("No leader found among all hosts")
    
    return "NA"

def _collect_metrics(servers: str):
    servers_list = servers.split(',')
    logger.debug(f'Servers are {servers_list}')

    formatted_command = m_tls_formatted_command if _str2bool(os.getenv("ZOOKEEPER_ENABLE_SSL", "false")) \
        else non_encrypted_formatted_command

    alive_nodes = _get_number_of_alive_nodes(servers_list, formatted_command)
    total_nodes = len(servers_list)
    status_code = _get_status_code(alive_nodes, total_nodes)
    logger.info(f'Current status code is {status_code}, {alive_nodes}/{total_nodes} nodes are alive.')

    leader_name = get_leader_node(servers_list)

    return f'zookeeper,leader={leader_name} status_code={status_code},alive_nodes={alive_nodes},total_nodes={total_nodes}'

def run():
    logger.info('Start collecting ZooKeeper health metrics...')
    servers = os.getenv('ZOOKEEPER_HOST')
    message = _collect_metrics(servers)
    print(message)


if __name__ == "__main__":
    __configure_logging(logger)
    run()  