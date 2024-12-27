import json
import os
import pytest
from airbyte_cdk.sources.types import StreamSlice
from source_wordpress.source import SourceWordpress

@pytest.fixture
def valid_config():
    with open(
        os.path.join(os.path.dirname(__file__), "../secrets/config.json"), "r"
    ) as f:
        return json.load(f)   

@pytest.fixture
def stream(valid_config):
    source = SourceWordpress()
    streams = source.streams(valid_config)
    return streams[0]


def test_data_retrieval_incremental(stream):
    response = stream.read_records(
        sync_mode="incremental",
        stream_slice=StreamSlice(partition={}, cursor_slice={"start_time": "2000-01-01T00:00:00", "end_time": "2024-12-15T23:59:59"}))
    records = list(response)
    assert len(records) > 0, "No data retrieved in incremental mode"

def test_api_connection(valid_config):
    """Test that the connection to the API is successful with valid credentials."""
    source = SourceWordpress()
    success = source.check(logger=None, config=valid_config)
    print(success)
    assert success, f"Connection failed with error: {error}"
