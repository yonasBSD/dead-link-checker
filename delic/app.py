'''Main module combining all logic'''

import json
import logging
import sys
from pathlib import Path

from apscheduler.schedulers.blocking import BlockingScheduler
from apscheduler.triggers.cron import CronTrigger

from delic.cli import parse_args
from delic.config import load_yaml_file
from delic.link_checker import check_site
from delic.notify import send_notification


def run():
    '''Run Dead Link Checker'''
    # Parse arguments
    args = parse_args()

    # Load config file
    try:
        config_path = Path(args.config)
        config = load_yaml_file(config_path)
    except FileNotFoundError:
        logging.error('Config file not found at path "%s"',
                      config_path.absolute())
        sys.exit(1)

    # Set logging to INFO, if verbose is enabled
    if args.verbose or config.get('verbose'):
        logging.basicConfig(level=logging.INFO)

    # Run job once or start scheduler
    crontab = config.get('cron')
    if crontab:
        logging.info(
            "Crontab found, starting scheduler with crontab: %s", crontab)
        scheduler = BlockingScheduler()
        def check_sites_job(): check_sites(config)
        scheduler.add_job(check_sites_job, CronTrigger.from_crontab(crontab))
        logging.info("Scheduler started for specified crontab")
        scheduler.start()
    else:
        logging.info("No cron specified, running checker once")
        check_sites(config)


def check_sites(config):
    # Check sites
    results = []
    workers_count = config.get('workers_per_site', 8)
    for site in config['sites']:
        result = check_site(site, workers_count)
        results.append(result)

    # Print results
    print(pretty_json(results))

    # Send notification if settings are provided
    notify_settings = config.get('notify', {})
    notify_provider = notify_settings.get('provider')
    if notify_provider:
        # Filter results with broken links
        broken_results = {
            x for x in results if x['summary']['urls_broken'] > 0}

        # Notify user of broken results
        if len(broken_results) > 0:
            notify_data = notify_settings.get('data', {})
            pretty_results = pretty_json(broken_results)
            send_notification(pretty_results, notify_provider, notify_data)


def pretty_json(object_):
    '''Returns pretty json'''
    return json.dumps(
        object_,
        indent=4,
    )
