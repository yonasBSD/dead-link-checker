'''Unit tests for link_checker'''

from unittest import mock

import pytest

from delic.link_checker import DelicHTMLParser
from delic.models import Link


# =================================================
# =                DelicHTMLParser                =
# =================================================
@pytest.mark.parametrize('tag,attr,tag_accepted', [
    ('a', 'href', True),
    ('img', 'src', True),
    ('img', 'srcset', True),
    ('link', 'href', True),
    ('script', 'src', True),
    ('source', 'srcset', True),
    ('p', 'class', False),
    ('a', 'class', False),
    ('a', 'id', False),
])
@pytest.mark.parametrize('schema,schema_accepted', [
    ('http', True),
    ('https', True),
    ('data', False),
    ('ftp', False),
    ('javascript', False),
    ('mailto', False),
    ('tel', False),
])
def test_delic_html_parser_absolute_url(schema, schema_accepted, tag, attr, tag_accepted):
    '''Test parsing of absolute urls'''
    # Setup mocks
    link_queue = mock.MagicMock()
    checked_urls = []

    # Setup test data
    html_snippet = f'<{tag} {attr}="{schema}://example.com/test">Test</{tag}>'

    # Call parser
    parser = DelicHTMLParser(
        link_queue=link_queue,
        checked_urls=checked_urls,
        page='http://example.com/test.html',
    )
    parser.feed(html_snippet)

    # Assert results
    if schema_accepted and tag_accepted:
        expected_link = Link(
            page='http://example.com/test.html',
            url=f'{schema}://example.com/test'
        )
        link_queue.put.assert_called_with(expected_link)
    else:
        link_queue.put.assert_not_called()


@pytest.mark.parametrize('link,expected_url', [
    ('test', 'http://example.com/subfolder/test'),
    ('/test', 'http://example.com/test'),
    ('/test#hashtag', 'http://example.com/test'),
])
def test_delic_html_parser_relative_url(link, expected_url):
    '''Test parsing of relative urls'''
    # Setup mocks
    link_queue = mock.MagicMock()
    checked_urls = []

    # Setup test data
    html_snippet = f'<a href="{link}">Test</a>'

    # Call parser
    parser = DelicHTMLParser(
        link_queue=link_queue,
        checked_urls=checked_urls,
        page='http://example.com/subfolder/index.html',
    )
    parser.feed(html_snippet)

    # Assert results
    expected_link = Link(
        page='http://example.com/subfolder/index.html',
        url=expected_url
    )
    link_queue.put.assert_called_with(expected_link)


def test_delic_html_parser_url_already_checked():
    '''Already checked urls should be ignored'''
    # Setup mocks
    link_queue = mock.MagicMock()
    checked_urls = ['http://example.com/test.html']

    # Setup test data
    html_snippet = f'<a href="test.html">Test</a>'

    # Call parser
    parser = DelicHTMLParser(
        link_queue=link_queue,
        checked_urls=checked_urls,
        page='http://example.com/index.html',
    )
    parser.feed(html_snippet)

    # Assert results
    link_queue.put.assert_not_called()
