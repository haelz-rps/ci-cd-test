#
# Copyright (c) 2024 Airbyte, Inc., all rights reserved.
#


import pytest

pytest_plugins = ("connector_acceptance_test.plugin",)


@pytest.fixture(scope="session", autouse=True)
def connector_setup():
    """Fixture for setting up external resources required by acceptance tests."""
    # Placeholder for any setup actions required before running tests
    yield
    # Placeholder for any teardown actions after tests
