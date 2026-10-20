"""Tests for the python-ai fail-closed auth. Run: python -m unittest test_security (needs fastapi and httpx)."""

import unittest

from fastapi import FastAPI, Request
from fastapi.testclient import TestClient

from security import install_api_key_middleware, internal_call_allowed, token_matches, unauthorized

# Synthetic test values, never real keys.
API_KEY = "test-api-key-" + "a" * 50
INTERNAL = "test-internal-" + "b" * 50


def build_client(api_key, internal_token):
    app = FastAPI()
    install_api_key_middleware(app, lambda: api_key)

    @app.get("/health")
    def health():
        return {"status": "ok"}

    @app.post("/predict/risk-score")
    def predict():
        return {"score": 1}

    @app.post("/retrain/auto")
    def retrain(request: Request):
        if not internal_call_allowed(request, lambda: internal_token):
            return unauthorized("Internal token required")
        return {"status": "scheduled"}

    return TestClient(app)


class TokenMatchesTest(unittest.TestCase):
    def test_empty_expected_never_matches(self):
        self.assertFalse(token_matches("", ""))
        self.assertFalse(token_matches("anything", ""))

    def test_missing_provided_never_matches(self):
        self.assertFalse(token_matches(None, API_KEY))
        self.assertFalse(token_matches("", API_KEY))

    def test_exact_match_only(self):
        self.assertTrue(token_matches(API_KEY, API_KEY))
        self.assertFalse(token_matches(API_KEY[:-1], API_KEY))
        self.assertFalse(token_matches(API_KEY + "x", API_KEY))


class ApiKeyMiddlewareTest(unittest.TestCase):
    def setUp(self):
        self.client = build_client(API_KEY, INTERNAL)

    def test_health_stays_public(self):
        self.assertEqual(self.client.get("/health").status_code, 200)

    def test_missing_key_is_401_and_handler_not_run(self):
        self.assertEqual(self.client.post("/predict/risk-score").status_code, 401)

    def test_wrong_key_is_401(self):
        r = self.client.post("/predict/risk-score", headers={"X-API-Key": "wrong"})
        self.assertEqual(r.status_code, 401)

    def test_valid_key_passes(self):
        r = self.client.post("/predict/risk-score", headers={"X-API-Key": API_KEY})
        self.assertEqual(r.status_code, 200)

    def test_unset_api_key_fails_closed(self):
        closed = build_client("", INTERNAL)
        self.assertEqual(closed.post("/predict/risk-score").status_code, 401)
        self.assertEqual(closed.post("/predict/risk-score", headers={"X-API-Key": ""}).status_code, 401)
        self.assertEqual(closed.post("/predict/risk-score", headers={"X-API-Key": "anything"}).status_code, 401)


class RetrainAutoTest(unittest.TestCase):
    def test_requires_api_key_and_internal_token(self):
        client = build_client(API_KEY, INTERNAL)
        self.assertEqual(client.post("/retrain/auto").status_code, 401)
        self.assertEqual(client.post("/retrain/auto", headers={"X-API-Key": API_KEY}).status_code, 401)
        self.assertEqual(
            client.post("/retrain/auto", headers={"X-API-Key": API_KEY, "X-Internal-Token": "wrong"}).status_code,
            401,
        )
        ok = client.post("/retrain/auto", headers={"X-API-Key": API_KEY, "X-Internal-Token": INTERNAL})
        self.assertEqual(ok.status_code, 200)

    def test_unset_internal_token_disables_retrain(self):
        client = build_client(API_KEY, "")
        r = client.post("/retrain/auto", headers={"X-API-Key": API_KEY, "X-Internal-Token": ""})
        self.assertEqual(r.status_code, 401)


if __name__ == "__main__":
    unittest.main()
