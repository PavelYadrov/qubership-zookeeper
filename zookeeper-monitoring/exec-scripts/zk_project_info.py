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

logger = logging.getLogger(__name__)


def __configure_logging(log):
    log.setLevel(logging.DEBUG)
    formatter = logging.Formatter(fmt='[%(asctime)s,%(msecs)03d][%(levelname)s] %(message)s',
                                  datefmt='%Y-%m-%dT%H:%M:%S')
    log_handler = RotatingFileHandler(filename='/opt/zookeeper-monitoring/exec-scripts/zk_project_info_metric.log',
                                      maxBytes=50 * 1024,
                                      backupCount=5)
    log_handler.setFormatter(formatter)
    log_handler.setLevel(logging.DEBUG if os.getenv('ZOOKEEPER_MONITORING_SCRIPT_DEBUG') else logging.INFO)
    log.addHandler(log_handler)
    err_handler = RotatingFileHandler(filename='/opt/zookeeper-monitoring/exec-scripts/zk_project_info_metric.err',
                                      maxBytes=50 * 1024,
                                      backupCount=5)
    err_handler.setFormatter(formatter)
    err_handler.setLevel(logging.ERROR)
    log.addHandler(err_handler)


def _prepare_zookeeper_pv_list(pv_names: list):
    pv_names_tags = []
    for index, pv_name in enumerate(pv_names, 1):
        pv_names_tags.append(f'zk_pv_name{index}={pv_name}')
    logger.debug(f'ZooKeeper persistent volumes list is {pv_names_tags}')
    return ','.join(pv_names_tags)


def _collect_metrics(pv_names: str):
    pv_names_list = pv_names.split(',') if pv_names else ['empty']
    pv_count = len(pv_names_list)
    pv_names_tags = _prepare_zookeeper_pv_list(pv_names_list)
    if pv_names:
        logger.info(f'{pv_count} persistent volumes with names [{pv_names_tags}] have been specified.')
    else:
        logger.info('Persistent volumes are not specified.')

    return f'zookeeper_ext,{pv_names_tags} pv_count={pv_count}'


def run():
    logger.info('Start collecting ZooKeeper persistent volumes metrics...')
    pv_names = os.getenv('ZOOKEEPER_PV_NAMES')
    message = _collect_metrics(pv_names)
    print(message)


if __name__ == "__main__":
    __configure_logging(logger)
    run()
