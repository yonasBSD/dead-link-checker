'''Unit tests for config'''

import pytest

from delic import config


def test_load_yaml_file(tmp_path):
    '''Should read a YAML file'''
    # Write test file
    yaml_content = '''test:
                        success: True
                   '''
    yaml_file = tmp_path / "test-config.yml"
    yaml_file.write_text(yaml_content)

    # Try to load file
    result = config.load_yaml_file(yaml_file)

    # Expected result
    expected = {
        'test': {
            'success': True
        }
    }

    # Assert result
    assert result == expected


def test_load_config_file(tmp_path):
    '''Should read a YAML file and update with defaults'''
    # Write test file
    yaml_content = '''test:
                        success: True
                   '''
    yaml_file = tmp_path / "test-config.yml"
    yaml_file.write_text(yaml_content)

    # Try to load file
    result = config.load_config_file(yaml_file)

    # Expected result
    expected = {
        'test': {
            'success': True
        },
        'verbose': False,
        'workers_per_site': 8,
    }

    # Assert result
    assert result == expected
