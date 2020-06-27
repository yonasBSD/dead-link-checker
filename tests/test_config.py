'''Unit tests for config'''

import pytest
import logging

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


def test_load_valid_config_file(tmp_path):
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


@pytest.mark.parametrize('verbose_config, expected_verbose', [
    (False, False),
    (True, True),
    ('test', True),
    (0, True),
    (1, True),
])
@pytest.mark.parametrize('worker_config, expected_workers', [
    (-10, 8),
    (-1, 8),
    (0, 8),
    (1, 1),
    (4, 4),
    ('A', 8),
])
def test_load_valid_config_file(worker_config, expected_workers, verbose_config, expected_verbose, tmp_path):
    '''Should read a YAML file and update with defaults'''
    # Write test file
    yaml_content = f'''
                    verbose: {verbose_config}
                    workers_per_site: {worker_config}
                    '''
    yaml_file = tmp_path / "test-config.yml"
    yaml_file.write_text(yaml_content)

    # Try to load file
    result = config.load_config_file(yaml_file)

    # Expected result
    expected = {
        'verbose': expected_verbose,
        'workers_per_site': expected_workers,
    }

    # Assert result
    assert result == expected
