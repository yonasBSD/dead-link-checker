'''Main module combining all logic'''

import logging
import sys
from pathlib import Path

from delic.cli import parse_args
from delic.config import load_yaml_file
from delic.link_checker import check_site


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

    # Check sites
    results = {}
    workers_count = config.get('workers_per_site', 8)
    for site in config['sites']:
        result = check_site(site, workers_count)
        results[site] = result

    # Print results
    print(repr(results))
