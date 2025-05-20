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

from kazoo.client import KazooClient
from kazoo.security import make_digest_acl, make_acl
from robot.api import logger
from robot.libraries.BuiltIn import BuiltIn

CA_CERT_PATH = '/tls/ca.crt'
TLS_CERT_PATH = '/tls/tls.crt'
TLS_KEY_PATH = '/tls/tls.key'


def _str2bool(v: str) -> bool:
    return v.lower() in ("yes", "true", "t", "1")


class ZookeeperLibrary(object):
    """
    ZooKeeper testing library for Robot Framework.
    """

    def __init__(self, zookeeper_os_project, zookeeper_host, zookeeper_port, zookeeper_enable_ssl):
        self._zookeeper_os_project = zookeeper_os_project
        self._zookeeper_host = zookeeper_host
        self._zookeeper_port = zookeeper_port
        self._use_ssl = _str2bool(zookeeper_enable_ssl)
        self._cafile = CA_CERT_PATH if self._use_ssl and os.path.exists(CA_CERT_PATH) else None
        self._certfile = TLS_CERT_PATH if self._use_ssl and os.path.exists(TLS_CERT_PATH) else None
        self._keyfile = TLS_KEY_PATH if self._use_ssl and os.path.exists(TLS_KEY_PATH) else None
        self.builtin = BuiltIn()

    def connect_to_zookeeper(self, scheme=None, username=None, password=None, access_control_list=None):
        """
        Connects to ZooKeeper client depending on authentication parameters.
        *Args:*\n
            _scheme_ (str, optional) - the authentication scheme, e.g. `digest` or `sasl`;\n
            _username_ (str, optional) - name of user;\n
            _password_ (str, optional) - password of user;\n
            _access_control_list_ (ACL, optional) - access control list;\n
        *Returns:*\n
            KazooClient - ZooKeeper client
        *Example:*\n
            | Connect To Zookeeper |
            | Connect To Zookeeper | digest | admin | admin | ACL(perms=31, acl_list=['ALL'], id=Id(scheme='digest', id='admin:x1nq8J5GOJVPY6zgzhtTtA9izLc=')) |
        """
        zookeeper_server = self.__get_zookeeper_server(self._zookeeper_host)
        if username and password and scheme:
            auth_data = (scheme, f'{username}:{password}')
            zk = KazooClient(hosts=zookeeper_server,
                             default_acl=[access_control_list] if access_control_list else None,
                             auth_data=[auth_data],
                             use_ssl=self._use_ssl,
                             ca=self._cafile,
                             certfile=self._certfile,
                             keyfile=self._keyfile)
        else:
            zk = KazooClient(hosts=zookeeper_server,
                             use_ssl=self._use_ssl,
                             ca=self._cafile,
                             certfile=self._certfile,
                             keyfile=self._keyfile)
        zk.start()
        logger.debug('ZooKeeper client is created and started: {}'.format(zk))
        return zk

    def connect_to_zookeeper_node(self, zookeeper_host):
        """
        Connects to ZooKeeper node.
        *Args:*\n
            _zookeeper_host_ (str) - ZooKeeper host;\n
        *Returns:*\n
            KazooClient - ZooKeeper client
        *Example:*\n
            | Connect To Zookeeper Node | zookeeper-1 |
        """
        zookeeper_server = self.__get_zookeeper_server(zookeeper_host)
        zk = KazooClient(hosts=zookeeper_server,
                         use_ssl=self._use_ssl,
                         ca=self._cafile,
                         certfile=self._certfile,
                         keyfile=self._keyfile)
        zk.start()
        logger.debug('ZooKeeper client is created and started: {}'.format(zk))
        return zk

    def __get_zookeeper_server(self, zookeeper_host):
        if self._zookeeper_os_project:
            return '{}.{}:{}'.format(zookeeper_host, self._zookeeper_os_project, self._zookeeper_port)
        else:
            return '{}:{}'.format(zookeeper_host, self._zookeeper_port)

    def disconnect_from_zookeeper(self, zk):
        """
        Disconnects from ZooKeeper client.
        *Args:*\n
            _zk_ (KazooClient) - ZooKeeper client;\n
        *Example:*\n
            | Disconnect From Zookeeper | zk |
        """
        zk.stop()
        zk.close()
        logger.debug('ZooKeeper client is stopped and closed.')

    def create_access_control_list(self, scheme, username, password=None, read=True, write=True, create=True, delete=True,
                                   admin=True, all=True):
        """
        Create a digest ACL for ZooKeeper with the given permissions.
        *Args:*\n
            _scheme_ (str) - the authorization scheme to use, e.g. `digest` or `sasl`;\n
            _username_ (str) - name of user to use for the ACL;\n
            _password_ (str, optional) - plain-text password to hash;\n
            _read_ (bool, optional) - read permission;\n
            _write_ (bool, optional) - write permission;\n
            _create_ (bool, optional) - create permission;\n
            _delete_ (bool, optional) - delete permission;\n
            _admin_ (bool, optional) - admin permission;\n
            _all_ (bool, optional) - all permission;\n
        *Returns:*\n
            ACL - access control list
        *Example:*\n
            | Create Access Control List | digest | admin | admin | read=False | all=False |
            | Create Access Control List | sasl | admin | admin | admin=False | all=False |
        """
        if scheme == "digest":
            access_control_list = make_digest_acl(username, password, read=read, write=write, create=create, delete=delete,
                                                  admin=admin, all=all)
        else:
            access_control_list = make_acl(scheme, username, read=read, write=write, create=create, delete=delete,
                                           admin=admin, all=all)
        logger.debug('Access control list is created: {}'.format(access_control_list))
        return access_control_list

    def set_access_control_lists(self, zk, node_path, acls):
        """
        Set the list of ACL for the node of the given path.
        *Args:*\n
            _zk_ (KazooClient) - ZooKeeper client;\n
            _node_path_ (str) - path of the node;\n
            _acls_ (list(ACL)) - access control list;\n
        *Example:*\n
            | Set Access Control Lists | zk | /zookeeper_acl/tests | [acl] |
        """
        zk.set_acls(node_path, acls)

    def create_node(self, zk, node_path, data=None):
        """
        Creates ZooKeeper node.
        *Args:*\n
            _zk_ (KazooClient) - ZooKeeper client;\n
            _node_path_ (str) - path of the node;\n
            _data_ (str) - initial bytes value of node;\n
        *Example:*\n
            | Create Node | zk | /zookeeper_crud | Creation data |
        """
        if data:
            zk.create(node_path, value=data.encode())
        else:
            zk.create(node_path)
        logger.debug('Node "{}" is created.'.format(node_path))

    def create_node_with_children(self, zk, node_path, children_number: int, data):
        zk.create(node_path)
        for child_id in range(children_number):
            zk.create(f"{node_path}/child{child_id}", value=data.encode())
        logger.debug(f'Node "{node_path}" with {children_number} children is created')

    def create_byte_node(self, zk, node_path, data):
        """
        Creates ZooKeeper node with byte encoding.
        *Args:*\n
            _zk_ (KazooClient) - ZooKeeper client;\n
            _node_path_ (str) - path of the node;\n
            _data_ (str) - initial bytes value of node;\n
        *Example:*\n
            | Create Byte Node | zk | /zookeeper_crud | Creation data |
        """
        logger.debug('Creating node "{}".'.format(data))
        zk.create(node_path, value=bytes(data, "cp437"))
        logger.debug('Node "{}" is created.'.format(node_path))

    def get_node_value(self, zk, node_path):
        """
        Get the value of the node.
        *Args:*\n
            _zk_ (KazooClient) - ZooKeeper client;\n
            _node_path_ (str) - path of the node;\n
        *Returns:*\n
            str - value of the node
        *Example:*\n
            | Get Node Value | zk | /zookeeper_crud/tests |
        """
        value, stat = zk.get(node_path)
        logger.debug('Node "{}" has value "{}" and stat "{}"'.format(node_path, value, stat))
        return value.decode()

    def get_byte_node_value(self, zk, node_path):
        """
        Get the value of the node with byte encoding.
        *Args:*\n
            _zk_ (KazooClient) - ZooKeeper client;\n
            _node_path_ (str) - path of the node;\n
        *Returns:*\n
            str - value of the node
        *Example:*\n
            | Get Byte Node Value | zk | /zookeeper_crud/tests |
        """
        value, stat = zk.get(node_path)
        logger.debug('Node "{}" has value "{}" and stat "{}"'.format(node_path, value, stat))
        return value.decode("cp437")

    def get_acls(self, zk, node_path):
        """
        Get the ACL of the node of the given path.
        *Args:*\n
            _zk_ (KazooClient) - ZooKeeper client;\n
            _node_path_ (str) - path of the node;\n
        *Returns:*\n
            list(ACL) - list of ACL
        *Example:*\n
            | Get Acls | zk | /zookeeper_acl/tests |
        """
        acls, stat = zk.get_acls(node_path)
        logger.debug('Node "{}" has acls "{}" and stat "{}"'.format(node_path, acls, stat))
        return acls

    def node_exists(self, zk, node_path):
        """
        Check if the node exists.
        *Args:*\n
            _zk_ (KazooClient) - ZooKeeper client;\n
            _node_path_ (str) - path of the node;\n
        *Returns:*\n
            bool - True if the node exists, False otherwise
        *Example:*\n
            | Node Exists | zk | /zookeeper_crud/tests |
        """
        stat = zk.exists(node_path)
        return stat is not None

    def update_node_value(self, zk, node_path, new_value):
        """
        Set new value of the node.
        *Args:*\n
            _zk_ (KazooClient) - ZooKeeper client;\n
            _node_path_ (str) - path of the node;\n
            _new_value_ (str) - new data value;\n
        *Example:*\n
            | Update Node Value | zk | /zookeeper_crud/tests | Modification data |
        """
        stat = zk.set(node_path, new_value.encode())
        logger.debug('Node "{}" is updated: {}'.format(node_path, stat))

    def delete_node(self, zk, node_path):
        """
        Delete the node.
        *Args:*\n
            _zk_ (KazooClient) - ZooKeeper client;\n
            _node_path_ (str) - path of the node;\n
        *Example:*\n
            | Delete Node | zk | /zookeeper_crud/tests |
        """
        zk.delete(node_path, recursive=True)
        logger.debug('Node "{}" is deleted.'.format(node_path))

    def find_minimum(self, first, second):
        """
        Find the minimum of two values.
        *Args:*\n
            _first_ (int) - first value to compare;\n
            _second_ (int) - second value to compare;\n
        *Example:*\n
            | Find Minimum | 4912 | 5120 |
        """
        return min(first, second)

    def execute_command(self, zk, cmd):
        """
        Executes command.
        *Args:*\n
            _zk_ (KazooClient) - ZooKeeper client;\n
            _cmd_ (str) - command to execute;\n
        *Example:*\n
            | Execute Command | zk | ruok |
        """
        return zk.command(cmd.encode())
    
