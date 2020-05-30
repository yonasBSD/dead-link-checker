'''Module to handle the config file'''

from pathlib import Path

import yaml


def load_yaml_file(path):
    '''Load and parse the YAML file at path'''
    with path.open(mode='r') as file_config:
        return yaml.safe_load(file_config)
