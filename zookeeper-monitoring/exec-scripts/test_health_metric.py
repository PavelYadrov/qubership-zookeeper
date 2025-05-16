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
from unittest.mock import patch, call, Mock

from health_metric import _collect_metrics, get_leader_node, get_server_state

ZOOKEEPER_HOST = "'zookeeper-1:2181','zookeeper-2:2181','zookeeper-3:2181'"


class TestHealthMetric(unittest.TestCase):

    def test_metrics_if_zookeeper_is_not_available(self):
        expected_message = 'zookeeper,leader=NA status_code=10,alive_nodes=0,total_nodes=3'
        actual_message = _collect_metrics(ZOOKEEPER_HOST)
        self.assertEqual(expected_message, actual_message)

    @mock.patch('health_metric._get_number_of_alive_nodes', mock.Mock(
        side_effect=lambda servers, formatted_command: 3))
    def test_metrics_if_all_zookeeper_servers_are_available(self):
        expected_message = 'zookeeper,leader=NA status_code=0,alive_nodes=3,total_nodes=3'
        actual_message = _collect_metrics(ZOOKEEPER_HOST)
        self.assertEqual(expected_message, actual_message)

    @mock.patch('health_metric._get_number_of_alive_nodes', mock.Mock(
        side_effect=lambda servers, formatted_command: 1))
    def test_metrics_if_not_all_zookeeper_servers_are_available(self):
        expected_message = 'zookeeper,leader=NA status_code=5,alive_nodes=1,total_nodes=3'
        actual_message = _collect_metrics(ZOOKEEPER_HOST)
        self.assertEqual(expected_message, actual_message)

    @patch('health_metric._get_number_of_alive_nodes')
    @patch('health_metric.get_leader_node')
    @patch('health_metric._str2bool')
    @patch('os.getenv')
    def test_collect_metrics_checks_for_leader(self, mock_getenv, mock_str2bool, mock_get_leader_info, mock_get_number_of_alive_nodes):
        # Setup
        mock_getenv.return_value = "true"  # Changed to "true"
        mock_str2bool.return_value = True  # Changed to True
        mock_get_number_of_alive_nodes.return_value = 3
        mock_get_leader_info.return_value = "zookeeper-2"  # Kept as zookeeper-2
        
        # Execute
        servers = "'zookeeper-1:2181','zookeeper-2:2181','zookeeper-3:2181'"
        actual_message = _collect_metrics(servers)
        
        # Assert
        expected_message = "zookeeper,leader=zookeeper-2 status_code=0,alive_nodes=3,total_nodes=3"
        self.assertEqual(actual_message, expected_message)
        
        mock_get_leader_info.assert_called_once_with(servers.split(','))
        mock_get_number_of_alive_nodes.assert_called_once()
        
        # Debug output
        print(f"Debug - actual_message: {actual_message}")
        print(f"Debug - mock_get_leader_info.call_count: {mock_get_leader_info.call_count}")
        print(f"Debug - mock_get_leader_info.call_args_list: {mock_get_leader_info.call_args_list}")
        print(f"Debug - Actual leader: {actual_message.split('leader=')[1].split(' ')[0]}")

    @patch('health_metric.get_server_state')
    def test_get_leader_info_real(self, mock_get_leader_info):
        mock_get_leader_info.side_effect = ["follower", "leader", "follower"]
        
        leader = get_leader_node(ZOOKEEPER_HOST.split(','))
        print(f"Debug - Actual leader: {leader}")
        self.assertNotEqual(leader, "NA")
        self.assertEqual(leader, "zookeeper-2")

if __name__ == "__main__":
    unittest.main()