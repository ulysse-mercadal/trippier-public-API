# Security Policy

## Supported versions

Security fixes land on `main` and the most recent tagged release.

## Reporting a vulnerability

Do not open a public issue for security problems. Report privately via GitHub
Security Advisories:

https://github.com/ulysse-mercadal/trippier-public-API/security/advisories/new

Or email security@trippier.dev.

Include:

- a description of the issue and its impact,
- steps to reproduce (a proof of concept if you have one),
- the affected service (`auth-api`, `poi-api`, `itinerary-api`, or `frontend`).

We aim to acknowledge reports within 72 hours and to ship a fix or mitigation
for confirmed issues as fast as is practical.

## Scope

In scope: authentication and API-key handling (`auth-api`), the HMAC
service-to-service auth (`X-Internal-Auth`), rate-limit / quota bypass, secret
leakage, and injection or SSRF in the POI / itinerary provider pipelines.

Out of scope: issues requiring an already-compromised host or physical access,
availability of third-party BYOK providers, and denial of service from unbounded
self-hosted usage.
