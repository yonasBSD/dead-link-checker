'''Unit tests for notify'''
import logging
from unittest import mock

from delic.notify import send_notification


@mock.patch('delic.notify.notifiers')
def test_send_notification(mock_notifiers):
    '''Should send a notification'''
    # Setup mocks
    mock_provider = mock_notifiers.get_notifier.return_value

    # Call function
    send_notification('test-message', 'test-provider', {'test': 'success'})

    # Assert results
    mock_notifiers.get_notifier.assert_called_with('test-provider')
    mock_provider.notify.assert_called_with(
        message='test-message',
        test='success',
    )
