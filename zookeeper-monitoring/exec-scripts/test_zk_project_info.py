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

from zk_project_info import _collect_metrics


class TestZkProjectInfo(unittest.TestCase):

    def test_metrics_if_pv_names_are_specified(self):
        pv_names = 'pv-zk-zookeeper-1,pv-zk-zookeeper-2,pv-zk-zookeeper-3'
        expected_message = f'zookeeper_ext,zk_pv_name1=pv-zk-zookeeper-1,zk_pv_name2=pv-zk-zookeeper-2,zk_pv_name3=pv-zk-zookeeper-3 pv_count=3'
        actual_message = _collect_metrics(pv_names)
        self.assertEqual(expected_message, actual_message)

    def test_metrics_if_one_pv_name_is_specified(self):
        pv_names = 'pv-zk-zookeeper-1'
        expected_message = f'zookeeper_ext,zk_pv_name1=pv-zk-zookeeper-1 pv_count=1'
        actual_message = _collect_metrics(pv_names)
        self.assertEqual(expected_message, actual_message)

    def test_metrics_if_pv_names_are_not_specified(self):
        pv_names = ''
        expected_message = f'zookeeper_ext,zk_pv_name1=empty pv_count=1'
        actual_message = _collect_metrics(pv_names)
        self.assertEqual(expected_message, actual_message)


if __name__ == "__main__":
    unittest.main()
