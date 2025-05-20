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
import ast
import logging
import os
import sys
from os.path import join, isfile

from process_znode_hierarchy import restore
from process_zookeeper_logs import copy_zookeeper_logs, \
    create_directory, remove_directory_with_content, is_file_system_shared
from zookeeper_client import ZooKeeperClient

ZOOKEEPER_RESTORE_TMP_DIR = '/opt/zookeeper/backup-storage/recover'

loggingLevel = logging.DEBUG if os.getenv('ZOOKEEPER_BACKUP_DAEMON_DEBUG') else logging.INFO
logging.basicConfig(level=loggingLevel,
                    format='[%(asctime)s,%(msecs)03d][%(levelname)s][category=Restore] %(message)s',
                    datefmt='%Y-%m-%dT%H:%M:%S')


class Restore:

    def __init__(self, storage_folder):
        self._project = os.getenv("NAMESPACE")
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

    def determine_mode(self):
        for file_name in os.listdir(self._storage_folder):
            file_path = join(self._storage_folder, file_name)
            if isfile(file_path):
                if 'snapshot.' in file_name or 'log.' in file_name:
                    return 'transactional'
        return 'hierarchical'

    def transactional_recovery(self):
        try:
            create_directory(ZOOKEEPER_RESTORE_TMP_DIR)
            copy_zookeeper_logs(self._storage_folder, ZOOKEEPER_RESTORE_TMP_DIR)
            from PlatformLibrary import PlatformLibrary
            is_managed_by_operator: str = "true"
            if os.getenv("MANAGED_BY_OPERATOR") and os.getenv("MANAGED_BY_OPERATOR").lower() == "false":
                is_managed_by_operator = "false"
            client = PlatformLibrary(managed_by_operator=is_managed_by_operator)
            client.scale_down_deployment_entities_by_service_name(self._zookeeper_host,
                                                                  self._project,
                                                                  with_check=True)
            client.scale_up_deployment_entities_by_service_name(self._zookeeper_host,
                                                                self._project,
                                                                with_check=True)
        except Exception:
            logging.exception('Exception occurred during transactional recovery:')
            raise
        finally:
            remove_directory_with_content(ZOOKEEPER_RESTORE_TMP_DIR)

    def hierarchical_recovery(self, znodes):
        restore(self._client, znodes, self._storage_folder)


if __name__ == "__main__":
    parser = argparse.ArgumentParser()
    parser.add_argument('folder')
    parser.add_argument('-d', '--znodes')
    args = parser.parse_args()

    restore_instance = Restore(args.folder)
    determined_mode = restore_instance.determine_mode()
    logging.info(f'Start {determined_mode} recovery from folder: {args.folder}.')
    if determined_mode == 'transactional':
        if not is_file_system_shared():
            raise Exception('Configuration is not suitable to restore from transactional backup.')
        restore_instance.transactional_recovery()
        logging.info('Transactional recovery is successful.')
    else:
        znodes = ast.literal_eval(args.znodes) if args.znodes else []
        restore_instance.hierarchical_recovery(znodes)
        logging.info(f"Hierarchical recovery of znodes '{znodes}' is successful.")

    logging.info('Recovery completed successfully.')
