"""DSN / TLS parameter construction for client databases (no network)."""
import glob
import os
import stat
import tempfile

import pytest
from psycopg.conninfo import conninfo_to_dict

from core import db
from core.db import (
    SSLConfigError,
    build_client_conn_string,
    client_postgres_conninfo,
    resolve_ssl_mode,
)


def gen_ca_pem() -> str:
    from cryptography import x509
    from cryptography.hazmat.primitives import hashes, serialization
    from cryptography.hazmat.primitives.asymmetric import ec
    import datetime

    key = ec.generate_private_key(ec.SECP256R1())
    name = x509.Name([x509.NameAttribute(x509.NameOID.COMMON_NAME, "test-ca")])
    now = datetime.datetime.now(datetime.timezone.utc)
    cert = (
        x509.CertificateBuilder().subject_name(name).issuer_name(name)
        .public_key(key.public_key()).serial_number(x509.random_serial_number())
        .not_valid_before(now - datetime.timedelta(days=1))
        .not_valid_after(now + datetime.timedelta(days=30))
        .add_extension(x509.BasicConstraints(ca=True, path_length=None), critical=True)
        .sign(key, hashes.SHA256())
    )
    return cert.public_bytes(serialization.Encoding.PEM).decode()


def src(**kw):
    base = {"source_type": "postgres", "host": "db.example.com", "port": 5432,
            "database_name": "core", "username": "ro", "password_encrypted": "pw"}
    base.update(kw)
    return base


def parse(src_):
    with client_postgres_conninfo(src_) as c:
        return conninfo_to_dict(c)


def test_legacy_boolean_mapping():
    assert resolve_ssl_mode(None, True) == "require"
    assert resolve_ssl_mode(None, None) == "require"   # NULL column -> secure default
    assert resolve_ssl_mode("", False) == "disable"
    assert resolve_ssl_mode("verify-full", False) == "verify-full"  # ssl_mode authoritative
    assert resolve_ssl_mode(" Verify-CA ", True) == "verify-ca"


def test_bad_mode_rejected():
    with pytest.raises(SSLConfigError):
        resolve_ssl_mode("prefer", True)
    with pytest.raises(SSLConfigError):
        parse(src(ssl_mode="allow"))


@pytest.mark.parametrize("host", ["localhost", "127.0.0.1", "cbs-simulator", "datalake-db", "host.docker.internal"])
def test_local_hosts_no_longer_forced_to_disable(host):
    assert parse(src(host=host, ssl_enabled=True))["sslmode"] == "require"
    assert parse(src(host=host, ssl_mode="require"))["sslmode"] == "require"
    # explicit disable is still honoured
    assert parse(src(host=host, ssl_mode="disable"))["sslmode"] == "disable"
    assert parse(src(host=host, ssl_enabled=False))["sslmode"] == "disable"


def test_disable_and_require_have_no_rootcert():
    for m in ("disable", "require"):
        assert "sslrootcert" not in parse(src(ssl_mode=m))


def test_cert_with_non_verify_mode_rejected():
    with pytest.raises(SSLConfigError):
        parse(src(ssl_mode="require", ssl_root_cert=gen_ca_pem()))


def test_invalid_pem_rejected():
    with pytest.raises(SSLConfigError):
        parse(src(ssl_mode="verify-full", ssl_root_cert="not a cert"))


def test_verify_with_custom_ca_uses_0600_temp_file_and_cleans_up():
    pem = gen_ca_pem()
    before = set(glob.glob(os.path.join(tempfile.gettempdir(), "dh-ca-*")))
    with client_postgres_conninfo(src(ssl_mode="verify-full", ssl_root_cert=pem)) as c:
        d = conninfo_to_dict(c)
        assert d["sslmode"] == "verify-full"
        path = d["sslrootcert"]
        assert os.path.isfile(path)
        assert stat.S_IMODE(os.stat(path).st_mode) == 0o600
        assert open(path).read().strip() == pem.strip()
    assert not os.path.exists(path)
    assert set(glob.glob(os.path.join(tempfile.gettempdir(), "dh-ca-*"))) == before


def test_ca_file_cleaned_up_on_error():
    pem = gen_ca_pem()
    paths = []
    with pytest.raises(RuntimeError):
        with client_postgres_conninfo(src(ssl_mode="verify-ca", ssl_root_cert=pem)) as c:
            paths.append(conninfo_to_dict(c)["sslrootcert"])
            raise RuntimeError("boom")
    assert paths and not os.path.exists(paths[0])


@pytest.mark.parametrize("mode", ["verify-ca", "verify-full"])
def test_verify_without_ca_uses_system_pool(mode, monkeypatch):
    monkeypatch.setattr(db, "_system_root_ref", lambda m: "/etc/ssl/cert.pem" if m == "verify-ca" else "system")
    d = parse(src(ssl_mode=mode))
    assert d["sslrootcert"] == ("system" if mode == "verify-full" else "/etc/ssl/cert.pem")


def test_verify_without_ca_and_no_bundle_fails_loudly(monkeypatch):
    monkeypatch.setattr(db.psycopg.pq, "version", lambda: 150000)
    monkeypatch.setattr(db.ssl, "get_default_verify_paths",
                        lambda: type("P", (), {"cafile": None, "openssl_cafile": "/nonexistent"})())
    monkeypatch.setitem(__import__("sys").modules, "certifi", None)
    with pytest.raises(SSLConfigError):
        parse(src(ssl_mode="verify-full"))


def test_credentials_are_escaped():
    nasty = "p w'd\\ \"x\"=@:/?#"
    d = parse(src(password_encrypted=nasty, username="u ser'", database_name="my db", ssl_mode="disable"))
    assert d["password"] == nasty
    assert d["user"] == "u ser'"
    assert d["dbname"] == "my db"


def test_build_client_conn_string_postgres_and_other_engines():
    d = conninfo_to_dict(build_client_conn_string(src(ssl_mode="require")))
    assert d["host"] == "db.example.com" and d["sslmode"] == "require"
    with pytest.raises(SSLConfigError):  # CA without a file path must not be silently dropped
        build_client_conn_string(src(ssl_mode="verify-ca", ssl_root_cert=gen_ca_pem()))
    url = build_client_conn_string(src(source_type="mysql", password_encrypted="p@ss:/w"))
    assert url == "mysql+pymysql://ro:p%40ss%3A%2Fw@db.example.com:5432/core"


def test_registration_model_ssl_validation():
    pytest.importorskip("fastapi")
    from pydantic import ValidationError
    import sys, types
    # routers.sources imports `main` (the app); stub it so only the model is exercised
    sys.modules.setdefault("main", types.SimpleNamespace(engine=None))
    from routers.sources import SourceRegistration

    base = dict(source_id="s", name="n", host="h", database_name="d")
    r = SourceRegistration(**base)
    assert (r.ssl_mode, r.ssl_enabled) == ("require", True)
    r = SourceRegistration(**base, ssl_enabled=False)
    assert (r.ssl_mode, r.ssl_enabled) == ("disable", False)
    r = SourceRegistration(**base, ssl_mode="verify-full", ssl_enabled=False)
    assert (r.ssl_mode, r.ssl_enabled) == ("verify-full", True)
    r_null_lists = SourceRegistration(**base, scopes=None, data_mappings=None)
    assert r_null_lists.scopes == [] and r_null_lists.data_mappings == []
    with pytest.raises(ValidationError):
        SourceRegistration(**base, ssl_mode="prefer")
    with pytest.raises(ValidationError):
        SourceRegistration(**base, ssl_mode="require", ssl_root_cert=gen_ca_pem())
