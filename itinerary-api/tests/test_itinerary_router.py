"""Integration tests for the /itinerary/generate router.

These tests spin up a minimal FastAPI app with only the itinerary router
(no rate-limit middleware) so they focus on routing, validation, and
dependency injection rather than middleware behaviour.
"""

from __future__ import annotations

from unittest.mock import AsyncMock, MagicMock

import pytest
from fastapi import FastAPI
from fastapi.testclient import TestClient

from app.config import get_itinerary_service, get_poi_client
from app.models.types import (
    Coordinates,
    DayPlan,
    ItineraryResponse,
    Poi,
    PoiType,
)
from app.routers.itinerary import router
from app.services.itinerary_service import ItineraryService

# ── fixtures ──────────────────────────────────────────────────────────────────

@pytest.fixture()
def app() -> FastAPI:
    """Minimal app: only the itinerary router, no middleware."""
    a = FastAPI()
    a.include_router(router)
    return a


@pytest.fixture()
def client(app: FastAPI) -> TestClient:
    return TestClient(app)


# ── direct pois in body ───────────────────────────────────────────────────────

def test_generate_with_direct_pois_returns_200(client: TestClient) -> None:
    resp = client.post("/itinerary/generate", json={
        "pois": [
            {"id": "1", "name": "Eiffel Tower", "type": "see",
             "coords": {"lat": 48.858, "lng": 2.294}},
            {"id": "2", "name": "Louvre", "type": "see",
             "coords": {"lat": 48.860, "lng": 2.337}},
        ],
        "days": 1,
    })
    assert resp.status_code == 200
    data = resp.json()
    assert "days" in data
    assert data["total_pois"] == 2


def test_generate_correct_number_of_days(client: TestClient) -> None:
    pois = [
        {"id": str(i), "name": f"Place {i}", "type": "see",
         "coords": {"lat": 48.85 + i * 0.01, "lng": 2.35}}
        for i in range(6)
    ]
    resp = client.post("/itinerary/generate", json={"pois": pois, "days": 3})
    assert resp.status_code == 200
    data = resp.json()
    assert len(data["days"]) == 3


def test_generate_respects_avoided_types(client: TestClient) -> None:
    resp = client.post("/itinerary/generate", json={
        "pois": [
            {"id": "1", "name": "Museum", "type": "see",
             "coords": {"lat": 48.85, "lng": 2.35}},
            {"id": "2", "name": "Restaurant", "type": "eat",
             "coords": {"lat": 48.86, "lng": 2.36}},
        ],
        "days": 1,
        "preferences": {"avoid": ["eat"]},
    })
    assert resp.status_code == 200
    data = resp.json()
    all_types = [p["type"] for day in data["days"] for p in day["pois"]]
    assert "eat" not in all_types


# ── validation errors ─────────────────────────────────────────────────────────

def test_generate_neither_pois_nor_query_returns_422(client: TestClient) -> None:
    resp = client.post("/itinerary/generate", json={"days": 1})
    assert resp.status_code == 422


def test_generate_empty_pois_without_query_returns_404(client: TestClient) -> None:
    resp = client.post("/itinerary/generate", json={"pois": [], "days": 1})
    assert resp.status_code == 404


def test_generate_days_too_large_returns_422(client: TestClient) -> None:
    resp = client.post("/itinerary/generate", json={
        "pois": [{"id": "1", "name": "P", "type": "see"}],
        "days": 31,  # max is 30
    })
    assert resp.status_code == 422


def test_generate_invalid_pace_returns_422(client: TestClient) -> None:
    resp = client.post("/itinerary/generate", json={
        "pois": [{"id": "1", "name": "P", "type": "see",
                  "coords": {"lat": 48.85, "lng": 2.35}}],
        "days": 1,
        "preferences": {"pace": "sprint"},  # invalid
    })
    assert resp.status_code == 422


# ── poi_query path (mocked poi_client) ───────────────────────────────────────

def test_generate_with_poi_query_calls_poi_client(app: FastAPI) -> None:
    mock_client = AsyncMock()
    mock_client.search.return_value = [
        Poi(
            id="overpass:1",
            name="Sacré-Cœur",
            type=PoiType.see,
            coords=Coordinates(lat=48.887, lng=2.343),
        )
    ]
    app.dependency_overrides[get_poi_client] = lambda: mock_client

    c = TestClient(app)
    resp = c.post("/itinerary/generate", json={
        "poi_query": {"lat": 48.887, "lng": 2.343, "radius": 1000},
        "days": 1,
    })

    app.dependency_overrides.clear()

    assert resp.status_code == 200
    mock_client.search.assert_called_once()
    data = resp.json()
    assert data["total_pois"] == 1


def test_generate_poi_query_empty_result_returns_404(app: FastAPI) -> None:
    mock_client = AsyncMock()
    mock_client.search.return_value = []
    app.dependency_overrides[get_poi_client] = lambda: mock_client

    c = TestClient(app)
    resp = c.post("/itinerary/generate", json={
        "poi_query": {"lat": 0.0, "lng": 0.0},
        "days": 1,
    })

    app.dependency_overrides.clear()

    assert resp.status_code == 404


# ── service dependency override ───────────────────────────────────────────────

def test_generate_uses_injected_service(app: FastAPI) -> None:
    """Verify that the router delegates to the injected ItineraryService."""
    fake_service = MagicMock(spec=ItineraryService)
    fake_service.generate.return_value = ItineraryResponse(
        days=[DayPlan(day=1, pois=[], estimated_duration_hours=0.0)],
        total_pois=0,
    )
    app.dependency_overrides[get_itinerary_service] = lambda: fake_service

    c = TestClient(app)
    resp = c.post("/itinerary/generate", json={
        "pois": [{"id": "1", "name": "P", "type": "see",
                  "coords": {"lat": 48.85, "lng": 2.35}}],
        "days": 1,
    })

    app.dependency_overrides.clear()

    assert resp.status_code == 200
    assert fake_service.generate.called
