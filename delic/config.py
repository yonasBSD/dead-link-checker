'''Module to handle the config file'''

import logging
from pathlib import Path
from typing import Dict

import yaml


def load_yaml_file(path: Path) -> Dict:
    '''Load and parse the YAML file at path'''
    with path.open(mode='r') as file_config:
        return yaml.safe_load(file_config)


def load_config_file(path: Path, args) -> Dict:
    '''Load YAML config file and set defaults. Arguments overrides config.'''
    # Load config
    config = load_yaml_file(path)

    # Set defaults
    config['verbose'] = True if args.verbose else config.get('verbose', False)
    config['workers_per_site'] = config.get('workers_per_site', 8)

    # Validate verbose
    if isinstance(config['verbose'], bool):
        set_log_level(config['verbose'])
    else:
        set_log_level(True)
        logging.warning('Verbose is not a bool: "%s". Defaulted to True.',
                        config['verbose'])
        config['verbose'] = True

    # Validate workers_per_site
    if not isinstance(config['workers_per_site'], int) or config['workers_per_site'] < 1:
        logging.warning('Invalid number of workers: "%s". Defaulted to 8 workers.',
                        config['workers_per_site'])
        config['workers_per_site'] = 8

    # Return result
    return config


def set_log_level(verbose):
    '''Set logging to INFO, if verbose is enabled'''
    if verbose:
        logging.basicConfig(level=logging.INFO)
