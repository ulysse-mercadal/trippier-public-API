"""Tests for the RateLimitMiddleware that delegates to auth-api."""

from __future__ import annotations

import httpx
import pytest
import respx
from fastapi import FastAPI
from fastapi.testclient import TestClient

from app.middleware.ratelimit import RateLimitMiddleware

AUTH_URL = "http://fake-auth-api"
SECRET = "test-internal-secret"
RATE_LIMIT_ENDPOINT = f"{AUTH_URL}/internal/check-rate-limit"


def make_app(auth_url: str = AUTH_URL) -> FastAPI:
    """Build a minimal FastAPI app with the RateLimitMiddleware attached."""
    app = FastAPI()
    app.add_middleware(
        RateLimitMiddleware,
        auth_api_url=auth_url,
        internal_secret=SECRET,
        cost=10,
    )

    @app.get("/health")
    def health() -> dict[str, str]:
        return {"status": "ok"}

    @app.get("/search")
    def search() -> dict[str, list[str]]:
        return {"results": []}

    return app


@pytest.fixture
def client() -> TestClient:
    return TestClient(make_app(), raise_server_exceptions=False)


# ── health is exempt ──────────────────────────────────────────────────────────

def test_health_is_exempt(client: TestClient) -> None:
    """GET /health must pass without X-API-Key and without calling auth-api."""
    resp = client.get("/health")
    assert resp.status_code == 200


# ── missing key ───────────────────────────────────────────────────────────────

def test_missing_api_key_returns_401(client: TestClient) -> None:
    resp = client.get("/search")
    assert resp.status_code == 401
    assert "X-API-Key" in resp.json()["error"]


# ── allowed ───────────────────────────────────────────────────────────────────

@respx.mock
def test_allowed_request_passes(client: TestClient) -> None:
    route = respx.post(RATE_LIMIT_ENDPOINT).mock(
        return_value=httpx.Response(
            200,
            json={"allowed": True, "remaining": 90, "limit": 100, "resets_in_secs": 3600},
        )
    )
    resp = client.get("/search", headers={"X-API-Key": "trp_valid_key"})
    assert resp.status_code == 200
    assert resp.headers.get("X-RateLimit-Remaining") == "90"

    # Verify HMAC header was sent (not the raw secret)
    sent_header = route.calls.last.request.headers.get("x-internal-auth", "")
    assert "." in sent_header, f"X-Internal-Auth should be '<ts>.<hmac>', got: {sent_header!r}"
    ts_part, sig_part = sent_header.split(".", 1)
    assert ts_part.isdigit(), "timestamp part must be numeric"
    assert len(sig_part) == 64, "signature should be sha256 hex (64 chars)"


# ── rate limit exceeded ───────────────────────────────────────────────────────

@respx.mock
def test_rate_limit_exceeded_returns_429(client: TestClient) -> None:
    respx.post(RATE_LIMIT_ENDPOINT).mock(
        return_value=httpx.Response(
            200,
            json={
                "allowed": False,
                "remaining": 0,
                "limit": 100,
                "resets_in_secs": 300,
                "error": "rate limit exceeded",
            },
        )
    )
    resp = client.get("/search", headers={"X-API-Key": "trp_exhausted_key"})
    assert resp.status_code == 429


# ── invalid key ───────────────────────────────────────────────────────────────

@respx.mock
def test_invalid_api_key_returns_401(client: TestClient) -> None:
    respx.post(RATE_LIMIT_ENDPOINT).mock(
        return_value=httpx.Response(
            401,
            json={"allowed": False, "error": "invalid api key"},
        )
    )
    resp = client.get("/search", headers={"X-API-Key": "trp_bad_key"})
    assert resp.status_code == 401


# ── auth-api unreachable ──────────────────────────────────────────────────────

def test_auth_api_down_returns_503() -> None:
    """When auth-api is unreachable the middleware must return 503."""
    app = make_app(auth_url="http://localhost:1")  # port 1 = always refused
    c = TestClient(app, raise_server_exceptions=False)
    resp = c.get("/search", headers={"X-API-Key": "trp_some_key"})
    assert resp.status_code == 503
