"""Fail-closed request authentication for the python-ai service.

Every route except /health needs a valid X-API-Key. If API_KEY is not configured,
no key is valid, so the service refuses requests instead of running open.
Internal-only routes such as /retrain/auto also need X-Internal-Token, which must
match INTERNAL_RETRAIN_TOKEN. If that is unset, those routes refuse every call.

This module imports only the standard library and fastapi, so it can be tested
without loading the ML stack.
"""

import hmac

from fastapi import Request
from fastapi.responses import JSONResponse

PUBLIC_PATHS = frozenset({"/health"})
API_KEY_HEADER = "X-API-Key"
INTERNAL_TOKEN_HEADER = "X-Internal-Token"


def token_matches(provided, expected):
    """Constant-time comparison. An empty expected value or a missing header never matches."""
    if not expected or not provided:
        return False
    return hmac.compare_digest(provided.encode("utf-8"), expected.encode("utf-8"))


def unauthorized(detail="Invalid or missing API key"):
    return JSONResponse(status_code=401, content={"detail": detail})


def install_api_key_middleware(app, get_api_key):
    """Require X-API-Key on every path except PUBLIC_PATHS. get_api_key is read per request."""

    @app.middleware("http")
    async def verify_api_key(request: Request, call_next):
        if request.url.path in PUBLIC_PATHS:
            return await call_next(request)
        if not token_matches(request.headers.get(API_KEY_HEADER), get_api_key()):
            return unauthorized()
        return await call_next(request)

    return verify_api_key


def internal_call_allowed(request: Request, get_internal_token):
    """True only when X-Internal-Token matches the configured internal token."""
    return token_matches(request.headers.get(INTERNAL_TOKEN_HEADER), get_internal_token())
