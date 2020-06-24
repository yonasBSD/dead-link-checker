'''Unit tests for app'''

import pytest
import types
from argparse import Namespace
from pathlib import Path
from unittest import mock

from apscheduler.triggers.cron import CronTrigger

from delic import app, models


def get_check_site_result() -> models.SiteResultList:
    '''Returns a possible result for check_site'''
    return models.SiteResultList.parse_obj([
        models.SiteResult(
            site='http://test-site-1',
            summary=models.SiteResultSummary(
                urls_checked=7,
                urls_broken=2,
            ),
            details=models.SiteResultDetails(
                broken=[
                    models.Link(
                        page='http://test-site-1',
                        url='http://test-site-1/404',
                        status='404 - Not Found',
                    ),
                    models.Link(
                        page='http://test-site-1',
                        url='http://test-site-a',
                        status='418 - Im a teapot',
                    ),
                ],
            ),
        ),
        models.SiteResult(
            site='http://test-site-2',
            summary=models.SiteResultSummary(
                urls_checked=5,
                urls_broken=0,
            ),
            details=models.SiteResultDetails(
                broken=[],
            ),
        ),
    ])


def get_config_dict():
    '''Returns a config dictionary'''
    return {
        'sites': [
            'http://test-site-1',
            'http://test-site-2',
        ],
        'notify': {
            'provider': 'test-notify-provider',
            'data': {
                'test-notify-data-1': 'success-1',
                'test-notify-data-2': 'success-2',
            }
        }
    }


@mock.patch('delic.app.send_notification')
@mock.patch('delic.app.check_site')
@mock.patch('delic.app.load_yaml_file')
@mock.patch('delic.app.parse_args')
def test_success_single_run(mock_parse_args: mock.MagicMock,
                            mock_load_config: mock.MagicMock,
                            mock_check_site: mock.MagicMock,
                            mock_notify: mock.MagicMock):
    '''Tests happy flow for a single run (no schedule)'''
    # Setup mocks
    mock_parse_args.return_value = Namespace(
        config='test-config.yml',
        verbose=True,
    )
    mock_load_config.return_value = get_config_dict()
    mock_check_site.side_effect = get_check_site_result().__root__

    # Call function
    app.run()

    # Assert results
    mock_parse_args.assert_called_with(mock.ANY)
    mock_load_config.assert_called_with(Path('test-config.yml'))
    mock_check_site.assert_has_calls([
        mock.call('http://test-site-1', 8),
        mock.call('http://test-site-2', 8),
    ])
    mock_check_site.call_count == 2
    expected_notify = get_check_site_result()
    del expected_notify.__root__[1]
    mock_notify.assert_called_with(
        expected_notify.json(**app.JSON_ARGS),
        'test-notify-provider',
        get_config_dict()['notify']['data']
    )


@mock.patch('delic.app.BlockingScheduler')
@mock.patch('delic.app.load_yaml_file')
@mock.patch('delic.app.parse_args')
def test_success_scheduled(mock_parse_args: mock.MagicMock,
                           mock_load_config: mock.MagicMock,
                           mock_scheduler: mock.MagicMock):
    '''Tests happy flow for a scheduled run'''
    # Setup mocks
    mock_parse_args.return_value = Namespace(
        config='test-config.yml',
        verbose=True,
    )
    config_dict = get_config_dict()
    config_dict['cron'] = '* * * * *'
    mock_load_config.return_value = config_dict
    mock_scheduler_instance = mock_scheduler.return_value

    # Call function
    app.run()

    # Assert results
    mock_parse_args.assert_called_with(mock.ANY)
    mock_load_config.assert_called_with(Path('test-config.yml'))
    assert mock_scheduler_instance.add_job.call_count == 1
    scheduled_def, schedule = mock_scheduler_instance.add_job.call_args[0]
    assert isinstance(scheduled_def, types.FunctionType)
    assert repr(schedule) == repr(CronTrigger.from_crontab('* * * * *'))
    mock_scheduler_instance.start.assert_called_with()


@mock.patch('delic.app.send_notification')
@mock.patch('delic.app.check_site')
@mock.patch('delic.app.load_yaml_file')
@mock.patch('delic.app.parse_args')
def test_success_no_notify_when_no_broken_links(mock_parse_args: mock.MagicMock,
                                                mock_load_config: mock.MagicMock,
                                                mock_check_site: mock.MagicMock,
                                                mock_notify: mock.MagicMock):
    '''Tests happy flow for a single run (no schedule)'''
    # Setup mocks
    mock_parse_args.return_value = Namespace(
        config='test-config.yml',
        verbose=True,
    )
    mock_load_config.return_value = get_config_dict()
    mock_check_site.return_value = get_check_site_result().__root__[1]

    # Call function
    app.run()

    # Assert results
    mock_parse_args.assert_called_with(mock.ANY)
    mock_load_config.assert_called_with(Path('test-config.yml'))
    mock_check_site.assert_has_calls([
        mock.call('http://test-site-1', 8),
        mock.call('http://test-site-2', 8),
    ])
    mock_check_site.call_count == 2
    mock_notify.assert_not_called()
