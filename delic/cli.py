'''Module to handle CLI interactions'''

import argparse


def parse_args():
    '''Parse and return arguments'''
    parser = argparse.ArgumentParser()
    parser.add_argument(
        '-c', '--config',
        help='Location of the config file. Default: "./config.yml"',
        default='config.yml',
    )
    parser.add_argument(
        '-v', '--verbose',
        help='Enable verbose output',
        action='store_true',
    )
    return parser.parse_args()
