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

def check_that_parameters_are_presented(environ, *variable_names) -> bool:
    for variable in variable_names:
        if not environ.get(variable):
            return False
    return True


def get_excluded_tags(environ) -> list:
    excluded_tags = []
    if not check_that_parameters_are_presented(environ,
                                               'ZOOKEEPER_BACKUP_DAEMON_HOST',
                                               'ZOOKEEPER_BACKUP_DAEMON_PORT'):
        excluded_tags.append('zookeeper_backup_daemon')
    pv_type = environ.get('PV_TYPE')

    if not pv_type or pv_type.lower() == 'standalone':
        excluded_tags.append('transactional_backup')

    tags = environ.get('TAGS')
    if not tags or tags.find("full_eviction_test") == -1:
        excluded_tags.append('full_eviction_test')
    if not check_that_parameters_are_presented(environ,
                                               'ZOOKEEPER_BACKUP_DAEMON_USERNAME',
                                               'ZOOKEEPER_BACKUP_DAEMON_PASSWORD'):
        excluded_tags.append('unauthorized_access')
    return excluded_tags
