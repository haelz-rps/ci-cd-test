import logging

import pytest
from source_wordpress import SourceWordpress

from .conftest import parametrized_configs

@parametrized_configs
def test_streams(config, is_valid):
    source = SourceWordpress()
    if is_valid:
        streams = source.streams(config)
        assert len(streams) == 1
    else:
        with pytest.raises(Exception) as exc_info:
            _ = source.streams(config)
            assert "The path from `authenticator_selection_path` is not found in the config." in repr(exc_info.value)
