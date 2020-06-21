'''This module handles notifying the user when broken links are found'''

import notifiers


def send_notification(message, providerName, providerData):
    '''Sends a notification to the user'''
    provider = notifiers.get_notifier(providerName)
    provider.notify(message=message, **providerData)
