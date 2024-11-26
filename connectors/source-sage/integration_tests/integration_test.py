import json
import os
import pytest
import requests
from source_sage.source import SourceSage, SageStream

# Define known issues
known_issues = ["DL02000001"]


@pytest.fixture
def valid_config():
    with open(
        os.path.join(os.path.dirname(__file__), "../secrets/secrets.json"), "r"
    ) as f:
        return json.load(f)


@pytest.fixture
def stream(valid_config):
    return SageStream(name="CUSTOMER", **valid_config)


# Integration Test: Full Data Retrieval
def test_full_refresh_data_retrieval(stream):
    records = list(stream.read_records(sync_mode="full_refresh"))
    assert len(records) > 0, "No records retrieved in full refresh mode"
    assert "RECORDNO" in records[0], "Expected 'RECORDNO' field missing from records"


# Integration Test: Incremental Sync
def test_incremental_sync(stream):
    # Set initial state
    initial_state = {"RECORDNO": "1000"}
    records = list(
        stream.read_records(sync_mode="incremental", stream_state=initial_state)
    )
    assert len(records) > 0, "No records retrieved in incremental sync"
    assert all(
        record["RECORDNO"] > "1000" for record in records
    ), "Incremental sync returned older records"


@pytest.mark.integration
def test_api_connection(valid_config):
    """Test that the connection to the API is successful with valid credentials."""
    source = SourceSage()
    success, error = source.check_connection(logger=None, config=valid_config)
    assert success, f"Connection failed with error: {error}"


@pytest.mark.integration
def test_live_response_data(valid_config):
    """Ensure that live data is returned in the response structure."""
    stream = SageStream(name="CUSTOMER", **valid_config)
    stream._authenticate()  # This will hit the API with live credentials

    # Check if the session_id attribute is set
    assert stream._session_id is not None, "Live API did not return a session ID."


@pytest.mark.integration
def test_data_retrieval_full_refresh(valid_config):
    """Test retrieving data in 'full_refresh' mode."""
    stream = SageStream(name="CUSTOMER", **valid_config)
    response = stream.read_records(sync_mode="full_refresh")
    data = list(response)
    assert len(data) > 0, "No data retrieved in full_refresh mode"
    assert isinstance(data[0], dict), "Expected data records to be of dict type"
    assert "RECORDNO" in data[0], "Expected 'RECORDNO' field in data record"


@pytest.mark.integration
def test_data_retrieval_incremental(valid_config):
    """Test retrieving data in 'incremental' mode."""
    stream = SageStream(name="CUSTOMER", **valid_config)
    response = stream.read_records(
        sync_mode="incremental", stream_state={"RECORDNO": 1000}
    )
    data = list(response)
    assert len(data) > 0, "No data retrieved in incremental mode"
    assert isinstance(data[0], dict), "Expected data records to be of dict type"
    assert "RECORDNO" in data[0], "Expected 'RECORDNO' field in data record"


@pytest.mark.integration
def test_response_format_validation(stream):
    try:
        # Retrieve records and check the expected format and fields in the response
        response = stream.read_records(sync_mode="full_refresh")
        data = list(response)
        assert len(data) > 0, "No data retrieved from the API"

        # Check essential fields in each record
        for record in data:
            assert isinstance(record, dict), "Expected record to be a dictionary"
            assert "RECORDNO" in record, "Expected 'RECORDNO' field in record"
            assert isinstance(
                record.get("RECORDNO"), (int, str)
            ), "'RECORDNO' should be of type int or str"

            # Optional: Convert to int if necessary
            record_no = record.get("RECORDNO")
            if isinstance(record_no, str):
                assert (
                    record_no.isdigit()
                ), "'RECORDNO' string should represent an integer"
                record_no = int(record_no)
    except requests.RequestException as e:
        pytest.fail(f"Response format validation failed: {str(e)}")


@pytest.mark.integration
def test_data_pagination(stream):
    try:
        # Retrieve all records page by page using built-in pagination
        all_data = []
        page_count = 0

        # Stream records in full_refresh mode, letting `read_records` handle pagination
        for record in stream.read_records(sync_mode="full_refresh"):
            all_data.append(record)
            page_count += 1

        assert len(all_data) > 0, "No data retrieved from paginated API request"
        assert page_count > 1, "Expected multiple pages of data for pagination test"

    except requests.RequestException as e:
        pytest.fail(f"Pagination test failed due to request error: {str(e)}")


@pytest.mark.integration
def test_live_request_body_data_with_whenmodified(stream):
    """Test that WHENMODIFIED state is correctly sent to the API in live requests."""
    stream_state = {"WHENMODIFIED": "2023-01-15T10:30:00"}
    request_body = stream.request_body_data(stream_state=stream_state)

    # Verify the correct timestamp is in the request XML
    assert (
        "2023-01-15T10:30:00" in request_body
    ), "Expected WHENMODIFIED timestamp missing in request body."
