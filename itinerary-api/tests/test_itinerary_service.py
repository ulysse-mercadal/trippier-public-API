"""Unit tests for ItineraryService — pure logic, no I/O."""

from __future__ import annotations

import pytest

from app.models.types import (
    Coordinates,
    ItineraryRequest,
    ItineraryResponse,
    Poi,
    PoiType,
    Preferences,
)
from app.services.itinerary_service import ItineraryService, _haversine, _parse_hours

# ── helpers ──────────────────────────────────────────────────────────────────

def make_poi(
    id: str,
    name: str,
    poi_type: PoiType = PoiType.see,
    lat: float | None = None,
    lng: float | None = None,
) -> Poi:
    coords = Coordinates(lat=lat, lng=lng) if lat is not None else None
    return Poi(id=id, name=name, type=poi_type, coords=coords)


@pytest.fixture
def service() -> ItineraryService:
    return ItineraryService()


# ── _haversine ────────────────────────────────────────────────────────────────

def test_haversine_same_point_is_zero():
    assert _haversine(48.85, 2.35, 48.85, 2.35) == pytest.approx(0.0)


def test_haversine_paris_to_london():
    # Paris (48.8566, 2.3522) → London (51.5074, -0.1278) ≈ 340 km
    dist = _haversine(48.8566, 2.3522, 51.5074, -0.1278)
    assert 330_000 < dist < 350_000, f"Expected ~340 km, got {dist:.0f} m"


def test_haversine_is_symmetric():
    d1 = _haversine(48.85, 2.35, 45.75, 4.85)
    d2 = _haversine(45.75, 4.85, 48.85, 2.35)
    assert d1 == pytest.approx(d2, rel=1e-9)


# ── _parse_hours ──────────────────────────────────────────────────────────────

def test_parse_hours_basic():
    assert _parse_hours("09:00", "21:00") == pytest.approx(12.0)


def test_parse_hours_with_minutes():
    assert _parse_hours("09:30", "18:00") == pytest.approx(8.5)


# ── _filter ──────────────────────────────────────────────────────────────────

def test_filter_removes_avoided_types(service: ItineraryService):
    pois = [
        make_poi("1", "Museum", PoiType.see),
        make_poi("2", "Restaurant", PoiType.eat),
        make_poi("3", "Bar", PoiType.drink),
    ]
    prefs = Preferences(avoid=[PoiType.eat, PoiType.drink])
    result = service._filter(pois, prefs)
    assert len(result) == 1
    assert result[0].type == PoiType.see


def test_filter_no_avoided_types_keeps_all(service: ItineraryService):
    pois = [make_poi(str(i), f"Place {i}") for i in range(5)]
    prefs = Preferences()
    result = service._filter(pois, prefs)
    assert len(result) == 5


def test_filter_empty_input(service: ItineraryService):
    result = service._filter([], Preferences(avoid=[PoiType.see]))
    assert result == []


# ── _sort_by_priority ─────────────────────────────────────────────────────────

def test_sort_by_priority_orders_correctly(service: ItineraryService):
    pois = [
        make_poi("1", "Bar", PoiType.drink),
        make_poi("2", "Museum", PoiType.see),
        make_poi("3", "Restaurant", PoiType.eat),
    ]
    result = service._sort_by_priority(pois, [PoiType.eat, PoiType.see, PoiType.drink])
    assert result[0].type == PoiType.eat
    assert result[1].type == PoiType.see
    assert result[2].type == PoiType.drink


def test_sort_by_priority_unprioritised_last(service: ItineraryService):
    pois = [
        make_poi("1", "Hotel", PoiType.sleep),
        make_poi("2", "Museum", PoiType.see),
    ]
    result = service._sort_by_priority(pois, [PoiType.see])
    assert result[0].type == PoiType.see
    assert result[1].type == PoiType.sleep


# ── _cluster ─────────────────────────────────────────────────────────────────

def test_cluster_returns_correct_number_of_days(service: ItineraryService):
    pois = [make_poi(str(i), f"POI {i}", lat=48.85 + i * 0.01, lng=2.35) for i in range(6)]
    clusters = service._cluster(pois, 3, None)
    assert len(clusters) == 3


def test_cluster_empty_pois(service: ItineraryService):
    clusters = service._cluster([], 3, start=None)
    assert len(clusters) == 3
    assert all(len(c) == 0 for c in clusters)


def test_cluster_all_pois_assigned(service: ItineraryService):
    pois = [make_poi(str(i), f"POI {i}", lat=48.0 + i * 0.1, lng=2.0) for i in range(9)]
    clusters = service._cluster(pois, 3, start=None)
    total = sum(len(c) for c in clusters)
    assert total == 9


# ── _nearest_neighbour ────────────────────────────────────────────────────────

def test_nearest_neighbour_orders_by_proximity(service: ItineraryService):
    # Three POIs roughly collinear west-to-east: we start at the western-most point.
    west  = make_poi("w", "West",   lat=48.85, lng=2.30)
    mid   = make_poi("m", "Middle", lat=48.85, lng=2.35)
    east  = make_poi("e", "East",   lat=48.85, lng=2.40)
    start = Coordinates(lat=48.85, lng=2.28)  # just west of "West"

    ordered = service._nearest_neighbour([east, mid, west], start)
    assert [p.id for p in ordered] == ["w", "m", "e"]


def test_nearest_neighbour_empty(service: ItineraryService):
    assert service._nearest_neighbour([], None) == []


# ── generate (integration of all steps) ──────────────────────────────────────

def test_generate_respects_days(service: ItineraryService):
    pois = [make_poi(str(i), f"Place {i}", lat=48.8 + i * 0.01, lng=2.3) for i in range(4)]
    req = ItineraryRequest(pois=pois, days=2)
    resp = service.generate(req, pois)
    assert isinstance(resp, ItineraryResponse)
    assert len(resp.days) == 2
    assert resp.total_pois == 4


def test_generate_filters_avoided(service: ItineraryService):
    pois = [
        make_poi("1", "Museum",     PoiType.see,   lat=48.85, lng=2.35),
        make_poi("2", "Restaurant", PoiType.eat,   lat=48.86, lng=2.36),
        make_poi("3", "Hotel",      PoiType.sleep, lat=48.84, lng=2.34),
    ]
    req = ItineraryRequest(
        pois=pois,
        days=1,
        preferences=Preferences(avoid=[PoiType.eat]),
    )
    resp = service.generate(req, pois)
    all_pois = [p for day in resp.days for p in day.pois]
    types_in_result = {p.type for p in all_pois}
    assert PoiType.eat not in types_in_result


def test_generate_empty_pois_returns_empty_days(service: ItineraryService):
    req = ItineraryRequest(pois=[], days=2)
    resp = service.generate(req, [])
    assert resp.total_pois == 0
    assert len(resp.days) == 2


def test_generate_pace_affects_duration(service: ItineraryService):
    pois = [make_poi(str(i), f"P {i}", lat=48.85, lng=2.35 + i * 0.01) for i in range(3)]

    req_relaxed   = ItineraryRequest(pois=pois, days=1, preferences=Preferences(pace="relaxed"))
    req_intensive = ItineraryRequest(pois=pois, days=1, preferences=Preferences(pace="intensive"))

    resp_relaxed   = service.generate(req_relaxed, pois)
    resp_intensive = service.generate(req_intensive, pois)

    dur_relaxed   = resp_relaxed.days[0].estimated_duration_hours
    dur_intensive = resp_intensive.days[0].estimated_duration_hours

    assert dur_relaxed >= dur_intensive
