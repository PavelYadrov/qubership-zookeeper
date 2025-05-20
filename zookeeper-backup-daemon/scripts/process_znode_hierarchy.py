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

import logging
import re
import shutil
import zipfile
import os


def backup(client, storage_folder, znodes):
    zk = client.connect_to_zookeeper()
    znodes_folder = f'{storage_folder}/znodes'
    try:
        os.makedirs(znodes_folder)
        if not znodes:
            visit_nodes(zk, "/", znodes_folder)
        else:
            for znode in znodes:
                visit_nodes(zk, znode, f'{znodes_folder}/')
        shutil.make_archive(znodes_folder, 'zip', znodes_folder)
    finally:
        client.disconnect_from_zookeeper(zk)
        if os.path.isdir(znodes_folder):
            shutil.rmtree(znodes_folder)


def visit_nodes(zk, path, storage_folder):
    logging.debug(f'On node {path}.')
    children = zk.get_children(path)
    logging.debug(f"Node '{path}' has {len(children)} children.")
    node_value = zk.get(path)
    value = node_value[0]
    data = node_value[1]
    logging.debug(f'Node value is [{value}], [{data}].')
    ephemeral_owner = data.ephemeralOwner
    logging.debug(f'Owner is {ephemeral_owner}.')
    if not ephemeral_owner:
        store_data(path, value, storage_folder)

        for child in children:
            separator = "" if path == "/" else "/"
            new_path = f'{path}{separator}{child}'
            visit_nodes(zk, new_path, storage_folder)


def store_data(path, value, storage_folder):
    fpath = f'{storage_folder}{path}'
    logging.debug(f'Path is {fpath}.')

    try:
        if not os.path.exists(fpath):
            os.makedirs(fpath)
    except OSError:
        logging.error(f'Creation of the directory {fpath} failed.')
        return
    else:
        logging.debug(f'The directory {fpath} is created successfully.')

    if value:
        cpath = f'{fpath}/content'
        logging.debug(f'Path for file is {cpath}.')
        f = open(cpath, 'w')
        f.write(value.decode("cp437"))
        f.close()


def restore(client, nodes_to_restore, storage_folder):
    zk = client.connect_to_zookeeper()
    znodes_folder = f'{storage_folder}/znodes'
    try:
        if os.path.isfile(f'{znodes_folder}.zip'):
            if not nodes_to_restore:
                nodes_to_restore = get_znodes_list_from_archive(znodes_folder)
            for znode in nodes_to_restore:
                extract_znode_from_archive(znode, znodes_folder)
                visit_nodes_while_restoring(zk, znode, znodes_folder)
        else:
            if not nodes_to_restore:
                raise Exception('Restoring operation requires specifying nodes to recover.')
            for node in nodes_to_restore:
                visit_nodes_while_restoring(zk, node, storage_folder)
    finally:
        client.disconnect_from_zookeeper(zk)
        if os.path.isdir(znodes_folder):
            shutil.rmtree(znodes_folder)


def get_znodes_list_from_archive(znodes_folder):
    with zipfile.ZipFile(f'{znodes_folder}.zip') as file:
        nodes_to_restore = [x[:-1] for x in file.namelist() if re.fullmatch(r'^[^/]+/$', x)]
    nodes_to_restore.remove("zookeeper")
    return nodes_to_restore


def extract_znode_from_archive(znode, znodes_folder):
    with zipfile.ZipFile(f'{znodes_folder}.zip') as archive:
        for file in archive.namelist():
            if file.startswith(znode + '/') and file != (znode + '/'):
                archive.extract(file, f'{znodes_folder}/')


def visit_nodes_while_restoring(zk, node, storage_folder):
    num_restored = 0
    path_to_node = f'{storage_folder}/{node}'
    logging.debug(f'Path to node {node} is {path_to_node}.')
    for root, dirs, files in os.walk(path_to_node):
        logging.debug(f'Root is [{root}], dirs are [{dirs}], files are [{files}].')
        znode = "/" + os.path.relpath(root, storage_folder)
        logging.debug(f'znode is {znode}.')
        if zk.exists(znode):
            logging.debug(f'znode {znode} exists already, deleting it.')
            zk.delete(znode, recursive=True)
            logging.debug(f'znode {znode} deleted.')
        data = ""
        if "content" in files:
            f = open(f'{root}/content', 'r')
            data = f.read()
            f.close()
        try:
            zk.create(znode, data.encode("cp437"))
        except:
            logging.error(f"znode {znode} isn't restored.")
        else:
            logging.debug(f'znode {znode} is restored with value {data}.')
            num_restored = num_restored + 1

    logging.debug(f'{num_restored} znode(s) is(are) restored.')
