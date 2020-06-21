'''This module handles notifying the user when broken links are found'''

import logging

import notifiers


def send_notification(message, providerName, providerData):
    '''Sends a notification to the user'''
    # Log call
    logging.info(
        'Preparing to send notification to user using provider "%s"',
        providerName
    )

    # Send notification to user
    provider = notifiers.get_notifier(providerName)
    result = provider.notify(message=message, **providerData)

    # Log result
    if result.status == notifiers.SUCCESS_STATUS:
        logging.info("Notification successfully sent")
    else:
        logging.error(
            "Notifying user failed with error(s): %s",
            repr(result.errors)
        )
