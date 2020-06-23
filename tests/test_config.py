'''Unit tests for config'''

import pytest

from delic.config import load_yaml_file


def test_load_yaml_file(tmp_path):
    '''Should read a YAML file'''
    # Write test file
    yaml_content = '''test:
                        success: True
                   '''
    yaml_file = tmp_path / "test-config.yml"
    yaml_file.write_text(yaml_content)

    # Try to load file
    result = load_yaml_file(yaml_file)

    # Expected result
    expected = {
        'test': {
            'success': True
        }
    }

    # Assert result
    assert result == expected
