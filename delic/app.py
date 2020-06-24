'''Main module combining all logic'''

import json
import logging
import sys
from pathlib import Path
from typing import List

from apscheduler.schedulers.blocking import BlockingScheduler
from apscheduler.triggers.cron import CronTrigger

from delic.cli import parse_args
from delic.config import load_config_file
from delic.link_checker import check_site
from delic.models import SiteResultList
from delic.notify import send_notification

JSON_ARGS = {
    'indent': 4
}


def run():
    '''Run Dead Link Checker'''
    # Parse arguments
    args = parse_args(sys.argv[1:])

    # Load config file
    try:
        config_path = Path(args.config)
        config = load_config_file(config_path)
    except FileNotFoundError:
        logging.error('Config file not found at path "%s"',
                      config_path.absolute())
        sys.exit(1)

    # Set logging to INFO, if verbose is enabled
    if args.verbose or config['verbose']:
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
    results = SiteResultList()
    workers_count = config['workers_per_site']
    for site in config['sites']:
        result = check_site(site, workers_count)
        results.__root__.append(result)

    # Print results
    print(results.json(**JSON_ARGS))

    # Send notification if settings are provided
    notify_settings = config.get('notify', {})
    notify_provider = notify_settings.get('provider')
    if notify_provider:
        # Filter results with broken links
        broken_results_list = [
            x for x in results.__root__ if x.summary.urls_broken > 0]

        # Notify user of broken results
        if len(broken_results_list) > 0:
            notify_data = notify_settings.get('data', {})
            broken_results = SiteResultList.parse_obj(broken_results_list)
            json_results = broken_results.json(**JSON_ARGS)
            send_notification(json_results, notify_provider, notify_data)
