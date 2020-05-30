'''Main module combining all logic'''

import logging
import sys
from pathlib import Path

from delic.cli import parse_args
from delic.config import load_yaml_file


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

    # Print config file
    print(repr(config))
