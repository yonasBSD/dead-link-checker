'''Unit tests for app'''

import pytest
from argparse import Namespace
from unittest import mock


@mock.patch('delic.app.send_notification')
@mock.patch('delic.app.check_site')
@mock.patch('delic.app.load_yaml_file')
@mock.patch('delic.app.parse_args')
def test_success_single_run(mock_parse_args, mock_load_config, mock_check_site, mock_notify):
    '''Tests happy flow for a single run (no schedule)'''
    # Setup mocks
    mock_parse_args.return_value = Namespace(
        config='test-config.yml',
        verbose=True,
    )
    mock_load_config.return_value = {
        'sites': ['http://test']
    }
