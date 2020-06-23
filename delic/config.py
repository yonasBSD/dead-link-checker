'''Module to handle the config file'''

from pathlib import Path
from typing import Dict

import yaml


def load_yaml_file(path: Path) -> Dict:
    '''Load and parse the YAML file at path'''
    with path.open(mode='r') as file_config:
        return yaml.safe_load(file_config)
