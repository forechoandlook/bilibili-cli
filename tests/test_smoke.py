"""Integration smoke tests for bilibili-cli.

These tests invoke the real CLI commands with ``--yaml`` against the live
Bilibili API using your local browser cookies/saved session.  They are
**skipped by default** and only run when explicitly requested::

    uv run pytest -m smoke -v

Only read-only operations are tested — no writes.
"""

from __future__ import annotations

import pytest
import subprocess
import yaml

smoke = pytest.mark.smoke


def _invoke(*args: str):
    """Run a CLI command with --yaml and return parsed payload."""
    cmd = ["./bili", *args, "--yaml"]
    result = subprocess.run(cmd, capture_output=True, text=True)

    if result.stdout:
        try:
            payload = yaml.safe_load(result.stdout)
        except yaml.YAMLError:
            payload = None
    else:
        payload = None

    class ResultMock:
        def __init__(self, exit_code, output, stderr):
            self.exit_code = exit_code
            self.output = output
            self.stderr = stderr

    return ResultMock(result.returncode, result.stdout, result.stderr), payload


# ── Auth ────────────────────────────────────────────────────────────────


@smoke
class TestAuth:
    """Verify authentication is working end-to-end."""

    def test_status(self):
        result, payload = _invoke("status")
        # Go version fails cleanly if not authenticated, test that envelope is correct for error
        if result.exit_code != 0:
            assert payload["ok"] is False
            assert payload["error"]["code"] == "not_authenticated"
        else:
            assert payload["ok"] is True
            assert payload["data"]["authenticated"] is True

    def test_whoami(self):
        result, payload = _invoke("whoami")
        if result.exit_code != 0:
            assert payload["ok"] is False
            assert payload["error"]["code"] in ("not_authenticated", "upstream_error", "rate_limited")
        else:
            assert payload["ok"] is True
            assert payload["data"]["user"]


# ── Read-only queries ───────────────────────────────────────────────────


@smoke
class TestReadOnly:
    """Read-only CLI smoke tests."""

    def test_hot(self):
        result, payload = _invoke("hot")
        assert result.exit_code == 0, f"hot failed: {result.output}"
        assert payload["ok"] is True

    def test_rank(self):
        result, payload = _invoke("rank")
        # In strict environments, -352 might be returned due to lack of cookie or WBI context.
        if result.exit_code != 0:
            assert payload["ok"] is False
            assert payload["error"]["code"] in ["not_authenticated", "upstream_error", "network_error"]
        else:
            assert payload["ok"] is True

    def test_search(self):
        result, payload = _invoke("search", "Python 编程")
        if result.exit_code != 0:
            assert payload["ok"] is False
            assert payload["error"]["code"] in ("upstream_error", "rate_limited", "network_error")
        else:
            assert payload["ok"] is True
