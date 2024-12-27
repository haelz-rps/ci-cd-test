import copy
import os
from typing import MutableMapping

import pytest

@pytest.fixture
def valid_config() -> MutableMapping:
    return _valid_config()

def _valid_config() -> MutableMapping:
    return copy.deepcopy(
        {
            "base_url": "http://localhost",
            "per_page": 100,
            "offset": 0,
        }

    )

@pytest.fixture
def invalid_config() -> MutableMapping:
    return _invalid_config()

def _invalid_config() -> MutableMapping:
    return copy.deepcopy(
        {
            "base_url": "invalid-url.com",
            "per_page": -1,
            "offset": -1,
        }
    )

parametrized_configs = pytest.mark.parametrize(
    "config, is_valid",
    [
        (_valid_config(), True),
        (_invalid_config(), False),
    ],
)
