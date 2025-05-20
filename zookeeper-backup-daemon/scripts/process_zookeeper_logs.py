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
import os
import sys
import time
from os.path import join, isfile
from shutil import copy2, rmtree

from parse_transaction_logs import LogFileHeader, Txn, END_OF_STREAM, EOS, UnknownType


def get_snapshot_and_transaction_logs(directory):
    logging.debug('Start to get snapshot and transaction logs.')
    snapshots = []
    transaction_logs = []
    # Find snapshots and transaction logs
    for file_name in os.listdir(directory):
        file_path = join(directory, file_name)
        if isfile(file_path):
            if 'snapshot.' in file_name:
                snapshots.append(file_path)
            if 'log.' in file_name:
                transaction_logs.append(file_path)

    if not snapshots:
        logging.error('There are no snapshots in ZooKeeper to perform backup.')
        sys.exit(1)

    # Find last created snapshot by modification time
    last_snapshot = snapshots[0]
    last_snapshot_stat_info = os.stat(last_snapshot)
    for snapshot in snapshots:
        snapshot_stat_info = os.stat(snapshot)
        if snapshot_stat_info.st_mtime > last_snapshot_stat_info.st_mtime:
            last_snapshot = snapshot
            last_snapshot_stat_info = snapshot_stat_info

    logging.info(f'Last created snapshot is {last_snapshot}.')

    # Find actual transaction logs that were written after last snapshot creation
    actual_transaction_logs = []
    for transaction_log in transaction_logs:
        transaction_log_stat_info = os.stat(transaction_log)
        if transaction_log_stat_info.st_mtime > last_snapshot_stat_info.st_mtime:
            actual_transaction_logs.append(transaction_log)

    logging.debug(f'Actual transaction logs are {actual_transaction_logs}.')
    return last_snapshot, actual_transaction_logs


def filter_and_store_transaction_logs(transaction_logs_files, storage_folder):
    logging.debug('Try to filter logs.')
    for transaction_logs_file in transaction_logs_files:
        filter_and_store_transaction_log(transaction_logs_file, storage_folder)
    logging.debug('Logs are filtered.')


def filter_and_store_transaction_log(transaction_logs_file, storage_folder):
    file_name = os.path.basename(transaction_logs_file)
    input_file = open(transaction_logs_file, 'rb')
    output_file = open(f'{storage_folder}/{file_name}', 'wb')
    log_header = LogFileHeader(input_file)
    if not log_header.is_valid():
        logging.error(f"Not a valid ZooKeeper transaction log '{transaction_logs_file}'.")
    output_file.write(log_header.data_bytes)

    start = None
    try:
        while True:
            transaction = Txn(input_file)
            if not start:
                start = transaction.header.time
                logging.debug('Log starts at %s and %ims' % (time.ctime(start / 1000), start % 1000))
            diff = transaction.header.time - start
            logging.debug('%09i,%03i %s' % (diff / 1000, diff % 1000, str(transaction)[33:]))
            output_file.write(transaction.transaction_bytes)
    except EOS:
        output_file.write(END_OF_STREAM)
    except UnknownType:
        output_file.write(END_OF_STREAM)
        logging.exception('Log file %s processing completed with error:', file_name)

    output_file.close()
    input_file.close()


def copy_zookeeper_logs(directory_from, directory_to):
    for file_name in os.listdir(directory_from):
        file_path = join(directory_from, file_name)
        if isfile(file_path):
            copy2(file_path, directory_to)
    logging.info(f"Files are copied from '{directory_from}' to '{directory_to}'.")


def copy_snapshot(snapshot, storage_folder):
    copy2(snapshot, storage_folder)


def create_directory(directory):
    if not os.path.exists(directory):
        os.makedirs(directory)
        logging.info(f"Directory '{directory}' is created.")
    else:
        logging.info("Directory '%s' exists.", directory)


def remove_directory_with_content(directory):
    logging.debug(f'Try to remove directory: {directory}.')
    if os.path.exists(directory):
        rmtree(directory)
        logging.info(f"Directory '{directory}' is removed.")
    else:
        logging.info(f"Directory '{directory}' doesn't exist.")


def is_file_system_shared():
    pv_type = os.getenv("PV_TYPE")
    return True if pv_type and pv_type != 'standalone' else False
