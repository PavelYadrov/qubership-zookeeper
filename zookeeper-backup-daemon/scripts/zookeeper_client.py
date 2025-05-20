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

import os
import logging

from kazoo.client import KazooClient


def _str2bool(v: str) -> bool:
    return v.lower() in ("yes", "true", "t", "1")


class ZooKeeperClient:

    def __init__(self, zookeeper_host, zookeeper_port, zookeeper_username, zookeeper_password):
        self._zookeeper_host = zookeeper_host
        self._zookeeper_port = zookeeper_port
        self._zookeeper_username = zookeeper_username
        self._zookeeper_password = zookeeper_password
        self._ca = None
        self._certfile = None
        self._keyfile = None
        self._use_ssl = False
        if _str2bool(os.getenv("ZOOKEEPER_ENABLE_SSL", "false")):
            self._ca = "/tls/ca.crt"
            self._certfile = "/tls/tls.crt"
            self._keyfile = "/tls/tls.key"
            self._use_ssl = True

    def connect_to_zookeeper(self, zookeeper_host: str =None):
        """
        Connects to ZooKeeper client.
        *Args:*\n
            _zookeeper_host_  ZooKeeper host (optional);\n
        *Returns:*\n
            KazooClient - ZooKeeper client
        """
        if not zookeeper_host:
            zookeeper_host = self._zookeeper_host
        zookeeper_server = f'{zookeeper_host}:{self._zookeeper_port}'

        if self._zookeeper_username and self._zookeeper_password:
            zk = KazooClient(hosts=zookeeper_server,
                             keyfile=self._keyfile,
                             certfile=self._certfile,
                             ca=self._ca,
                             use_ssl=self._use_ssl,
                             auth_data=[('sasl', f'{self._zookeeper_username}:{self._zookeeper_password}')])
        else:
            zk = KazooClient(hosts=zookeeper_server,
                             keyfile=self._keyfile,
                             certfile=self._certfile,
                             ca=self._ca,
                             use_ssl=self._use_ssl)
        zk.start()
        logging.debug(f"ZooKeeper client '{zk}' is created and started.")
        return zk

    @staticmethod
    def execute_command(zk, command: str):
        """
        Executes command.
        *Args:*\n
            _zk_ (KazooClient) - ZooKeeper client;\n
            _command_ (str) - command to execute;\n
        *Returns:*\n
            str - command execution result
        """
        return zk.command(command.encode())

    @staticmethod
    def disconnect_from_zookeeper(zk):
        """
        Disconnects from ZooKeeper client.
        *Args:*\n
            _zk_ (KazooClient) - ZooKeeper client;\n
        """
        zk.stop()
        zk.close()
        logging.debug(f"ZooKeeper client '{zk}' is stopped and closed.")
