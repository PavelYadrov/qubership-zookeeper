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
from logging.handlers import RotatingFileHandler

import requests

ZOOKEEPER_BACKUP_DAEMON_USERNAME = os.getenv('ZOOKEEPER_BACKUP_DAEMON_USERNAME')
ZOOKEEPER_BACKUP_DAEMON_PASSWORD = os.getenv('ZOOKEEPER_BACKUP_DAEMON_PASSWORD')
logger = logging.getLogger(__name__)


def __configure_logging(log):
    log.setLevel(logging.DEBUG)
    formatter = logging.Formatter(fmt='[%(asctime)s,%(msecs)03d][%(levelname)s] %(message)s',
                                  datefmt='%Y-%m-%dT%H:%M:%S')
    log_handler = RotatingFileHandler(filename='/opt/zookeeper-monitoring/exec-scripts/backup_metric.log',
                                      maxBytes=50 * 1024,
                                      backupCount=5)
    log_handler.setFormatter(formatter)
    log_handler.setLevel(logging.DEBUG if os.getenv('ZOOKEEPER_MONITORING_SCRIPT_DEBUG') else logging.INFO)
    log.addHandler(log_handler)
    err_handler = RotatingFileHandler(filename='/opt/zookeeper-monitoring/exec-scripts/backup_metric.err',
                                      maxBytes=50 * 1024,
                                      backupCount=5)
    err_handler.setFormatter(formatter)
    err_handler.setLevel(logging.ERROR)
    log.addHandler(err_handler)


def _str2bool(v: str) -> bool:
    return v and v.lower() in ("yes", "true", "t", "1")

BACKUP_DAEMON_TLS_ENABLED = _str2bool(os.getenv('BACKUP_DAEMON_TLS_ENABLED'))

TLS_ROOT_CA_CERTIFICATE = '/tls/backup/ca.crt' if BACKUP_DAEMON_TLS_ENABLED is not None \
                                                  and BACKUP_DAEMON_TLS_ENABLED else True


def _get_request_with_path(url: str, path: str):
    try:
        response = requests.get(f'{url}/{path}',
                                auth=(ZOOKEEPER_BACKUP_DAEMON_USERNAME, ZOOKEEPER_BACKUP_DAEMON_PASSWORD),
                                verify=TLS_ROOT_CA_CERTIFICATE)
        return response.json()
    except Exception:
        logger.exception('ZooKeeper Backup Daemon is not available. Backup metrics are not written.')
        return ''


def collect_metrics(zookeeper_backup_daemon_url: str):
    logger.info('Start to collect metrics')

    # If request fails, it means that ZooKeeper Backup Daemon is not available. In this case, we should interrupt all
    # further calculations and write activity status of ZooKeeper Backup Daemon to Grafana.
    health = _get_request_with_path(zookeeper_backup_daemon_url, 'health')
    if not health:
        print(f'zookeeper_backup_metric status=-2')
        return

    print(_collect_status_metrics(health))

    storage = health['storage']
    print(_collect_storage_metrics(storage))
    print(_collect_last_backup_metrics(zookeeper_backup_daemon_url, storage))
    print(_collect_successful_backups_metrics(zookeeper_backup_daemon_url, storage))


def _collect_status_metrics(health):
    logger.info('Start to collect status metrics.')

    status = health['status']
    status_code = _get_status_code(status)

    return f'zookeeper_backup_metric status={status_code}'


def _collect_storage_metrics(storage):
    logger.info('Start to collect storage metrics.')

    s3_enabled = _str2bool(os.getenv('S3_ENABLED', 'false'))
    storage_type = 's3' if s3_enabled else 'fs'
    storage_code = _get_storage_code(storage_type)

    storage_size = storage.get('total_space', 0)
    storage_free_space = storage.get('free_space', 0)
    available_backups_count = storage['dump_count']

    return f'zookeeper_backup_metric storage_size={storage_size},storage_free_space={storage_free_space},' \
           f'storage_type={storage_code},backups_count={available_backups_count}'


def _collect_last_backup_metrics(zookeeper_backup_daemon_url: str, storage):
    logger.info('Start to collect last backup metrics.')

    if storage['dump_count']:
        last_backup = storage['last']

        last_backup_id = last_backup['id']
        last_backup_time = last_backup['ts']
        last_backup_metrics = last_backup['metrics']
        last_backup_size = last_backup_metrics['size']
        last_backup_spent_time = last_backup_metrics['spent_time']

        last_backup_status = 'Failed' if last_backup['failed'] else 'Successful'
        response = requests.get(f'{zookeeper_backup_daemon_url}/jobstatus/{last_backup_id}',
                                auth=(ZOOKEEPER_BACKUP_DAEMON_USERNAME, ZOOKEEPER_BACKUP_DAEMON_PASSWORD),
                                verify=TLS_ROOT_CA_CERTIFICATE)
        if response.status_code != 404:
            last_backup_status = response.json().get('status') or 'Failed'
        backup_status_code = _get_backup_status_code(last_backup_status)

        return f'zookeeper_backup_metric last_backup_time={last_backup_time},last_backup_status={backup_status_code},' \
               f'last_backup_size={last_backup_size},last_backup_spent_time={last_backup_spent_time}'
    else:
        return f'zookeeper_backup_metric last_backup_time=-1,last_backup_status=-1'


def _collect_successful_backups_metrics(zookeeper_backup_daemon_url: str, storage):
    logger.info('Start to collect successful backups metrics.')

    successful_backups_count = _get_count_of_successful_backups(zookeeper_backup_daemon_url)
    last_successful_backup_time = -1
    if successful_backups_count:
        last_successful_backup_time = storage['lastSuccessful']['ts']

    return f'zookeeper_backup_metric successful_backups_count={successful_backups_count},' \
           f'last_successful_backup_time={last_successful_backup_time}'


def _get_count_of_successful_backups(zookeeper_backup_daemon_url: str):
    backups_list = _get_request_with_path(zookeeper_backup_daemon_url, 'listbackups')
    logger.debug(f'IDs are {backups_list}')
    successful_backups_count = 0
    for backup in backups_list:
        backup_info_json = _get_request_with_path(zookeeper_backup_daemon_url, f'listbackups/{backup}')
        if not backup_info_json['failed']:
            logger.debug(f'Backup {backup} is successful: {backup_info_json}')
            successful_backups_count += 1
    logger.debug(f'The number of successful backups is {successful_backups_count}')
    return successful_backups_count


def _get_status_code(status: str):
    if status == 'UP':
        return 0
    elif status == 'Warning':
        return 3
    else:
        return -1


def _get_backup_status_code(status: str):
    if status == 'Successful':
        return 1
    elif status == 'Processing':
        return 2
    elif status == 'Queued':
        return 3
    elif status == 'Failed':
        return 4
    else:
        return -1


def _get_storage_code(storage_type: str):
    if storage_type == 'fs':
        return 1
    elif storage_type == 's3':
        return 2
    else:
        return -1


def run():
    logger.info('Start collecting ZooKeeper backup metrics...')
    try:
        zookeeper_backup_daemon_host = os.getenv('ZOOKEEPER_BACKUP_DAEMON_HOST')
        protocol, port = ('https', 8443) if BACKUP_DAEMON_TLS_ENABLED else ('http', 8080)
        if zookeeper_backup_daemon_host:
            zookeeper_backup_daemon_url = f'{protocol}://{zookeeper_backup_daemon_host}:{port}'
            collect_metrics(zookeeper_backup_daemon_url)
        else:
            logger.info('To calculate backup metrics specify ZOOKEEPER_BACKUP_DAEMON_HOST variable.')
            print(f'zookeeper_backup_metric status=-3')
    except Exception:
        logger.exception('Exception occurred during script execution:')
        raise


if __name__ == "__main__":
    __configure_logging(logger)
    run()
