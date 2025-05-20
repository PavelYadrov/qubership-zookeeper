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

import unittest
from unittest import mock

from backup_metric import _collect_status_metrics, _collect_storage_metrics, _collect_last_backup_metrics, \
    _collect_successful_backups_metrics
from requests import Response

ZOOKEEPER_BACKUP_DAEMON_URL = "http://zookeeper-backup-daemon:8080"


class TestBackupMetric(unittest.TestCase):

    def test_status_metrics_when_status_is_up(self):
        health = {'status': 'UP', 'backup_queue_size': 0}
        expected_message = 'zookeeper_backup_metric status=0'
        actual_message = _collect_status_metrics(health)
        self.assertEqual(expected_message, actual_message)

    def test_status_metrics_when_status_is_warning(self):
        health = {'status': 'Warning', 'storage': {}, 'backup_queue_size': 0}
        expected_message = 'zookeeper_backup_metric status=3'
        actual_message = _collect_status_metrics(health)
        self.assertEqual(expected_message, actual_message)

    def test_storage_metrics(self):
        storage = {'dump_count': 13, 'size': 3718119424, 'free_space': 103644528640, 'total_space': 107362648064,
                   'free_inodes': 51916555, 'total_inodes': 52428272, 'used_inodes': 511717, 'lastSuccessful': {},
                   'last': {}}
        expected_message = 'zookeeper_backup_metric storage_size=107362648064,storage_free_space=103644528640,' \
                           'storage_type=1,backups_count=13'
        actual_message = _collect_storage_metrics(storage)
        self.assertEqual(expected_message, actual_message)

    @mock.patch('os.getenv', mock.Mock(side_effect=lambda param, default: 'true'))
    def test_s3_storage_metrics(self):
        storage = {'dump_count': 21, 'lastSuccessful': {}, 'last': {}}
        expected_message = 'zookeeper_backup_metric storage_size=0,storage_free_space=0,' \
                           'storage_type=2,backups_count=21'
        actual_message = _collect_storage_metrics(storage)
        self.assertEqual(expected_message, actual_message)

    @mock.patch('requests.get', mock.Mock(side_effect=lambda url, auth, verify: get_response_for_completed_backup()))
    def test_last_backup_metrics_if_it_exists(self):
        storage = {'dump_count': 13, 'size': 3718119424, 'free_space': 103644528640, 'total_space': 107362648064,
                   'free_inodes': 51916555, 'total_inodes': 52428272, 'used_inodes': 511717, 'lastSuccessful': {},
                   'last': {'id': '20200410T160500', 'failed': False, 'locked': False, 'sharded': False,
                            'ts': 1586534700000, 'metrics': {'exit_code': 0, 'spent_time': 5139, 'size': 6719}}}
        expected_message = 'zookeeper_backup_metric last_backup_time=1586534700000,last_backup_status=1,' \
                           'last_backup_size=6719,last_backup_spent_time=5139'
        actual_message = _collect_last_backup_metrics(ZOOKEEPER_BACKUP_DAEMON_URL, storage)
        self.assertEqual(expected_message, actual_message)

    def test_last_backup_metrics_if_it_does_not_exist(self):
        storage = {'dump_count': 0, 'size': 3718119424, 'free_space': 103644528640, 'total_space': 107362648064,
                   'free_inodes': 51916555, 'total_inodes': 52428272, 'used_inodes': 511717, 'lastSuccessful': {},
                   'last': {}}
        expected_message = 'zookeeper_backup_metric last_backup_time=-1,last_backup_status=-1'
        actual_message = _collect_last_backup_metrics(ZOOKEEPER_BACKUP_DAEMON_URL, storage)
        self.assertEqual(expected_message, actual_message)

    @mock.patch('backup_metric._get_request_with_path', mock.Mock(
        side_effect=lambda url, path: get_result_with_successful_backups(path)))
    def test_successful_backups_metrics_if_all_backups_are_successful(self):
        storage = {'dump_count': 2, 'size': 3718119424, 'free_space': 103644528640, 'total_space': 107362648064,
                   'free_inodes': 51916555, 'total_inodes': 52428272, 'used_inodes': 511717,
                   'lastSuccessful': {'id': '20200413T000500', 'failed': False, 'locked': False, 'sharded': False,
                                      'ts': 1586736300000,
                                      'metrics': {'exit_code': 0, 'spent_time': 5001, 'size': 6658}},
                   'last': {}}
        expected_message = \
            'zookeeper_backup_metric successful_backups_count=2,last_successful_backup_time=1586736300000'
        actual_message = _collect_successful_backups_metrics(ZOOKEEPER_BACKUP_DAEMON_URL, storage)
        self.assertEqual(expected_message, actual_message)

    @mock.patch('backup_metric._get_request_with_path', mock.Mock(
        side_effect=lambda url, path: get_result_with_unsuccessful_backups(path)))
    def test_successful_backups_metrics_if_there_is_no_successful_backup(self):
        storage = {'dump_count': 2, 'size': 3718119424, 'free_space': 103644528640, 'total_space': 107362648064,
                   'free_inodes': 51916555, 'total_inodes': 52428272, 'used_inodes': 511717, 'lastSuccessful': {},
                   'last': {}}
        expected_message = \
            'zookeeper_backup_metric successful_backups_count=0,last_successful_backup_time=-1'
        actual_message = _collect_successful_backups_metrics(ZOOKEEPER_BACKUP_DAEMON_URL, storage)
        self.assertEqual(expected_message, actual_message)

    @mock.patch('backup_metric._get_request_with_path', mock.Mock(
        side_effect=lambda url, path: get_result_with_different_backups(path)))
    def test_successful_backups_metrics_if_there_are_different_backups(self):
        storage = {'dump_count': 2, 'size': 3718119425, 'free_space': 103644528640, 'total_space': 107362648064,
                   'free_inodes': 51916555, 'total_inodes': 52428272, 'used_inodes': 511717,
                   'lastSuccessful': {'id': '20200413T000500', 'failed': False, 'locked': False, 'sharded': False,
                                      'ts': 1586736300000,
                                      'metrics': {'exit_code': 0, 'spent_time': 5001, 'size': 6658}},
                   'last': {}}
        expected_message = \
            'zookeeper_backup_metric successful_backups_count=1,last_successful_backup_time=1586736300000'
        actual_message = _collect_successful_backups_metrics(ZOOKEEPER_BACKUP_DAEMON_URL, storage)
        self.assertEqual(expected_message, actual_message)


def get_response_for_completed_backup():
    response = Response()
    response.status_code = 404
    response.raw = {'message': 'Sorry, no job \'20200410T160500\' recorded in database'}
    return response


def get_result_with_successful_backups(path):
    if path == 'listbackups':
        return ['20200309T230500', '20200413T000500']
    elif path == 'listbackups/20200309T230500':
        return {'is_granular': False, 'db_list': 'full backup', 'id': '20200309T230500', 'failed': False,
                'locked': False, 'sharded': False, 'ts': 1583795100000, 'exit_code': 0, 'spent_time': '2862ms',
                'size': '6564b', 'valid': True, 'evictable': True}
    elif path == 'listbackups/20200413T000500':
        return {'is_granular': False, 'db_list': 'full backup', 'id': '20200413T000500', 'failed': False,
                'locked': False, 'sharded': False, 'ts': 1586736300000, 'exit_code': 0, 'spent_time': '5001ms',
                'size': '6658b', 'valid': True, 'evictable': True}
    else:
        return {}


def get_result_with_unsuccessful_backups(path):
    if path == 'listbackups':
        return ['20200411T160500', '20200413T100500']
    elif path == 'listbackups/20200411T160500':
        return {'is_granular': False, 'db_list': 'full backup', 'id': '20200411T160500', 'failed': True,
                'locked': False, 'sharded': False, 'ts': 1586621100000, 'exit_code': 1, 'spent_time': '880ms',
                'size': '1638b', 'exception': 'Traceback (most recent call last): ...', 'valid': False,
                'evictable': True}
    elif path == 'listbackups/20200413T100500':
        return {'is_granular': False, 'db_list': 'full backup', 'id': '20200413T100500', 'failed': True,
                'locked': False, 'sharded': False, 'ts': 1586772300000, 'exit_code': 1, 'spent_time': '850ms',
                'size': '1757b', 'exception': 'Traceback (most recent call last): ...', 'valid': False,
                'evictable': True}
    else:
        return {}


def get_result_with_different_backups(path):
    if path == 'listbackups':
        return ['20200411T160500', '20200413T000500']
    elif path == 'listbackups/20200411T160500':
        return {'is_granular': False, 'db_list': 'full backup', 'id': '20200411T160500', 'failed': True,
                'locked': False, 'sharded': False, 'ts': 1586621100000, 'exit_code': 1, 'spent_time': '880ms',
                'size': '1638b', 'exception': 'Traceback (most recent call last): ...', 'valid': False,
                'evictable': True}
    elif path == 'listbackups/20200413T000500':
        return {'is_granular': False, 'db_list': 'full backup', 'id': '20200413T000500', 'failed': False,
                'locked': False, 'sharded': False, 'ts': 1586736300000, 'exit_code': 0, 'spent_time': '5001ms',
                'size': '6658b', 'valid': True, 'evictable': True}
    else:
        return {}


if __name__ == "__main__":
    unittest.main()
