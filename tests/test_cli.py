'''Unit tests for cli'''

import pytest

from delic.cli import parse_args


# ================================
# =        SecretToken        =
# ================================

@pytest.mark.parametrize("config_in,config_out", [
    ([], '/config.yml'),
    (['-c', './test-config.yml'], './test-config.yml'),
    (['--config', './test-config.yml'], './test-config.yml'),
])
@pytest.mark.parametrize("verbose_in,verbose_out", [
    ([], False),
    (['-v'], True),
    (['--verbose'], True),
])
def test_args(verbose_in, verbose_out, config_in, config_out):
    '''Tests if the arguments are parsed as expected'''
    # Build args input
    args = [
        *config_in,
        *verbose_in,
    ]

    # Parse args
    result = parse_args(args)

    # Assert results
    assert result.config == config_out
    assert result.verbose == verbose_out
