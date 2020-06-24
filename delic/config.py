'''Module to handle the config file'''

from pathlib import Path
from typing import Dict

import yaml


def load_yaml_file(path: Path) -> Dict:
    '''Load and parse the YAML file at path'''
    with path.open(mode='r') as file_config:
        return yaml.safe_load(file_config)


def load_config_file(path: Path) -> Dict:
    '''Load YAML config file and set defaults'''
    config = load_yaml_file(path)
    config['verbose'] = config.get('verbose', False)
    config['workers_per_site'] = config.get('workers_per_site', 8)
    return config
