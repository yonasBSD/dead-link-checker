'''This module handles the actual link checking'''
# pylint: disable=no-member

import logging
import queue
import threading
from html.parser import HTMLParser
from http.client import responses
from typing import List
from urllib.parse import urljoin

import requests

from delic.models import Link, SiteResult, SiteResultDetails, SiteResultSummary

LINK_TAGS = {
    'a': ['href'],
    'img': ['src', 'srcset'],
    'link': ['href'],
    'script': ['src'],
    'source': ['srcset'],
}

IGNORED_SCHEMAS = (
    'data:',
    'ftp:',
    'javascript:',
    'mailto:',
    'tel:',
)

REQUESTS_HEADERS = {
    'User-Agent': 'Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:77.0) Gecko/20100101 Firefox/77.0',
}


class DelicHTMLParser(HTMLParser):
    def __init__(self, link_queue, checked_urls, page):
        super().__init__()
        self.link_queue = link_queue
        self.checked_urls = checked_urls
        self.page = page

    def handle_starttag(self, tag, attrs):
        if tag in LINK_TAGS:
            attr_values = (attr[1]
                           for attr
                           in attrs
                           if attr[0] in LINK_TAGS[tag])
            for attr_value in attr_values:
                # Check if schema is ignored
                if attr_value.startswith(IGNORED_SCHEMAS):
                    continue  # Ignore url

                # Extract url
                target_url = urljoin(self.page, attr_value)
                cleaned_url = target_url.split('#')[0]

                # Add url to queue
                if cleaned_url not in self.checked_urls:
                    new_link = Link(
                        url=cleaned_url,
                        page=self.page,
                    )
                    self.link_queue.put(new_link)


def check_site(base_url, workers_count) -> SiteResult:
    '''Check all links of a single site'''
    # Init
    checked_urls: List[str] = []
    broken_links: List[Link] = []
    link_queue = queue.Queue()

    # Log start
    msg = "Start link checking with %s workers for %s"
    logging.info(msg, workers_count, base_url)

    # Define worker
    def check_link_worker():
        while True:
            link = link_queue.get()
            if link.url not in checked_urls:
                checked_urls.append(link.url)
                check_link(link_queue,
                           checked_urls,
                           broken_links,
                           base_url,
                           link)
            link_queue.task_done()

    # Start worker thread
    for _ in range(workers_count):
        threading.Thread(target=check_link_worker, daemon=True).start()

    # Queue base URL
    base_link = Link(
        url=base_url,
        page='',
    )
    link_queue.put(base_link)

    # Wait until the queue fully processed
    link_queue.join()

    # Return results
    return SiteResult(
        site=base_url,
        summary=SiteResultSummary(
            urls_checked=len(checked_urls),
            urls_broken=len(broken_links),
        ),
        details=SiteResultDetails(
            broken=broken_links,
        ),
    )


def check_link(link_queue, checked_urls, broken_links, base_url, link: Link):
    '''Check a single link'''
    # Create parser
    parser = DelicHTMLParser(link_queue, checked_urls, link.url)

    # Fetch header
    logging.info('Checking URL: %s', link.url)
    req = requests.head(link.url, headers=REQUESTS_HEADERS)

    # Retry with GET if method not allowed
    if req.status_code == requests.codes.method_not_allowed:
        req = requests.get(link.url, headers=REQUESTS_HEADERS)

    # Check status of request
    if req.status_code >= 400:
        link.status = f"{req.status_code} - {responses[req.status_code]}"
        broken_links.append(link)

    # Link is HTML page and is internal
    # Fetch and parse page
    content_type = req.headers.get('content-type', '')
    if content_type.startswith('text/html') and link.url.startswith(base_url):
        req_html = requests.get(link.url, headers=REQUESTS_HEADERS)
        parser.feed(req_html.text)
