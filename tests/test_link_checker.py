'''Unit tests for link_checker'''

import logging
from unittest import mock

import pytest
import responses

from delic.link_checker import DelicHTMLParser, check_site, check_link_worker, check_link
from delic.models import Link, SiteResult, SiteResultSummary, SiteResultDetails


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


# =================================================
# =                   check_site                  =
# =================================================
@mock.patch('delic.link_checker.threading.Thread')
@mock.patch('delic.link_checker.queue.Queue')
def test_check_site(mock_queue, mock_thread):
    '''Should start the worker threads and feed initial link to queue'''
    # Setup mocks
    mock_queue_instance: mock.MagicMock = mock_queue.return_value
    mock_thread_instance: mock.MagicMock = mock_thread.return_value

    # Call function
    result = check_site('http://example.com', 4)

    # Expected result
    expected = SiteResult(
        site='http://example.com',
        summary=SiteResultSummary(
            urls_checked=0,
            urls_broken=0,
        ),
        details=SiteResultDetails(
            broken=[],
        ),
    )

    # Assert results
    assert mock_thread.call_count == 4
    assert mock_thread.call_args_list == 4 * [mock.call(
        target=check_link_worker,
        args=(mock_queue_instance, [], [], 'http://example.com'),
        daemon=True,
    )]
    assert mock_thread_instance.start.call_count == 4
    mock_queue_instance.put.assert_called_with(Link(
        page='',
        url='http://example.com',
    ))
    mock_queue_instance.join.assert_called_with()


# =================================================
# =               check_link_worker               =
# =================================================

# Fix_ME

# =================================================
# =                   check_link                  =
# =================================================
STATUS_SUCCESS_MAP = [
    (200, True),
    (400, False),
    (401, False),
    (500, False),
]


@pytest.mark.parametrize('status,success', STATUS_SUCCESS_MAP)
@mock.patch('delic.link_checker.DelicHTMLParser')
@responses.activate
def test_check_link_internal_html_page(mock_parser, status, success):
    '''Internal HTML pages should be fetched and fed to parser'''
    # Setup Requests mock
    url = 'http://example.com/test.html'
    responses.add(responses.HEAD, url, status=status, content_type='text/html')
    responses.add(responses.GET, url, body='test-html-page', status=200)

    # Setup mocks
    broken_links = []
    mock_parser_instance = mock_parser.return_value

    # Call function
    link = Link(
        page='http://example.com/index.html',
        url=url,
    )
    check_link(
        link_queue=None,
        checked_urls=[],
        broken_links=broken_links,
        base_url='http://example.com',
        link=link,
    )

    # Assert results
    if success:
        assert len(broken_links) == 0
        mock_parser_instance.feed.assert_called_with('test-html-page')
    else:
        assert len(broken_links) == 1
        assert str(status) in broken_links[0].status
        mock_parser_instance.feed.assert_not_called()


@pytest.mark.parametrize('status,success', STATUS_SUCCESS_MAP)
@mock.patch('delic.link_checker.DelicHTMLParser')
@responses.activate
def test_check_link_internal_other(mock_parser, status, success):
    '''Other links should only be checked with HEAD'''
    # Setup Requests mock
    url = 'http://example.com/test.html'
    responses.add(responses.HEAD, url, status=status,
                  content_type='image/jpeg')

    # Setup mocks
    broken_links = []
    mock_parser_instance = mock_parser.return_value

    # Call function
    check_link(
        link_queue=None,
        checked_urls=[],
        broken_links=broken_links,
        base_url='http://example.com',
        link=Link(
            page='http://example.com/index.html',
            url=url,
        ),
    )

    # Assert results
    mock_parser_instance.feed.assert_not_called()
    if success:
        assert len(broken_links) == 0
    else:
        assert len(broken_links) == 1
        assert str(status) in broken_links[0].status


@pytest.mark.parametrize('status,success', STATUS_SUCCESS_MAP)
@mock.patch('delic.link_checker.DelicHTMLParser')
@responses.activate
def test_check_link_external_html_page(mock_parser, status, success):
    '''External HTML pages should only be checked with HEAD'''
    # Setup Requests mock
    url = 'http://external.com'
    responses.add(responses.HEAD, url, status=status, content_type='text/html')

    # Setup mocks
    broken_links = []
    mock_parser_instance = mock_parser.return_value

    # Call function
    check_link(
        link_queue=None,
        checked_urls=[],
        broken_links=broken_links,
        base_url='http://example.com',
        link=Link(
            page='http://example.com/index.html',
            url=url,
        ),
    )

    # Assert results
    mock_parser_instance.feed.assert_not_called()
    if success:
        assert len(broken_links) == 0
    else:
        assert len(broken_links) == 1
        assert str(status) in broken_links[0].status


@pytest.mark.parametrize('status,success', STATUS_SUCCESS_MAP)
@mock.patch('delic.link_checker.DelicHTMLParser')
@responses.activate
def test_check_link_retry_with_get(mock_parser, status, success):
    '''Should retry with GET when HEAD returns 405 (Method not allowed)'''
    # Setup Requests mock
    url = 'http://example.com/test.html'
    responses.add(responses.HEAD, url, status=405)
    responses.add(responses.GET, url, status=status)

    # Setup mocks
    broken_links = []
    mock_parser_instance = mock_parser.return_value

    # Call function
    check_link(
        link_queue=None,
        checked_urls=[],
        broken_links=broken_links,
        base_url='http://example.com',
        link=Link(
            page='http://example.com/index.html',
            url=url,
        ),
    )

    # Assert results
    mock_parser_instance.feed.assert_not_called()
    if success:
        assert len(broken_links) == 0
    else:
        assert len(broken_links) == 1
        assert str(status) in broken_links[0].status
