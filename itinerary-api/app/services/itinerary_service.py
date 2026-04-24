"""Core itinerary generation logic.

The service splits a flat list of POIs across N days by:
  1. Filtering out avoided types and sorting by user priorities.
  2. Clustering POIs geographically so each day is spatially coherent.
  3. Ordering each cluster with a nearest-neighbour heuristic.
  4. (Optional) Delegating natural-language summary generation to an LLM.
"""

from __future__ import annotations

import math

from app.models.types import (
    Coordinates,
    DayPlan,
    ItineraryRequest,
    ItineraryResponse,
    Poi,
    PoiType,
    Preferences,
)


class ItineraryService:
    """Builds day-by-day itineraries from a list of POIs."""

    def generate(self, request: ItineraryRequest, pois: list[Poi]) -> ItineraryResponse:
        pois = self._filter(pois, request.preferences)
        pois = self._sort_by_priority(pois, request.preferences.priorities)
        clusters = self._cluster(pois, request.days, request.start_location)
        day_plans = [
            self._build_day(day_index + 1, cluster, request.preferences)
            for day_index, cluster in enumerate(clusters)
        ]
        return ItineraryResponse(
            days=day_plans,
            total_pois=sum(len(d.pois) for d in day_plans),
        )

    def _filter(self, pois: list[Poi], prefs: Preferences) -> list[Poi]:
        avoided = set(prefs.avoid)
        return [p for p in pois if p.type not in avoided]

    def _sort_by_priority(self, pois: list[Poi], priorities: list[PoiType]) -> list[Poi]:
        priority_index = {t: i for i, t in enumerate(priorities)}
        return sorted(pois, key=lambda p: priority_index.get(p.type, len(priorities)))

    def _cluster(
        self,
        pois: list[Poi],
        days: int,
        start: Coordinates | None,
    ) -> list[list[Poi]]:
        if not pois:
            return [[] for _ in range(days)]

        pois_with_coords = [p for p in pois if p.coords and not p.coords.approximate]
        pois_no_coords = [p for p in pois if p not in pois_with_coords]

        clusters: list[list[Poi]] = [[] for _ in range(days)]
        centroids = self._initial_centroids(pois_with_coords, days, start)

        for _ in range(10):
            for c in clusters:
                c.clear()
            for poi in pois_with_coords:
                coords = poi.coords
                if coords is None:
                    continue
                dists = [
                    _haversine(coords.lat, coords.lng, centroids[j][0], centroids[j][1])
                    for j in range(days)
                ]
                clusters[dists.index(min(dists))].append(poi)
            for i, cluster in enumerate(clusters):
                if cluster:
                    lats = [p.coords.lat for p in cluster if p.coords is not None]
                    lngs = [p.coords.lng for p in cluster if p.coords is not None]
                    centroids[i] = (sum(lats) / len(cluster), sum(lngs) / len(cluster))

        for i, poi in enumerate(pois_no_coords):
            clusters[i % days].append(poi)

        return [self._nearest_neighbour(c, start) for c in clusters]

    def _initial_centroids(
        self,
        pois: list[Poi],
        k: int,
        start: Coordinates | None,
    ) -> list[tuple[float, float]]:
        if not pois:
            lat = start.lat if start else 0.0
            lng = start.lng if start else 0.0
            return [(lat, lng)] * k
        step = max(1, len(pois) // k)
        return [
            (c.lat, c.lng)
            for i in range(k)
            if (c := pois[i * step].coords) is not None
        ]

    def _nearest_neighbour(
        self,
        pois: list[Poi],
        start: Coordinates | None,
    ) -> list[Poi]:
        if not pois:
            return []
        remaining = list(pois)
        if start:
            cur_lat, cur_lng = start.lat, start.lng
        elif pois[0].coords:
            cur_lat, cur_lng = pois[0].coords.lat, pois[0].coords.lng
        else:
            return pois

        ordered: list[Poi] = []
        while remaining:
            nearest = min(
                remaining,
                key=lambda p: _haversine(cur_lat, cur_lng, p.coords.lat, p.coords.lng)
                if p.coords and not p.coords.approximate
                else float("inf"),
            )
            ordered.append(nearest)
            remaining.remove(nearest)
            if nearest.coords and not nearest.coords.approximate:
                cur_lat, cur_lng = nearest.coords.lat, nearest.coords.lng

        return ordered

    def _build_day(self, day: int, pois: list[Poi], prefs: Preferences) -> DayPlan:
        hours_available = _parse_hours(prefs.start_time, prefs.end_time)
        pace_factor = {"relaxed": 1.5, "moderate": 1.0, "intensive": 0.7}.get(prefs.pace, 1.0)
        duration = min(len(pois) * pace_factor, hours_available)
        return DayPlan(day=day, pois=pois, estimated_duration_hours=round(duration, 1))


def _haversine(lat1: float, lng1: float, lat2: float, lng2: float) -> float:
    earth_radius = 6_371_000.0
    lat1_rad = math.radians(lat1)
    lat2_rad = math.radians(lat2)
    delta_lat = math.radians(lat2 - lat1)
    delta_lng = math.radians(lng2 - lng1)
    a = (
        math.sin(delta_lat / 2) ** 2
        + math.cos(lat1_rad) * math.cos(lat2_rad) * math.sin(delta_lng / 2) ** 2
    )
    return earth_radius * 2 * math.atan2(math.sqrt(a), math.sqrt(1 - a))


def _parse_hours(start: str, end: str) -> float:
    sh, sm = map(int, start.split(":"))
    eh, em = map(int, end.split(":"))
    return (eh * 60 + em - sh * 60 - sm) / 60
