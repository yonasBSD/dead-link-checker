'''Unit tests for config'''

import pytest
import logging
from argparse import Namespace

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

    # Set args
    args = Namespace(
        verbose=False,
    )

    # Try to load file
    result = config.load_config_file(yaml_file, args)

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


def get_boolean_map(default):
    '''Return a boolean map with a default'''
    return [
        (False, False),
        (True, True),
        ('test', default),
        (0, default),
        (1, default),
    ]


@pytest.mark.parametrize('args,verbose_arg', [
    (Namespace(verbose=True), True),
    (Namespace(verbose=False), False),
])
@pytest.mark.parametrize('verbose_config, expected_verbose', get_boolean_map(True))
@pytest.mark.parametrize('worker_config, expected_workers', [
    (-10, 8),
    (-1, 8),
    (0, 8),
    (1, 1),
    (4, 4),
    ('A', 8),
])
@pytest.mark.parametrize('internal_only_config, expected_internal_only', get_boolean_map(False))
def test_load_valid_config_file(internal_only_config,
                                expected_internal_only,
                                worker_config,
                                expected_workers,
                                verbose_config,
                                expected_verbose,
                                args,
                                verbose_arg,
                                tmp_path):
    '''Should read a YAML file and update with defaults'''
    # Write test file
    yaml_content = f'''
                    verbose: {verbose_config}
                    workers_per_site: {worker_config}
                    internal_links_only: {internal_only_config}
                    '''
    yaml_file = tmp_path / "test-config.yml"
    yaml_file.write_text(yaml_content)

    # Try to load file
    result = config.load_config_file(yaml_file, args)

    # Expected result
    expected = {
        'verbose': verbose_arg or expected_verbose,
        'workers_per_site': expected_workers,
        'internal_links_only': expected_internal_only
    }

    # Assert result
    assert result == expected
