'''This module handles the actual link checking'''

import logging
import queue
import threading
from html.parser import HTMLParser
from urllib.parse import urljoin

import requests

LINK_TAGS = {
    'a': ['href'],
    'img': ['src', 'srcset'],
    'link': ['href'],
    'script': ['src'],
    'source': ['srcset'],
}


class DelicHTMLParser(HTMLParser):
    def __init__(self, link_queue, checked_urls, base_url):
        super().__init__()
        self.link_queue = link_queue
        self.checked_urls = checked_urls
        self.base_url = base_url

    def handle_starttag(self, tag, attrs):
        if tag in LINK_TAGS:
            attr_values = (attr[1]
                           for attr
                           in attrs
                           if attr[0] in LINK_TAGS[tag])
            for attr_value in attr_values:
                # Extract url
                target_url = urljoin(self.base_url, attr_value)
                cleaned_url = target_url.split('#')[0]

                # Add url to queue
                if cleaned_url.startswith(self.base_url) and cleaned_url not in self.checked_urls:
                    self.link_queue.put(cleaned_url)


def check_site(base_url):
    '''Check all links of a single site'''
    # Init
    checked_urls = []
    link_queue = queue.Queue()

    # Define worker
    def check_link_worker():
        while True:
            link = link_queue.get()
            if link not in checked_urls:
                checked_urls.append(link)
                check_link(link_queue, checked_urls, base_url, link)
            link_queue.task_done()

    # Start worker thread
    for _ in range(4):
        threading.Thread(target=check_link_worker, daemon=True).start()

    # Queue base URL
    link_queue.put(base_url)

    # Wait until the queue fully processed
    link_queue.join()

    # Return results
    checked_urls.sort()
    for url in checked_urls:
        print(url)


def check_link(link_queue, checked_urls, base_url, url):
    '''Check a single link'''
    # Create parser
    parser = DelicHTMLParser(link_queue, checked_urls, base_url)

    # Fetch and parse content
    logging.info('Checking URL: %s', url)
    req = requests.get(url)
    parser.feed(req.text)
