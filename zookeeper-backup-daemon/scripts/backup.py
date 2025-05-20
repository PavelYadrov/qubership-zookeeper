#!/usr/bin/python
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

import argparse
import logging
import os
import re
import sys
import ast

import requests

from process_znode_hierarchy import backup
from process_zookeeper_logs import get_snapshot_and_transaction_logs, \
    filter_and_store_transaction_logs, copy_snapshot, \
    create_directory, remove_directory_with_content, is_file_system_shared
from zookeeper_client import ZooKeeperClient

REQUEST_HEADERS = {
    'Accept': 'application/json',
    'Content-type': 'application/json'
}

ZOOKEEPER_BACKUP_TMP_DIR = '/opt/zookeeper/backup-storage/tmp'

loggingLevel = logging.DEBUG if os.getenv('ZOOKEEPER_BACKUP_DAEMON_DEBUG') else logging.INFO
logging.basicConfig(level=loggingLevel,
                    format='[%(asctime)s,%(msecs)03d][%(levelname)s][category=Backup] %(message)s',
                    datefmt='%Y-%m-%dT%H:%M:%S')


class Backup:

    def __init__(self, storage_folder):
        self._zookeeper_host = os.getenv("ZOOKEEPER_HOST")
        self._zookeeper_port = os.getenv("ZOOKEEPER_PORT")
        self._zookeeper_username = os.getenv("ZOOKEEPER_ADMIN_USERNAME")
        self._zookeeper_password = os.getenv("ZOOKEEPER_ADMIN_PASSWORD")
        if not self._zookeeper_host or not self._zookeeper_port:
            logging.error("ZooKeeper service name or port isn't specified.")
            sys.exit(1)

        self._client = ZooKeeperClient(self._zookeeper_host, self._zookeeper_port,
                                       self._zookeeper_username,
                                       self._zookeeper_password)
        self._storage_folder = storage_folder

    def transactional_backup(self):
        try:
            create_directory(ZOOKEEPER_BACKUP_TMP_DIR)
            zookeeper_servers = self.__get_zookeeper_servers()
            logging.info(f'ZooKeeper servers: {", ".join(zookeeper_servers)}.')
            zookeeper_leader = self.__find_zookeeper_leader(zookeeper_servers)
            logging.info(f'ZooKeeper leader: {zookeeper_leader}.')
            if zookeeper_leader is None:
                raise Exception(f"ZooKeeper leader isn't found in servers: {zookeeper_servers}.")

            self.__copy_logs_from_zookeeper_leader(zookeeper_leader)
            snapshot, transaction_logs = get_snapshot_and_transaction_logs(ZOOKEEPER_BACKUP_TMP_DIR)
            copy_snapshot(snapshot, self._storage_folder)
            filter_and_store_transaction_logs(transaction_logs, self._storage_folder)
        except Exception:
            logging.exception('Exception occurred during transactional backup:')
            raise
        finally:
            remove_directory_with_content(ZOOKEEPER_BACKUP_TMP_DIR)

    def hierarchical_backup(self, znodes):
        backup(self._client, self._storage_folder, znodes)

    def __get_zookeeper_servers(self):
        zk = self._client.connect_to_zookeeper()
        try:
            conf_response = self._client.execute_command(zk, "conf")
        finally:
            self._client.disconnect_from_zookeeper(zk)
        logging.debug(f'ZooKeeper config:\n {conf_response}')
        # Extract servers from line like 'server.1=zookeeper-1.zookeeper-service:2888:3888:participant;0.0.0.0:2181'
        return re.findall(r'server\.[0-9]+=([\w\-.]+):', conf_response)

    def __find_zookeeper_leader(self, servers):
        for server in servers:
            logging.info(f'Connect to ZooKeeper server: {server}.')
            zk = self._client.connect_to_zookeeper(server)
            try:
                srvr_response = self._client.execute_command(zk, "srvr")
            finally:
                self._client.disconnect_from_zookeeper(zk)
            if 'Mode: leader' in srvr_response:
                return server
        logging.error('There is no ability to find leader.')
        return None

    @staticmethod
    def __copy_logs_from_zookeeper_leader(leader):
        logging.debug(f'Start to copy logs from ZooKeeper leader: {leader}.')
        response = requests.post(f'http://{leader}:8081/store', headers=REQUEST_HEADERS)
        if response.json()['Status'] != 'Ok':
            logging.error(f'There is problem with copying data from ZooKeeper leader: {leader}, '
                          f'details: {response.json()}')
            sys.exit(1)
        logging.info('Snapshot and logs are copied successfully.')


if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument('folder')
    parser.add_argument('-mode')
    parser.add_argument('-d', '--znodes')
    args = parser.parse_args()

    backup_instance = Backup(args.folder)
    if args.mode and args.mode == 'transactional':
        if not is_file_system_shared():
            raise Exception('Configuration is not suitable to make transactional backup.')
        logging.info(f'Start transactional backup to folder: {args.folder}.')
        backup_instance.transactional_backup()
        logging.info('Transactional backup is successful.')
    else:
        znodes = ast.literal_eval(args.znodes) if args.znodes else []
        logging.info(f'Start hierarchical backup to folder: {args.folder}.')
        backup_instance.hierarchical_backup(znodes)
        logging.info('Hierarchical backup is successful.')

    logging.info('Backup completed successfully.')
