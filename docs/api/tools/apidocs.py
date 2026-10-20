#!/usr/bin/env python3
"""
AuditSphere API documentation generator.

Builds docs/api/openapi.yaml (OpenAPI 3.1) and docs/api/inventory.{md,json}
from the source of truth:

  * the Gin route registrations of every Go service (config.yaml lists the files),
  * the Go structs those handlers bind and return (request/response schemas),
  * swag annotations where a handler still has them,
  * the production Kong declarative config (public paths, what is routed),
  * python-ai's own FastAPI OpenAPI document (snapshot) and its decorators,
  * the Nuxt server routes (frontend/server/api),
  * every API call the website makes (frontend stores/composables/pages/...),
  * hand-written overlays in docs/api/overlays/*.yaml for what code can't say.

Commands:
  apidocs.py generate            write openapi.yaml + inventory.md + inventory.json
  apidocs.py check               regenerate in memory; exit 1 if the committed files
                                 are stale, an overlay points at a route that no longer
                                 exists, a Go file registers routes but is not listed in
                                 config.yaml, or the website calls an endpoint that is not
                                 registered/routed and is not listed as a known issue
  apidocs.py refresh-python-ai [--url URL]
                                 re-snapshot python-ai's /openapi.json (local stack)

Only the standard library and PyYAML are needed.
"""
from __future__ import annotations

import argparse
import copy
import fnmatch
import json
import re
import sys
import urllib.request
from collections import OrderedDict, defaultdict
from pathlib import Path

import yaml

ROOT = Path(__file__).resolve().parents[3]
DOCS = ROOT / "docs" / "api"
TOOLS = DOCS / "tools"
OVERLAYS = DOCS / "overlays"

HTTP_METHODS = ["GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"]
METHOD_ORDER = {m: i for i, m in enumerate(HTTP_METHODS)}

STATUS_NAMES = {
    "StatusOK": 200, "StatusCreated": 201, "StatusAccepted": 202, "StatusNoContent": 204,
    "StatusPartialContent": 206, "StatusMovedPermanently": 301, "StatusFound": 302,
    "StatusSeeOther": 303, "StatusTemporaryRedirect": 307, "StatusNotModified": 304,
    "StatusBadRequest": 400, "StatusUnauthorized": 401, "StatusForbidden": 403,
    "StatusNotFound": 404, "StatusMethodNotAllowed": 405, "StatusConflict": 409,
    "StatusGone": 410, "StatusRequestEntityTooLarge": 413, "StatusUnsupportedMediaType": 415,
    "StatusUnprocessableEntity": 422, "StatusTooManyRequests": 429,
    "StatusInternalServerError": 500, "StatusNotImplemented": 501, "StatusBadGateway": 502,
    "StatusServiceUnavailable": 503, "StatusGatewayTimeout": 504,
}
RESPONSE_HELPER_STATUS = {
    "BadRequest": 400, "Unauthorized": 401, "Forbidden": 403, "NotFound": 404,
    "InternalServerError": 500, "Conflict": 409,
}
ERROR_TEXT = {
    400: "Invalid request (bad id, body or parameter)", 401: "Missing or invalid token",
    403: "Authenticated but not allowed (role/permission)", 404: "Not found",
    405: "Method not allowed", 409: "Conflict (duplicate or invalid state transition)",
    413: "Payload too large", 422: "Validation failed", 429: "Rate limited by Kong",
    500: "Server error", 502: "Upstream error", 503: "Dependency unavailable",
}


# ---------------------------------------------------------------------------
# small helpers
# ---------------------------------------------------------------------------

def load_yaml(path: Path, default=None):
    if not path.exists():
        return default
    with path.open() as fh:
        data = yaml.safe_load(fh)
    return default if data is None else data


def deep_merge(base, extra):
    """Merge extra into base (returns new). Dicts merge, everything else replaces."""
    if isinstance(base, dict) and isinstance(extra, dict):
        out = dict(base)
        for k, v in extra.items():
            out[k] = deep_merge(base[k], v) if k in base else copy.deepcopy(v)
        return out
    return copy.deepcopy(extra)


def line_of(text: str, pos: int) -> int:
    return text.count("\n", 0, pos) + 1


def words_from_camel(name: str) -> str:
    s = re.sub(r"([a-z0-9])([A-Z])", r"\1 \2", name)
    s = re.sub(r"([A-Z]+)([A-Z][a-z])", r"\1 \2", s)
    return s.replace("_", " ").strip()


def norm_hint(s: str) -> str:
    s = s.lower()
    for junk in ("controller", "ctrl", "svc", "service", "interface", "impl", "handler", "get", "()", "_", "-", "."):
        s = s.replace(junk, "")
    return s


# ---------------------------------------------------------------------------
# lexing (Go and JS/TS): strings, comments, balanced groups
# ---------------------------------------------------------------------------

def skip_go_string(s: str, i: int) -> int:
    q = s[i]
    if q == "`":
        j = s.find("`", i + 1)
        return len(s) if j < 0 else j + 1
    j = i + 1
    while j < len(s):
        c = s[j]
        if c == "\\":
            j += 2
            continue
        if c == q or c == "\n":
            return j + 1
        j += 1
    return j


def skip_js_braces(s: str, i: int) -> int:
    depth = 0
    j = i
    while j < len(s):
        c = s[j]
        if c in "\"'`":
            j = skip_js_string(s, j)
            continue
        if c == "{":
            depth += 1
        elif c == "}":
            depth -= 1
            if depth == 0:
                return j + 1
        j += 1
    return j


def skip_js_string(s: str, i: int) -> int:
    q = s[i]
    j = i + 1
    if q != "`":
        while j < len(s):
            c = s[j]
            if c == "\\":
                j += 2
                continue
            if c == q or c == "\n":
                return j + 1
            j += 1
        return j
    while j < len(s):
        c = s[j]
        if c == "\\":
            j += 2
            continue
        if c == "`":
            return j + 1
        if c == "$" and j + 1 < len(s) and s[j + 1] == "{":
            j = skip_js_braces(s, j + 1)
            continue
        j += 1
    return j


def strip_comments(src: str, lang: str) -> str:
    """Blank out comments, keeping offsets and newlines."""
    skip = skip_go_string if lang == "go" else skip_js_string
    out = list(src)
    i, n = 0, len(src)
    while i < n:
        c = src[i]
        if c == "/" and i + 1 < n and src[i + 1] == "/":
            j = src.find("\n", i)
            j = n if j < 0 else j
            for k in range(i, j):
                out[k] = " "
            i = j
        elif c == "/" and i + 1 < n and src[i + 1] == "*":
            j = src.find("*/", i + 2)
            j = n if j < 0 else j + 2
            for k in range(i, j):
                if src[k] != "\n":
                    out[k] = " "
            i = j
        elif c in ("\"", "`") or (c == "'" and lang != "go") or (c == "'" and lang == "go"):
            i = skip(src, i)
        else:
            i += 1
    return "".join(out)


def read_group(s: str, i: int, lang: str = "go"):
    """s[i] is an opening bracket. Returns (end_index_after_close, [top-level comma parts])."""
    skip = skip_go_string if lang == "go" else skip_js_string
    depth = 0
    parts = []
    start = i + 1
    j = i
    while j < len(s):
        c = s[j]
        if c in "\"'`":
            j = skip(s, j)
            continue
        if c in "([{":
            depth += 1
        elif c in ")]}":
            depth -= 1
            if depth == 0:
                parts.append(s[start:j])
                parts = [p.strip() for p in parts]
                if parts and parts[-1] == "":
                    parts.pop()
                return j + 1, parts
        elif c == "," and depth == 1:
            parts.append(s[start:j])
            start = j + 1
        j += 1
    return len(s), [p.strip() for p in parts]


def split_top(s: str, sep: str, lang: str = "go"):
    skip = skip_go_string if lang == "go" else skip_js_string
    parts, depth, start, j = [], 0, 0, 0
    while j < len(s):
        c = s[j]
        if c in "\"'`":
            j = skip(s, j)
            continue
        if c in "([{":
            depth += 1
        elif c in ")]}":
            depth -= 1
        elif depth == 0 and s.startswith(sep, j):
            parts.append(s[start:j])
            j += len(sep)
            start = j
            continue
        j += 1
    parts.append(s[start:])
    return [p.strip() for p in parts]


def go_str_literal(expr: str):
    expr = expr.strip()
    m = re.fullmatch(r'"((?:[^"\\]|\\.)*)"', expr)
    if m:
        return m.group(1)
    m = re.fullmatch(r"`([^`]*)`", expr)
    if m:
        return m.group(1)
    return None


# ---------------------------------------------------------------------------
# Go source index (types, funcs, consts) per service
# ---------------------------------------------------------------------------

class GoFunc:
    def __init__(self, **kw):
        self.__dict__.update(kw)


class GoType:
    def __init__(self, **kw):
        self.__dict__.update(kw)


class GoIndex:
    """Index of one Go module: struct/interface/named types, funcs, string consts and []string vars."""

    def __init__(self, svc_dir: Path):
        self.dir = svc_dir
        self.types: dict[tuple[str, str], GoType] = {}
        self.types_by_name: dict[str, list[GoType]] = defaultdict(list)
        self.funcs: list[GoFunc] = []
        self.funcs_by_name: dict[str, list[GoFunc]] = defaultdict(list)
        self.consts: dict[str, str] = {}
        self.string_slices: dict[str, list[str]] = {}
        self.files: dict[Path, tuple[str, str, str]] = {}  # path -> (pkg, src, stripped)
        for path in sorted(svc_dir.rglob("*.go")):
            rel = path.relative_to(svc_dir).as_posix()
            if rel.endswith("_test.go") or rel.startswith(("vendor/", "docs/", "cmd/docs/")):
                continue
            src = path.read_text(errors="replace")
            st = strip_comments(src, "go")
            m = re.search(r"(?m)^package\s+(\w+)", st)
            pkg = m.group(1) if m else "main"
            self.files[path] = (pkg, src, st)
            self._index_types(path, pkg, st)
            self._index_funcs(path, pkg, src, st)
            self._index_consts(st)

    # -- types ---------------------------------------------------------------
    def _index_types(self, path, pkg, st):
        for m in re.finditer(r"(?m)^type\b", st):
            j = m.end()
            while j < len(st) and st[j] in " \t":
                j += 1
            if j < len(st) and st[j] == "(":
                end, _ = read_group(st, j)
                self._parse_specs(path, pkg, st, j + 1, end - 1, many=True)
            else:
                self._parse_specs(path, pkg, st, j, len(st), many=False)

    def _parse_specs(self, path, pkg, st, start, end, many):
        k = start
        while k < end:
            while k < end and st[k] in " \t\r\n;":
                k += 1
            if k >= end:
                return
            m = re.compile(r"(\w+)\s*(\[[^\]]*\])?\s*(=\s*)?").match(st, k)
            if not m:
                return
            name = m.group(1)
            k = m.end()
            tm = re.compile(r"(struct|interface)\s*\{").match(st, k)
            if tm:
                brace = st.index("{", k)
                close, _ = read_group(st, brace)
                kind = tm.group(1)
                body = st[brace + 1:close - 1]
                t = GoType(pkg=pkg, name=name, kind=kind, body=body, expr=None, file=path,
                           line=line_of(st, k))
                k = close
            else:
                nl = st.find("\n", k)
                nl = end if nl < 0 or nl > end else nl
                expr = st[k:nl].strip()
                t = GoType(pkg=pkg, name=name, kind="named", body=None, expr=expr, file=path,
                           line=line_of(st, k))
                k = nl
            self.types.setdefault((pkg, name), t)
            self.types_by_name[name].append(t)
            if not many:
                return

    # -- funcs ---------------------------------------------------------------
    FUNC_RE = re.compile(r"(?m)^func\s*(?:\(\s*(\w+)?\s*\*?\s*(\w+)(?:\[[^\]]*\])?\s*\)\s*)?(\w+)\s*(?:\[[^\]]*\])?\s*\(")

    def _index_funcs(self, path, pkg, src, st):
        for m in self.FUNC_RE.finditer(st):
            recv_var, recv_type, name = m.group(1), m.group(2), m.group(3)
            pend, params = read_group(st, m.end() - 1)
            brace = st.find("{", pend)
            if brace < 0:
                continue
            results = st[pend:brace].strip()
            bend, _ = read_group(st, brace)
            body = st[brace + 1:bend - 1]
            # preceding comment block (from the original source) for swag annotations
            doc = []
            lines = src[:m.start()].split("\n")
            for ln in reversed(lines[:-1] if lines and lines[-1] == "" else lines):
                if ln.strip().startswith("//"):
                    doc.append(ln.strip()[2:].strip())
                else:
                    break
            f = GoFunc(pkg=pkg, recv_var=recv_var, recv_type=recv_type, name=name, params=params,
                       results=results, body=body, body_offset=brace + 1, file=path,
                       line=line_of(st, m.start()), doc=list(reversed(doc)))
            self.funcs.append(f)
            self.funcs_by_name[name].append(f)

    def _index_consts(self, st):
        for m in re.finditer(r'(?m)^\s*(\w+)\s*(?:string\s*)?=\s*"([^"\n]*)"\s*$', st):
            self.consts.setdefault(m.group(1), m.group(2))
        for m in re.finditer(r"(?m)^\s*(?:var\s+)?(\w+)\s*=\s*\[\]string\s*\{", st):
            end, parts = read_group(st, st.index("{", m.start()))
            vals = [go_str_literal(p) for p in parts]
            if all(v is not None for v in vals):
                self.string_slices[m.group(1)] = vals

    # -- lookups -------------------------------------------------------------
    def find_type(self, name: str, pkg_hint: str | None = None):
        if pkg_hint and (pkg_hint, name) in self.types:
            return self.types[(pkg_hint, name)]
        cands = self.types_by_name.get(name, [])
        return cands[0] if cands else None

    def struct_fields(self, t: GoType):
        """Yields (names[], type_expr, tag_str) for a struct body."""
        if not t or t.kind != "struct":
            return []
        return parse_struct_fields(t.body)

    def interface_methods(self, t: GoType):
        out = {}
        if not t or t.kind != "interface":
            return out
        for m in re.finditer(r"(?m)^\s*(\w+)\s*\(", t.body):
            pend, _ = read_group(t.body, m.end() - 1)
            nl = t.body.find("\n", pend)
            out[m.group(1)] = t.body[pend:nl if nl >= 0 else len(t.body)].strip()
        return out


def parse_struct_fields(body: str):
    decls, depth, start, j = [], 0, 0, 0
    while j < len(body):
        c = body[j]
        if c in "\"`'":
            j = skip_go_string(body, j)
            continue
        if c in "([{":
            depth += 1
        elif c in ")]}":
            depth -= 1
        elif depth == 0 and c in "\n;":
            decls.append(body[start:j])
            start = j + 1
        j += 1
    decls.append(body[start:])
    out = []
    for d in decls:
        d = d.strip()
        if not d:
            continue
        tag = ""
        m = re.search(r"`([^`]*)`\s*$", d)
        if m:
            tag, d = m.group(1), d[:m.start()].strip()
        else:
            m = re.search(r'\s"((?:[^"\\]|\\.)*)"\s*$', d)
            if m:
                tag, d = m.group(1).replace('\\"', '"'), d[:m.start()].strip()
        if re.fullmatch(r"\*?[\w.]+(\[[^\]]*\])?", d):
            out.append(([], d, tag))
            continue
        m = re.match(r"^([A-Za-z_]\w*(?:\s*,\s*[A-Za-z_]\w*)*)\s+(.+)$", d, re.S)
        if m:
            out.append(([n.strip() for n in m.group(1).split(",")], m.group(2).strip(), tag))
    return out


def tag_value(tag: str, key: str):
    m = re.search(rf'(?:^|\s){key}:"([^"]*)"', tag)
    return m.group(1) if m else None


def split_results(results: str):
    r = results.strip()
    if not r:
        return []
    if r.startswith("("):
        _, parts = read_group(r, 0)
    else:
        parts = [r]
    out = []
    for p in parts:
        p = p.strip()
        m = re.match(r"^(\w+)\s+(\S.*)$", p)
        if m and not p.startswith(("map[", "chan ", "func(", "struct", "interface")) and "." not in m.group(1):
            out.append(m.group(2).strip())
        else:
            out.append(p)
    return out


# ---------------------------------------------------------------------------
# Go type expression -> JSON schema
# ---------------------------------------------------------------------------

INT_TYPES = {"int", "int8", "int16", "int32", "int64", "uint", "uint8", "uint16", "uint32", "uint64", "uintptr", "byte", "rune"}
QUALIFIED_SCHEMAS = {
    "time.Time": {"type": "string", "format": "date-time"},
    "time.Duration": {"type": "integer", "description": "nanoseconds"},
    "uuid.UUID": {"type": "string", "format": "uuid"},
    "uuid.NullUUID": {"type": ["string", "null"], "format": "uuid"},
    "gorm.DeletedAt": {"type": ["string", "null"], "format": "date-time"},
    "datatypes.JSON": {"description": "free-form JSON"},
    "json.RawMessage": {"description": "free-form JSON"},
    "datatypes.JSONMap": {"type": "object", "additionalProperties": True},
    "datatypes.Date": {"type": "string", "format": "date-time"},
    "pq.StringArray": {"type": "array", "items": {"type": "string"}},
    "pq.Int64Array": {"type": "array", "items": {"type": "integer"}},
    "pq.Float64Array": {"type": "array", "items": {"type": "number"}},
    "decimal.Decimal": {"type": "string", "description": "decimal number serialised as a string"},
    "sql.NullString": {"type": "object", "properties": {"String": {"type": "string"}, "Valid": {"type": "boolean"}}},
    "sql.NullTime": {"type": "object", "properties": {"Time": {"type": "string", "format": "date-time"}, "Valid": {"type": "boolean"}}},
    "sql.NullInt64": {"type": "object", "properties": {"Int64": {"type": "integer"}, "Valid": {"type": "boolean"}}},
    "multipart.FileHeader": {"type": "string", "format": "binary"},
    "gin.H": {"type": "object", "additionalProperties": True},
    "jwt.MapClaims": {"type": "object", "additionalProperties": True},
}
ENTITY_READONLY = {"created_at", "updated_at", "deleted_at", "createdAt", "updatedAt", "deletedAt", "CreatedAt", "UpdatedAt", "DeletedAt"}


class SchemaBuilder:
    """Turns Go types of one service into OpenAPI component schemas."""

    def __init__(self, svc_short: str, index: GoIndex, components: dict):
        self.svc = svc_short
        self.idx = index
        self.components = components  # shared dict name -> schema
        self.names: dict[tuple[str, str], str] = {}
        self.building: set = set()

    def comp_name(self, t: GoType) -> str:
        key = (t.pkg, t.name)
        if key in self.names:
            return self.names[key]
        name = f"{self.svc}.{t.name}"
        taken = set(self.names.values())
        if name in taken:
            name = f"{self.svc}.{t.pkg}.{t.name}"
        self.names[key] = name
        return name

    def schema(self, texpr: str, pkg: str, depth: int = 0) -> dict:
        t = (texpr or "").strip()
        if not t:
            return {}
        if depth > 12:
            return {}
        if t.startswith("*"):
            inner = self.schema(t[1:], pkg, depth + 1)
            return nullable(inner)
        if t.startswith("[]"):
            if t == "[]byte":
                return {"type": "string", "format": "byte"}
            return {"type": "array", "items": self.schema(t[2:], pkg, depth + 1)}
        m = re.match(r"^\[\d+\](.+)$", t)
        if m:
            return {"type": "array", "items": self.schema(m.group(1), pkg, depth + 1)}
        if t.startswith("map["):
            end, parts = read_group(t, 3)
            return {"type": "object", "additionalProperties": self.schema(t[end:], pkg, depth + 1) or True}
        if t in ("interface{}", "any") or t.startswith("interface"):
            return {}
        if t.startswith("struct"):
            brace = t.find("{")
            close, _ = read_group(t, brace)
            return self.struct_schema(t[brace + 1:close - 1], pkg, depth + 1, request=False)
        if t.startswith("func(") or t.startswith("chan "):
            return {}
        if t == "string":
            return {"type": "string"}
        if t == "bool":
            return {"type": "boolean"}
        if t in INT_TYPES:
            return {"type": "integer"} if t not in ("int64", "uint64") else {"type": "integer", "format": "int64"}
        if t in ("float32", "float64"):
            return {"type": "number"}
        if t == "error":
            return {"type": "string"}
        gm = re.match(r"^([\w.]+)\[(.+)\]$", t)
        if gm:  # generics, e.g. datatypes.JSONType[[]Foo], datatypes.JSONSlice[Foo]
            base, arg = gm.group(1), gm.group(2)
            inner = self.schema(arg, pkg, depth + 1)
            if base.endswith("JSONSlice"):
                return {"type": "array", "items": inner}
            return inner
        if t in QUALIFIED_SCHEMAS:
            return copy.deepcopy(QUALIFIED_SCHEMAS[t])
        if "." in t:
            qual, name = t.rsplit(".", 1)
            gt = self.idx.find_type(name, qual) or self.idx.find_type(name)
            if gt is None or (gt.pkg != qual and (qual, name) not in self.idx.types and qual not in ("models", "dto", "domain", "entity", "services", "service", "types", "response", "request")):
                if gt is None:
                    return {"description": f"Go type {t}"}
        else:
            gt = self.idx.find_type(t, pkg)
            if gt is None:
                return {"description": f"Go type {t}"}
        return self.type_ref(gt, depth)

    def type_ref(self, gt: GoType, depth: int = 0) -> dict:
        if gt.kind == "named":
            expr = gt.expr or ""
            if expr.startswith("="):
                expr = expr[1:].strip()
            if gt.name in ("Status",) and expr == "string":
                return {"type": "string"}
            return self.schema(expr, gt.pkg, depth + 1)
        if gt.kind == "interface":
            return {}
        name = self.comp_name(gt)
        if name not in self.components and (gt.pkg, gt.name) not in self.building:
            self.building.add((gt.pkg, gt.name))
            sch = self.struct_schema(gt.body, gt.pkg, depth + 1, request=False)
            sch.setdefault("description", f"Go: {gt.pkg}.{gt.name} ({gt.file.relative_to(ROOT).as_posix()})")
            self.components[name] = sch
            self.building.discard((gt.pkg, gt.name))
        return {"$ref": f"#/components/schemas/{name}"}

    def struct_schema(self, body: str, pkg: str, depth: int, request: bool) -> dict:
        props: dict = OrderedDict()
        required = []
        for names, texpr, tag in parse_struct_fields(body):
            jt = tag_value(tag, "json")
            jname = None
            omitempty = False
            if jt is not None:
                bits = jt.split(",")
                if bits[0] == "-" and len(bits) == 1:
                    continue
                jname = bits[0] or None
                omitempty = "omitempty" in bits[1:]
            if not names:  # embedded
                base = texpr.lstrip("*")
                if jname:
                    props[jname] = self.schema(texpr, pkg, depth + 1)
                    continue
                if base == "gorm.Model":
                    props.update({"ID": {"type": "integer", "readOnly": True}, "CreatedAt": {"type": "string", "format": "date-time", "readOnly": True},
                                  "UpdatedAt": {"type": "string", "format": "date-time", "readOnly": True},
                                  "DeletedAt": {"type": ["string", "null"], "format": "date-time", "readOnly": True}})
                    continue
                qual, _, nm = base.rpartition(".")
                gt = self.idx.find_type(nm, qual or pkg)
                if gt and gt.kind == "struct":
                    inner = self.struct_schema(gt.body, gt.pkg, depth + 1, request)
                    props.update(inner.get("properties", {}))
                    required += inner.get("required", [])
                continue
            for n in names:
                if not n[:1].isupper():
                    continue
                key = jname or n
                s = self.schema(texpr, pkg, depth + 1)
                binding = (tag_value(tag, "binding") or "") + "," + (tag_value(tag, "validate") or "")
                if re.search(r"(^|,)required(,|$)", binding):
                    required.append(key)
                ex = tag_value(tag, "example")
                if ex is not None and "$ref" not in s:
                    s = dict(s, example=ex)
                enum = re.search(r"oneof=([^,]+)", binding)
                if enum and "$ref" not in s:
                    s = dict(s, enum=enum.group(1).split(" "))
                props[key] = s
        has_ts = any(k in props for k in ("created_at", "createdAt", "CreatedAt"))
        for k in list(props):
            if (k in ENTITY_READONLY or (has_ts and k in ("id", "ID"))) and "$ref" not in props[k]:
                props[k] = dict(props[k], readOnly=True)
        out = {"type": "object", "properties": dict(props)}
        if required:
            out["required"] = sorted(set(required), key=required.index)
        return out


def nullable(s: dict) -> dict:
    if not s:
        return s
    if "$ref" in s:
        return s
    t = s.get("type")
    if isinstance(t, str):
        return dict(s, type=[t, "null"])
    return s


# ---------------------------------------------------------------------------
# Kong
# ---------------------------------------------------------------------------

class Kong:
    def __init__(self, path: Path):
        self.path = path
        data = load_yaml(path, {})
        self.routes = []
        for svc in data.get("services", []) or []:
            host = svc.get("host") or ""
            spath = svc.get("path") or ""
            if svc.get("url"):
                m = re.match(r"^\w+://([^/:]+)(?::\d+)?(/.*)?$", svc["url"])
                if m:
                    host, spath = m.group(1), m.group(2) or ""
            for r in svc.get("routes", []) or []:
                for p in r.get("paths", []) or []:
                    self.routes.append({"prefix": p, "kong_service": svc["name"], "host": host,
                                        "strip": r.get("strip_path", True), "svc_path": spath,
                                        "route": r.get("name")})

    @staticmethod
    def target_service(host: str) -> str:
        for name in ("auth-service", "audit-service", "master-service", "risk-service", "analytics-service", "python-ai"):
            if name in host:
                return name
        return host

    def match(self, public_path: str):
        cands = [r for r in self.routes if public_path.startswith(r["prefix"])]
        return max(cands, key=lambda r: len(r["prefix"])) if cands else None

    def upstream_path(self, r, public_path: str) -> str:
        rest = public_path[len(r["prefix"]):] if r["strip"] else public_path
        base = r["svc_path"].rstrip("/")
        if r["strip"]:
            return (base + "/" + rest.lstrip("/")) if rest else (base or "/")
        return base + rest if base else rest

    def public_paths(self, service: str, internal: str):
        """Public paths (canonical first, then aliases) under which `internal` of `service` is reachable."""
        sample = re.sub(r"\{[^}]+\}", "x", internal)
        out = []
        for r in self.routes:
            if self.target_service(r["host"]) != service:
                continue
            if not r["strip"]:
                if sample.startswith(r["prefix"]):
                    best = self.match(sample)
                    if best is r or (best and best["prefix"] == r["prefix"] and self.target_service(best["host"]) == service):
                        out.append((internal, r))
            else:
                base = r["svc_path"].rstrip("/")
                if base and not (sample == base or sample.startswith(base + "/")):
                    continue
                pub = r["prefix"].rstrip("/") + internal[len(base):]
                pub_sample = re.sub(r"\{[^}]+\}", "x", pub)
                best = self.match(pub_sample)
                if best is r and self.upstream_path(r, pub_sample) == sample:
                    out.append((pub, r))
        out.sort(key=lambda pr: (not (pr[1]["strip"] is False or not pr[1]["svc_path"]), len(pr[0])))
        # canonical: a non-stripping route, else the stripping route without a service path (python-ai)
        seen, uniq = set(), []
        for p, r in out:
            if p not in seen:
                seen.add(p)
                uniq.append((p, r))
        return uniq


# ---------------------------------------------------------------------------
# Gin route extraction
# ---------------------------------------------------------------------------

ROUTE_CALL_RE = re.compile(r"([A-Za-z_]\w*(?:\.[A-Za-z_]\w*)*)\s*\.\s*(Group|Use|GET|POST|PUT|PATCH|DELETE|HEAD|OPTIONS|Any|Handle)\s*\(")


def gin_path(p: str) -> str:
    p = re.sub(r":(\w+)", r"{\1}", p)
    p = re.sub(r"\*(\w+)", r"{\1}", p)
    return p


def join_path(a: str, b: str) -> str:
    if not b:
        return a or ""
    if not a:
        return b
    return a.rstrip("/") + "/" + b.lstrip("/") if b != "/" else a.rstrip("/") + "/"


def extract_routes(index: GoIndex, svc_name: str, file_cfg: dict):
    path = index.dir / file_cfg["file"]
    pkg, src, st = index.files[path]
    groups = {}
    for var, root in (file_cfg.get("roots") or {}).items():
        if isinstance(root, dict):
            groups[var] = (root.get("prefix", ""), list(root.get("middleware", [])))
        else:
            groups[var] = (root or "", [])
    # simple assignments (middleware values, controller constructors), by position
    assigns = []
    for m in re.finditer(r"(?m)\b(\w+)\s*:=\s*", st):
        k = m.end()
        nl = st.find("\n", k)
        seg = st[k:nl if nl >= 0 else len(st)]
        p = seg.find("(")
        if p >= 0 and "{" not in seg[:p]:
            end, _ = read_group(st, k + p)
            expr = st[k:end]
        else:
            expr = seg
        assigns.append((m.start(), m.group(1), expr.strip()))

    def assign_before(name, pos):
        val = None
        for apos, n, e in assigns:
            if apos >= pos:
                break
            if n == name:
                val = e
        return val

    routes = []
    for m in ROUTE_CALL_RE.finditer(st):
        recv, verb = m.group(1), m.group(2)
        if recv not in groups:
            continue
        end, args = read_group(st, m.end() - 1)
        prefix, mws = groups[recv]
        if verb == "Group":
            line_start = st.rfind("\n", 0, m.start()) + 1
            lm = re.search(r"(\w+)\s*:?=\s*$", st[line_start:m.start()])
            sub = go_str_literal(args[0]) if args else ""
            if lm and sub is not None:
                groups[lm.group(1)] = (join_path(prefix, sub), mws + args[1:])
            continue
        if verb == "Use":
            groups[recv] = (prefix, mws + args)
            continue
        if verb in ("Any", "Handle"):
            continue
        sub = go_str_literal(args[0]) if args else None
        if sub is None:
            continue
        full = join_path(prefix, sub)
        handlers = args[1:]
        mw_exprs = []
        for e in mws + handlers[:-1]:
            if re.fullmatch(r"\w+", e):
                resolved = assign_before(e, m.start())
                mw_exprs.append(resolved or e)
            else:
                mw_exprs.append(e)
        routes.append({
            "service": svc_name, "method": verb, "internal_path": gin_path(full) or "/",
            "gin_path": full, "handler_expr": handlers[-1] if handlers else "",
            "middleware": mw_exprs, "file": path, "line": line_of(st, m.start()),
            "pkg": pkg, "assign_before": assign_before, "pos": m.start(),
        })
    return routes


def describe_middleware(mws, index: GoIndex):
    auth = False
    notes = []
    roles = []
    for e in mws:
        if "Authenticate" in e:
            auth = True
        rm = re.search(r"Require(Roles|SelfOrRoles|RolesForStatusChange)\s*\(", e)
        if rm:
            end, args = read_group(e, e.index("(", rm.start()))
            vals = []
            for a in args:
                lit = go_str_literal(a)
                if lit is not None:
                    vals.append(lit)
                elif a.endswith("..."):
                    vals += index.string_slices.get(a[:-3].split(".")[-1], [a])
            kind = rm.group(1)
            if kind == "Roles":
                roles.append(("roles", vals))
            elif kind == "SelfOrRoles":
                roles.append(("self_or_roles", vals[1:] if vals else []))
                notes.append(f"Allowed for the user named by path parameter `{vals[0] if vals else 'id'}` (self) or roles {', '.join(vals[1:])}")
            else:
                # (db, table, initialStatus, roles...)
                roles.append(("status_change_roles", vals[2:] if len(vals) > 2 else vals))
                notes.append(f"Changing `status` away from `{vals[1] if len(vals) > 1 else 'Draft'}` requires one of roles {', '.join(vals[2:])}; other writes are open to any authenticated user")
        elif "Authenticate" not in e and not re.search(r"Require", e):
            short = re.sub(r"\(.*", "()", e.strip())
            notes.append(f"Extra middleware: `{short}`")
    return auth, roles, notes


# ---------------------------------------------------------------------------
# Handler analysis
# ---------------------------------------------------------------------------

CRUD_KINDS = {
    "List": "list", "ListWithItems": "list", "GetByID": "get", "Create": "create", "CreateWithHook": "create",
    "CreateWithHooks": "create", "Update": "update", "UpdateWithHook": "update", "UpdateWithHooks": "update",
    "Delete": "delete",
}


class HandlerInfo:
    def __init__(self):
        self.kind = "unknown"
        self.name = None
        self.loc = None
        self.query = OrderedDict()  # name -> {default, required, schema, description}
        self.body_schema = None
        self.body_kind = None  # json | multipart
        self.form_fields = OrderedDict()
        self.files = OrderedDict()
        self.responses = OrderedDict()  # status -> {"kind": json|binary|redirect|empty, "schema": ...}
        self.error_statuses = set()
        self.perms = []
        self.swag = {}
        self.model = None
        self.uuid_params = set()
        self.preloads = []
        self.notes = []


class Analyzer:
    def __init__(self, index: GoIndex, sb: SchemaBuilder):
        self.idx = index
        self.sb = sb

    # -- resolution of the handler expression --------------------------------
    def analyze(self, route) -> HandlerInfo:
        info = HandlerInfo()
        expr = route["handler_expr"].strip()
        crud_calls = list(re.finditer(r"\bcrud\.(\w+)\s*\(", expr))
        if crud_calls:
            cm = crud_calls[-1]
            if cm.group(1) in CRUD_KINDS:
                return self.crud(info, cm, expr, route)
        if expr.startswith("func("):
            sig_end, params = read_group(expr, 4)
            brace = expr.find("{", sig_end)
            body = expr[brace + 1:expr.rfind("}")]
            ctx = self.ctx_var(params) or "c"
            info.kind = "inline"
            info.loc = f"{route['file'].relative_to(ROOT).as_posix()}:{route['line']}"
            self.analyze_body(info, body, ctx, route["pkg"], None, None)
            return info
        f = self.resolve_func(expr, route)
        if f is None:
            info.notes.append(f"handler `{expr[:80]}` not resolved by the generator")
            return info
        info.kind = "func"
        info.name = f"{f.recv_type + '.' if f.recv_type else ''}{f.name}"
        info.loc = f"{f.file.relative_to(ROOT).as_posix()}:{f.line}"
        info.swag = parse_swag(f.doc)
        body = f.body
        ctx = self.ctx_var(f.params)
        if ctx is None:
            # handler factory: the gin.HandlerFunc is returned from the body
            im = re.search(r"func\s*\(\s*(\w+)\s+\*gin\.Context\s*\)", body)
            ctx = im.group(1) if im else "c"
        self.analyze_body(info, body, ctx, f.pkg, f, None)
        return info

    @staticmethod
    def ctx_var(params):
        for p in params if isinstance(params, list) else [params]:
            m = re.match(r"^(\w+)\s+\*gin\.Context$", p.strip())
            if m:
                return m.group(1)
        return None

    def resolve_func(self, expr: str, route):
        expr = expr.strip()
        call = None
        m = re.match(r"^([\w.()]+?)\s*\((.*)\)$", expr, re.S)
        if m and not expr.endswith("()") or (m and expr.count("(") > 1 and not re.fullmatch(r"[\w.]+\(\)\.\w+", expr)):
            call = m
        chain = expr
        if call and not re.fullmatch(r"[\w.]+(\(\))?(\.\w+)+", expr):
            chain = call.group(1)
        parts = re.split(r"\.", chain.replace("()", ""))
        name = parts[-1]
        allc = self.idx.funcs_by_name.get(name, [])
        cands = [f for f in allc if any("*gin.Context" in p for p in f.params) or "gin.HandlerFunc" in f.results]
        if not cands:
            cands = allc
        if not cands:
            return None
        if len(cands) == 1:
            return cands[0]
        # `ctl.List` inside `func (ctl *Controller) Register(g)`: same receiver type
        reg = next((f for f in self.idx.funcs if f.file == route["file"] and f.body_offset <= route["pos"] <= f.body_offset + len(f.body)), None)
        if reg is not None and reg.recv_type and parts[0] == reg.recv_var and len(parts) == 2:
            same = [f for f in cands if f.recv_type == reg.recv_type and f.pkg == reg.pkg]
            if len(same) == 1:
                return same[0]
        hints = []
        if len(parts) >= 2:
            recv_var = parts[0]
            assigned = route["assign_before"](recv_var, route["pos"])
            if assigned:
                am = re.search(r"New(\w+)\s*\(", assigned)
                if am:
                    hints.append(am.group(1))
            hints += [p for p in parts[1:-1] if p not in ("h", "r", "registry", "controller", "ctrl")]
            if parts[0] not in ("h", "r", "c", "ctx"):
                hints.append(parts[0])
        hints = [norm_hint(h) for h in hints if norm_hint(h)]
        same_file = [f for f in cands if f.file == route["file"]]
        if len(same_file) == 1:
            return same_file[0]

        def score(f):
            rt = norm_hint(f.recv_type or "")
            fp = norm_hint(f.file.stem)
            s = 0
            for h in hints:
                if rt and (h == rt):
                    s += 10
                elif rt and (h in rt or rt in h):
                    s += 5
                if h and (h in fp or fp in h):
                    s += 2
            if f.recv_type is None and len(parts) >= 2:
                s += 1 if norm_hint(parts[0]) in norm_hint(f.pkg) else 0
            return s

        ranked = sorted(cands, key=score, reverse=True)
        if score(ranked[0]) > 0 and (len(ranked) == 1 or score(ranked[0]) > score(ranked[1])):
            return ranked[0]
        return None

    # -- crud ------------------------------------------------------------------
    def crud(self, info, cm, expr, route):
        kind = CRUD_KINDS[cm.group(1)]
        end, args = read_group(expr, expr.index("(", cm.start()))
        model_name = go_str_literal(args[1]) if len(args) > 1 else None
        tm = re.search(r"&\s*(\[\])?([\w.]+)\s*\{\s*\}", args[2] if len(args) > 2 else "")
        info.kind = f"crud.{kind}"
        info.name = f"crud.{cm.group(1)}({model_name})"
        crud_file = self.idx.dir / "controllers" / "crud" / "crud_handler.go"
        info.loc = f"{route['file'].relative_to(ROOT).as_posix()}:{route['line']} (generic {crud_file.relative_to(ROOT).as_posix()})"
        model_ref = self.sb.schema(tm.group(2), route["pkg"]) if tm else {}
        info.model = model_name
        info.preloads = [go_str_literal(a) for a in args[3:] if go_str_literal(a)]
        if kind in ("get", "update", "delete"):
            info.uuid_params.add("id")
        if kind == "list":
            info.query["page"] = {"schema": {"type": "integer", "default": 1, "minimum": 1}, "description": "Page number (1-based)"}
            info.query["page_size"] = {"schema": {"type": "integer", "default": 20, "minimum": 1, "maximum": 100}, "description": "Page size; values outside 1..100 fall back to 20"}
            info.query["search"] = {"schema": {"type": "string"}, "description": "Case-insensitive match on the model's `title`/`name` column when it has one"}
            info.query["order"] = {"schema": {"type": "string", "default": "created_at DESC"}, "description": "`<column> [ASC|DESC]`; only real columns are accepted (400 otherwise)"}
            info.notes.append("Any other query parameter whose snake_case name is a column of the model is applied as an equality filter (e.g. `?status=Draft`, `?companyId=...`); unknown parameters are ignored.")
            info.responses[200] = {"kind": "json", "schema": envelope({"type": "object", "properties": {
                "items": {"type": "array", "items": model_ref}, "pagination": {"$ref": "#/components/schemas/common.Pagination"}}})}
            info.error_statuses |= {400, 500}
        elif kind == "get":
            info.responses[200] = {"kind": "json", "schema": envelope(model_ref)}
            info.error_statuses |= {400, 404, 500}
        elif kind == "create":
            info.body_kind, info.body_schema = "json", model_ref
            info.responses[201] = {"kind": "json", "schema": envelope(model_ref)}
            info.error_statuses |= {400, 500}
        elif kind == "update":
            info.body_kind = "json"
            info.body_schema = {"description": "Partial update: a JSON object with only the fields to change (same field names as the model). Unknown fields are ignored.",
                                "allOf": [model_ref]} if model_ref else {"type": "object"}
            info.responses[200] = {"kind": "json", "schema": envelope(model_ref)}
            info.error_statuses |= {400, 404, 500}
        elif kind == "delete":
            info.notes.append("Soft delete (sets deleted_at).")
            info.responses[200] = {"kind": "json", "schema": {"$ref": "#/components/schemas/common.Message"}}
            info.error_statuses |= {400, 404, 500}
        return info

    # -- body analysis -----------------------------------------------------------
    def analyze_body(self, info, body, ctx, pkg, func, recv_type, depth=0):
        helpers = self.helper_bodies(body, ctx, func) if depth == 0 else []
        texts = [(body, func)] + helpers
        for text, fn in texts:
            for m in re.finditer(rf"\b{ctx}\.(Query|DefaultQuery|GetQuery|QueryArray|GetQueryArray)\(\s*\"([^\"]+)\"(?:\s*,\s*\"([^\"]*)\")?", text):
                name = m.group(2)
                q = info.query.setdefault(name, {"schema": {"type": "string"}})
                if m.group(3) is not None:
                    q["schema"] = dict(q["schema"], default=m.group(3))
                if m.group(1) in ("QueryArray", "GetQueryArray"):
                    q["schema"] = {"type": "array", "items": {"type": "string"}}
            for m in re.finditer(rf"\b{ctx}\.(?:PostForm|DefaultPostForm|GetPostForm|PostFormArray)\(\s*\"([^\"]+)\"", text):
                info.form_fields.setdefault(m.group(1), {"type": "string"})
            for m in re.finditer(rf"\b{ctx}\.FormFile\(\s*\"([^\"]+)\"", text):
                info.files.setdefault(m.group(1), {"type": "string", "format": "binary"})
            for m in re.finditer(r"\.File\[\s*\"([^\"]+)\"\s*\]", text):
                info.files.setdefault(m.group(1), {"type": "array", "items": {"type": "string", "format": "binary"}})
            for m in re.finditer(r"\.Value\[\s*\"([^\"]+)\"\s*\]", text):
                info.form_fields.setdefault(m.group(1), {"type": "string"})
            for m in re.finditer(r"uuid\.Parse\(\s*\w+\.Param\(\s*\"(\w+)\"", text):
                info.uuid_params.add(m.group(1))
            for m in re.finditer(r"(\w+)\s*:=\s*\w+\.Param\(\s*\"(\w+)\"\s*\)", text):
                if re.search(rf"uuid\.Parse\(\s*{m.group(1)}\s*\)", text):
                    info.uuid_params.add(m.group(2))
            for m in re.finditer(r"\b(Perm[A-Z]\w*)\b", text):
                val = self.idx.consts.get(m.group(1))
                if val and val not in info.perms:
                    info.perms.append(val)
            # query struct binding
            for m in re.finditer(rf"\b{ctx}\.ShouldBindQuery\(\s*&(\w+)", text):
                t = self.var_type(m.group(1), text, fn, pkg)
                if t:
                    gt = self.lookup_type(t, fn.pkg if fn else pkg)
                    for names, texpr, tag in self.idx.struct_fields(gt):
                        fname = tag_value(tag, "form") or (names[0] if names else None)
                        if fname and fname != "-":
                            info.query.setdefault(fname.split(",")[0], {"schema": self.sb.schema(texpr, gt.pkg)})
        # request body (main body only, then helpers)
        if info.body_schema is None:
            for text, fn in texts:
                bm = None
                for m in re.finditer(r"\b(?:\w+\.)*(\w*(?i:bind|validate|decode|unmarshal)\w*)\s*\(\s*(?:[\w.]+\s*,\s*)?&(\w+)\s*[,)]", text):
                    fname = m.group(1)
                    if fname in ("ShouldBindQuery", "ShouldBindUri", "BindQuery", "BindUri"):
                        continue
                    if fname == "Unmarshal" and not re.search(r"GetRawData|Request\.Body", text):
                        continue
                    bm = m
                    break
                if bm:
                    var = bm.group(2)
                    sch = self.var_schema(var, text, fn, pkg)
                    if bm.group(1) in ("ShouldBind", "Bind") and (info.files or info.form_fields):
                        info.body_kind = "multipart"
                    else:
                        info.body_kind = "json"
                    info.body_schema = sch if sch is not None else {"type": "object"}
                    break
        if info.body_kind is None and (info.files or info.form_fields):
            info.body_kind = "multipart"
        # responses
        for text, fn in texts:
            self.scan_responses(info, text, ctx, fn, pkg)
        if func is not None and info.swag.get("success"):
            for code, sch in self.swag_success(info.swag["success"], func.pkg).items():
                cur = info.responses.get(code)
                if not cur or not cur.get("schema") or is_generic_envelope(cur.get("schema")):
                    info.responses[code] = {"kind": "json", "schema": sch}

    def helper_bodies(self, body, ctx, func):
        """Bodies of same-receiver methods / package funcs called with the gin context."""
        out = []
        seen = set()
        for m in re.finditer(rf"\b(?:(\w+)\.)?(\w+)\(\s*{ctx}\s*[,)]", body):
            recv, name = m.group(1), m.group(2)
            if name in seen or name in ("JSON", "Param", "Query", "String", "Data", "File", "Status", "Header"):
                continue
            cands = self.idx.funcs_by_name.get(name, [])
            pick = None
            if func is not None and recv == func.recv_var:
                pick = next((f for f in cands if f.recv_type == func.recv_type), None)
            elif recv is None:
                pick = next((f for f in cands if f.recv_type is None and (func is None or f.pkg == func.pkg)), None)
            if pick and pick is not func:
                seen.add(name)
                out.append((pick.body, pick))
        return out[:20]

    def scan_responses(self, info, text, ctx, fn, pkg):
        for m in re.finditer(r"\bresponse\.(OK|Created|Success)\s*\(", text):
            end, args = read_group(text, m.end() - 1)
            if m.group(1) == "Success":
                code = self.status_code(args[1]) if len(args) > 1 else 200
                data = args[3] if len(args) > 3 else None
            else:
                code = 200 if m.group(1) == "OK" else 201
                data = args[2] if len(args) > 2 else None
            if code is None:
                continue
            if data is None or data.strip() == "nil":
                self.add_response(info, code, {"kind": "json", "schema": {"$ref": "#/components/schemas/common.Message"}})
                continue
            ds = self.expr_schema(data, text, fn, pkg)
            self.add_response(info, code, {"kind": "json", "schema": envelope(ds)})
        for m in re.finditer(r"\bresponse\.(BadRequest|Unauthorized|Forbidden|NotFound|InternalServerError|Conflict)\s*\(", text):
            info.error_statuses.add(RESPONSE_HELPER_STATUS[m.group(1)])
        for m in re.finditer(r"\bresponse\.(?:Error|ErrorWithFields)\s*\(\s*\w+\s*,\s*([\w.]+)", text):
            code = self.status_code(m.group(1))
            if code:
                info.error_statuses.add(code)
        for m in re.finditer(rf"\b{ctx}\.(JSON|IndentedJSON|PureJSON|AbortWithStatusJSON)\s*\(", text):
            end, args = read_group(text, m.end() - 1)
            if len(args) < 2:
                continue
            code = self.status_code(args[0])
            if code is None:
                continue
            if code >= 400:
                info.error_statuses.add(code)
                continue
            self.add_response(info, code, {"kind": "json", "schema": self.expr_schema(args[1], text, fn, pkg)})
        for m in re.finditer(rf"\b{ctx}\.(File|FileAttachment|FileFromFS|Data|DataFromReader|Stream)\s*\(", text):
            ct = None
            if m.group(1) in ("Data", "DataFromReader"):
                end, args = read_group(text, m.end() - 1)
                if len(args) > 1:
                    ct = go_str_literal(args[1]) if m.group(1) == "Data" else (go_str_literal(args[2]) if len(args) > 2 else None)
            self.add_response(info, 200, {"kind": "binary", "content_type": ct})
        if re.search(rf"\b{ctx}\.Redirect\s*\(", text):
            self.add_response(info, 302, {"kind": "redirect"})
        if re.search(r"http\.ServeContent|http\.ServeFile|\.Writer\.Write\(|io\.Copy\(\s*\w+\.Writer", text):
            self.add_response(info, 200, {"kind": "binary"})
        for m in re.finditer(rf"\b{ctx}\.(?:Status|AbortWithStatus)\s*\(\s*([\w.]+)", text):
            code = self.status_code(m.group(1))
            if code and code >= 400:
                info.error_statuses.add(code)
            elif code == 204:
                self.add_response(info, 204, {"kind": "empty"})

    @staticmethod
    def add_response(info, code, resp):
        cur = info.responses.get(code)
        if cur is None:
            info.responses[code] = resp
            return
        if cur["kind"] == "json" and resp["kind"] == "json":
            # keep the most informative shape
            if score_schema(resp.get("schema")) > score_schema(cur.get("schema")):
                info.responses[code] = resp
        elif cur["kind"] != "binary" and resp["kind"] == "binary":
            info.responses[code] = dict(resp, alt=cur)
        elif cur["kind"] == "binary" and resp["kind"] == "json":
            cur["alt"] = resp

    @staticmethod
    def status_code(expr: str):
        expr = expr.strip()
        if expr.isdigit():
            return int(expr)
        m = re.match(r"^http\.(\w+)$", expr)
        if m:
            return STATUS_NAMES.get(m.group(1))
        return None

    # -- expression typing ----------------------------------------------------------
    def lookup_type(self, texpr, pkg):
        if "@" in texpr:
            texpr, pkg = texpr.split("@", 1)
        t = texpr.strip().lstrip("*")
        while t.startswith("[]"):
            t = t[2:].lstrip("*")
        if "." in t:
            q, n = t.rsplit(".", 1)
            return self.idx.find_type(n, q) or self.idx.find_type(n)
        return self.idx.find_type(t, pkg)

    def var_type(self, var, text, fn, pkg):
        m = re.search(rf"\bvar\s+{var}\s+([^\n=]+)", text)
        if m:
            return m.group(1).strip()
        m = re.search(rf"\b{var}\s*:?=\s*&?((?:\[\])?\*?[\w.]+)\s*\{{", text)
        if m and m.group(1) not in ("gin.H", "map", "func"):
            return m.group(1)
        m = re.search(rf"\b{var}\s*:?=\s*new\(\s*([\w.]+)\s*\)", text)
        if m:
            return m.group(1)
        m = re.search(rf"\b{var}\s*:?=\s*make\(\s*([^,)]+)", text)
        if m:
            return m.group(1).strip()
        return None

    def var_schema(self, var, text, fn, pkg):
        m = re.search(rf"\bvar\s+{var}\s+struct\s*\{{", text)
        if m:
            brace = text.index("{", m.start())
            end, _ = read_group(text, brace)
            return self.sb.struct_schema(text[brace + 1:end - 1], fn.pkg if fn else pkg, 0, request=True)
        m = re.search(rf"\b{var}\s*:?=\s*&?struct\s*\{{", text)
        if m:
            brace = text.index("{", m.start())
            end, _ = read_group(text, brace)
            return self.sb.struct_schema(text[brace + 1:end - 1], fn.pkg if fn else pkg, 0, request=True)
        t = self.var_type(var, text, fn, pkg)
        if t:
            if t.startswith("map["):
                return {"type": "object", "additionalProperties": True}
            gt = self.lookup_type(t, fn.pkg if fn else pkg)
            if gt is not None and gt.kind == "struct" and not t.startswith("[]"):
                # request structs get `required` from binding tags: build inline-able component
                return self.sb.type_ref(gt)
            return self.sb.schema(t, fn.pkg if fn else pkg)
        return None

    def expr_schema(self, expr, text, fn, pkg, depth=0):
        if expr is None:
            return None
        e = expr.strip()
        if depth > 4:
            return {}
        if e.startswith("&"):
            e = e[1:]
        if e == "nil":
            return None
        if re.match(r"^(gin\.H|map\[string\](interface\{\}|any))\s*\{", e):
            brace = e.index("{", e.index("H") if e.startswith("gin.H") else e.index("]") + 1)
            end, parts = read_group(e, brace)
            props = OrderedDict()
            for p in parts:
                kv = split_top(p, ":")
                if len(kv) < 2:
                    continue
                key = go_str_literal(kv[0])
                if key is None:
                    continue
                props[key] = self.value_schema(":".join(kv[1:]).strip(), text, fn, pkg, depth + 1)
            return {"type": "object", "properties": dict(props)}
        if re.match(r"^\[\]gin\.H\s*\{", e):
            return {"type": "array", "items": {"type": "object"}}
        m = re.match(r"^((?:\[\])?\*?[\w.]+)\s*\{", e)
        if m and m.group(1) not in ("func",):
            return self.sb.schema(m.group(1), fn.pkg if fn else pkg)
        if re.fullmatch(r"\w+", e):
            return self.ident_schema(e, text, fn, pkg, depth)
        cm = re.match(r"^((?:\w+\.)*)(\w+)\s*\(", e)
        if cm and read_group(e, cm.end() - 1)[0] == len(e):
            chain = cm.group(1).rstrip(".")
            res = None
            if chain:
                res = self.method_results(chain, cm.group(2), fn, pkg)
            if res is None:
                cands = [f for f in self.idx.funcs_by_name.get(cm.group(2), []) if f.recv_type is None]
                if cands:
                    res = [r + f"@{cands[0].pkg}" for r in split_results(cands[0].results)]
            if res:
                t, tp = (res[0].split("@", 1) + [pkg])[:2]
                return self.sb.schema(t, tp)
        m = re.fullmatch(r"(\w+)\.(\w+)", e)
        if m:
            base = self.ident_type(m.group(1), text, fn, pkg)
            if base:
                gt = self.lookup_type(base, fn.pkg if fn else pkg)
                for names, texpr, tag in self.idx.struct_fields(gt):
                    if m.group(2) in names:
                        return self.sb.schema(texpr, gt.pkg)
        return self.value_schema(e, text, fn, pkg, depth + 1)

    def value_schema(self, v, text, fn, pkg, depth):
        v = v.strip()
        if go_str_literal(v) is not None:
            return {"type": "string", "example": go_str_literal(v)} if len(go_str_literal(v)) < 60 else {"type": "string"}
        if v in ("true", "false"):
            return {"type": "boolean"}
        if re.fullmatch(r"-?\d+", v):
            return {"type": "integer"}
        if re.fullmatch(r"-?\d+\.\d+", v):
            return {"type": "number"}
        if v.startswith(("len(", "int(", "int64(", "cap(")):
            return {"type": "integer"}
        if v.startswith(("float64(", "math.")):
            return {"type": "number"}
        if v.startswith(("fmt.Sprintf", "strings.", "err.Error()", "string(")) or v.endswith(".String()") or v.endswith(".Error()"):
            return {"type": "string"}
        if v.startswith("time.Now()"):
            return {"type": "string", "format": "date-time"}
        if v.startswith(("gin.H", "map[", "[]gin.H")) or re.match(r"^(?:\[\])?\*?[\w.]+\s*\{", v) or re.fullmatch(r"&?\w+(\.\w+)?", v):
            return self.expr_schema(v, text, fn, pkg, depth) or {}
        return {}

    def ident_type(self, var, text, fn, pkg):
        t = self.var_type(var, text, fn, pkg)
        if t:
            return t
        # var, err := X.Method(...)
        for m in re.finditer(r"(?m)^\s*([\w\s,]+?)\s*:?=\s*([\w.]+(?:\(\))?(?:\.\w+)*?)\.(\w+)\s*\(", text):
            lhs = [x.strip() for x in m.group(1).split(",")]
            if var not in lhs:
                continue
            idx = lhs.index(var)
            results = self.method_results(m.group(2), m.group(3), fn, pkg)
            if results and idx < len(results):
                return results[idx]
        # var, err := Func(...)
        for m in re.finditer(r"(?m)^\s*([\w\s,]+?)\s*:?=\s*(\w+)\s*\(", text):
            lhs = [x.strip() for x in m.group(1).split(",")]
            if var not in lhs:
                continue
            idx = lhs.index(var)
            cands = [f for f in self.idx.funcs_by_name.get(m.group(2), []) if f.recv_type is None]
            if cands:
                results = split_results(cands[0].results)
                if idx < len(results):
                    return qualify(results[idx], cands[0].pkg)
        # function parameter
        if fn is not None:
            for p in fn.params:
                pm = re.match(rf"^{var}\s+(.+)$", p.strip())
                if pm:
                    return pm.group(1)
        return None

    def ident_schema(self, var, text, fn, pkg, depth):
        t = self.ident_type(var, text, fn, pkg)
        if not t:
            return {}
        if t.startswith("gin.H") or t.startswith("map["):
            # find a literal assignment
            m = re.search(rf"\b{var}\s*:?=\s*(gin\.H\s*\{{)", text)
            if m:
                end, _ = read_group(text, text.index("{", m.start()))
                return self.expr_schema(text[m.start(1):end], text, fn, pkg, depth + 1)
            return {"type": "object", "additionalProperties": True}
        p = t.split("|", 1)
        tpkg = fn.pkg if fn else pkg
        if "@" in t:
            t, tpkg = t.split("@", 1)
        return self.sb.schema(t, tpkg)

    def method_results(self, chain, method, fn, pkg):
        parts = chain.replace("()", "").split(".")
        field_type = None
        if fn is not None and fn.recv_type and parts and parts[0] == fn.recv_var and len(parts) >= 2:
            st = self.idx.find_type(fn.recv_type, fn.pkg)
            cur = st
            for fld in parts[1:]:
                ft = None
                for names, texpr, tag in self.idx.struct_fields(cur):
                    if fld in names or (not names and texpr.lstrip("*").split(".")[-1] == fld):
                        ft = texpr
                        break
                if ft is None:
                    cur = None
                    break
                field_type = ft
                cur = self.lookup_type(ft, cur.pkg)
            if cur is not None:
                if cur.kind == "interface":
                    sig = self.idx.interface_methods(cur).get(method)
                    if sig:
                        return [qualify(r, cur.pkg) + f"@{cur.pkg}" for r in split_results(sig)]
                f = next((f for f in self.idx.funcs_by_name.get(method, []) if f.recv_type == cur.name), None)
                if f:
                    return [r + f"@{f.pkg}" for r in split_results(f.results)]
                # interface implemented by another struct: name-similar receivers
                hint = norm_hint(cur.name)
                for f in self.idx.funcs_by_name.get(method, []):
                    if f.recv_type and (norm_hint(f.recv_type) == hint or hint in norm_hint(f.recv_type)):
                        return [r + f"@{f.pkg}" for r in split_results(f.results)]
        cands = [f for f in self.idx.funcs_by_name.get(method, []) if f.recv_type]
        if len(parts) == 1 and not cands:
            cands = self.idx.funcs_by_name.get(method, [])
        if cands:
            sigs = {tuple(split_results(f.results)) for f in cands}
            if len(sigs) == 1:
                return [r + f"@{cands[0].pkg}" for r in split_results(cands[0].results)]
            hint = norm_hint(parts[-1]) if parts else ""
            for f in cands:
                if hint and f.recv_type and hint in norm_hint(f.recv_type):
                    return [r + f"@{f.pkg}" for r in split_results(f.results)]
        return None

    def swag_success(self, items, pkg):
        out = {}
        for code, kind, texpr in items:
            m = re.match(r"^([\w.]+)\{data=([^}]+)\}$", texpr)
            if m:
                inner = m.group(2)
                data = self.sb.schema(inner.replace("[]", "", 1) if inner.startswith("[]") else inner, pkg)
                if inner.startswith("[]"):
                    data = {"type": "array", "items": data}
                out[code] = envelope(data)
            elif texpr.endswith("Response") and texpr.split(".")[-1] == "Response":
                out[code] = envelope(None)
            else:
                s = self.sb.schema(texpr, pkg)
                out[code] = {"type": "array", "items": s} if kind == "array" else s
        return out


def qualify(texpr, pkg):
    return texpr


def score_schema(s):
    if not s:
        return 0
    return len(json.dumps(s))


def envelope(data_schema):
    if data_schema is None or data_schema == {}:
        return {"$ref": "#/components/schemas/common.Envelope"}
    return {"allOf": [{"$ref": "#/components/schemas/common.Envelope"},
                      {"type": "object", "properties": {"data": data_schema}}]}


def is_generic_envelope(s):
    return s == {"$ref": "#/components/schemas/common.Envelope"}


def parse_swag(doc_lines):
    out = {"success": []}
    for ln in doc_lines:
        m = re.match(r"^@(\w+)\s+(.*)$", ln)
        if not m:
            continue
        k, v = m.group(1), m.group(2).strip()
        if k == "Summary":
            out["summary"] = v
        elif k == "Description":
            out["description"] = v
        elif k == "Success":
            sm = re.match(r"^(\d+)\s+\{(\w+)\}\s+(\S+)", v)
            if sm:
                out["success"].append((int(sm.group(1)), sm.group(2), sm.group(3)))
        elif k == "Param":
            out.setdefault("params", []).append(v)
    return out


# ---------------------------------------------------------------------------
# python-ai and Nuxt
# ---------------------------------------------------------------------------

def fastapi_routes(main_py: Path):
    src = main_py.read_text()
    out = []
    for m in re.finditer(r'@app\.(get|post|put|patch|delete)\(\s*"([^"]+)"', src):
        out.append((m.group(1).upper(), m.group(2), line_of(src, m.start())))
    return out


def nitro_routes(nitro_dir: Path):
    out = []
    if not nitro_dir.exists():
        return out
    for f in sorted(nitro_dir.rglob("*.ts")):
        rel = f.relative_to(nitro_dir).with_suffix("").as_posix()
        method = None
        m = re.match(r"^(.*)\.(get|post|put|patch|delete)$", rel)
        if m:
            rel, method = m.group(1), m.group(2).upper()
        rel = re.sub(r"\[(\w+)\]", r"{\1}", rel)
        rel = re.sub(r"/index$", "", rel)
        out.append((method, "/api/" + rel, f))
    return out


# ---------------------------------------------------------------------------
# Frontend scanning
# ---------------------------------------------------------------------------

BASE_HELPERS = {
    "getAuditServiceBaseUrl": "/api/v1", "getMasterServiceBaseUrl": "/api/v1", "getRiskServiceBaseUrl": "/api/v1",
    "getAuthServiceBaseUrl": "/api/v1", "getDataHubBaseUrl": "/api/v1",
    "getAnalyticsServiceBaseUrl": "/api/analytics", "getPythonAiBaseUrl": "/api/python-ai",
}
CONFIG_KEYS = {
    "apiBase": "/api/v1", "analyticsApiBase": "/api/analytics", "pythonAiBaseUrl": "/api/python-ai",
    "authServiceBaseUrl": "/api/v1", "auditServiceBaseUrl": "/api/v1", "riskServiceBaseUrl": "/api/v1",
    "masterServiceBaseUrl": "/api/v1", "dataHubBaseUrl": "/api/v1",
}
FETCH_RE = re.compile(r"(?<![\w$.])(\$fetch(?:\.raw)?|fetch|useFetch|useLazyFetch|apiFetch)\s*(?=[<(])")
UNKNOWN = "{*}"


class FrontendFile:
    def __init__(self, path: Path):
        self.path = path
        self.raw = path.read_text(errors="replace")
        if path.suffix == ".vue":
            chunks = []
            masked = [" " if c != "\n" else "\n" for c in self.raw]
            for m in re.finditer(r"<script\b[^>]*>(.*?)</script>", self.raw, re.S):
                for k in range(m.start(1), m.end(1)):
                    masked[k] = self.raw[k]
            self.script = "".join(masked)
            self.template = self.raw
        else:
            self.script = self.raw
            self.template = ""
        self.code = strip_comments(self.script, "js")

    def rel(self):
        return self.path.relative_to(ROOT / "frontend").as_posix()


def js_literal_value(e):
    e = e.strip()
    m = re.fullmatch(r"'((?:[^'\\]|\\.)*)'|\"((?:[^\"\\]|\\.)*)\"", e)
    if m:
        return m.group(1) if m.group(1) is not None else m.group(2)
    return None


class FrontendScanner:
    def __init__(self, cfg):
        self.cfg = cfg
        self.calls = []       # dicts: file, line, method(s), url(s), raw
        self.references = []  # literal API paths not tied to a call
        self.unresolved = []

    def files(self):
        base = ROOT / self.cfg["dir"]
        for d in self.cfg["scan_dirs"]:
            p = base / d
            if not p.exists():
                continue
            for f in sorted(p.rglob("*")):
                if f.suffix not in self.cfg["extensions"] or not f.is_file():
                    continue
                rel = f.relative_to(base).as_posix()
                if any(fnmatch.fnmatch(rel, pat) or fnmatch.fnmatch("/" + rel, pat) for pat in self.cfg.get("exclude", [])):
                    continue
                yield f

    def scan(self):
        for f in self.files():
            ff = FrontendFile(f)
            consumed = set()
            wrappers = []
            for m in FETCH_RE.finditer(ff.code):
                j = m.end()
                while j < len(ff.code) and ff.code[j] in " \t\n":
                    j += 1
                if j < len(ff.code) and ff.code[j] == "<":
                    j = skip_generic(ff.code, j)
                    while j < len(ff.code) and ff.code[j] in " \t\n":
                        j += 1
                if j >= len(ff.code) or ff.code[j] != "(":
                    continue
                end, args = read_group(ff.code, j, "js")
                if not args:
                    continue
                kind = m.group(1)
                consumed.add((m.start(), end))
                wrapper = self.wrapper_of(ff, m.start(), args)
                if wrapper:
                    wrappers.append(wrapper)
                    continue
                urls = self.eval_expr(args[0], ff, m.start())
                methods = self.methods_of(args[1] if len(args) > 1 else "", ff, m.start())
                rec = {"file": ff.rel(), "line": line_of(ff.code, m.start()), "methods": methods,
                       "raw": re.sub(r"\s+", " ", args[0])[:160], "client": kind}
                good = []
                for u in urls:
                    nu = normalize_url(u, kind == "apiFetch")
                    if nu:
                        good.append(nu)
                if good:
                    rec["paths"] = sorted(set(good))
                    self.calls.append(rec)
                elif kind != "fetch" or any(UNKNOWN in u for u in urls):
                    rec["paths"] = []
                    rec["values"] = urls
                    self.unresolved.append(rec)
            # calls through local wrappers such as safeApiFetch('/caatt/benford', opts)
            for w in wrappers:
                for cm in re.finditer(rf"(?<![\w$.]){re.escape(w['name'])}\s*\(", ff.code):
                    if w["def_start"] <= cm.start() < w["body_start"]:
                        continue
                    end, args = read_group(ff.code, cm.end() - 1, "js")
                    if len(args) <= w["url_idx"]:
                        continue
                    consumed.add((cm.start(), end))
                    tails = self.eval_expr(args[w["url_idx"]], ff, cm.start())
                    opts = args[w["opt_idx"]] if w["opt_idx"] is not None and len(args) > w["opt_idx"] else ""
                    methods = self.methods_of(opts, ff, cm.start())
                    paths = sorted({nu for pre in w["prefixes"] for t in tails for nu in [normalize_url(pre + t, False)] if nu})
                    rec = {"file": ff.rel(), "line": line_of(ff.code, cm.start()), "methods": methods,
                           "raw": f"{w['name']}(" + re.sub(r"\s+", " ", ", ".join(args))[:140] + ")", "client": "wrapper:" + w["name"]}
                    if paths and not any(p.endswith(UNKNOWN) and p.count("/") <= 2 for p in paths):
                        rec["paths"] = paths
                        self.calls.append(rec)
                    else:
                        rec["paths"], rec["values"] = [], tails
                        self.unresolved.append(rec)
            # stray literal references (window.open, href, helper args...)
            for lit_pos, lit_end, text in iter_js_literals(ff.code):
                if any(a <= lit_pos < b for a, b in consumed):
                    continue
                if not (re.match(r"^[`'\"]\s*\$\{[^}]*(Base|base|BASE|apiBase|getBaseUrl)[^}]*\}\s*/", text) or re.match(r"^[`'\"]/(api|uploads)/", text)):
                    continue
                vals = self.eval_expr(text, ff, lit_pos)
                for v in vals:
                    nu = normalize_url(v, False)
                    if nu and nu.lstrip("localhost:") not in ("/api/v1", "/api/analytics", "/api/python-ai", "/uploads") and not nu.endswith("/" + UNKNOWN + "/" + UNKNOWN):
                        self.references.append({"file": ff.rel(), "line": line_of(ff.code, lit_pos), "path": nu,
                                                "raw": re.sub(r"\s+", " ", text)[:160]})
            if ff.template:
                for m in re.finditer(r"[\"'`](/(?:api|uploads)/[^\"'`\s]+)[\"'`]", ff.template):
                    if ff.script[m.start()] != " ":  # inside <script>: handled above
                        continue
                    nu = normalize_url(re.sub(r"\$\{[^}]*\}", UNKNOWN, m.group(1)), False)
                    if nu:
                        self.references.append({"file": ff.rel(), "line": line_of(ff.raw, m.start()), "path": nu, "raw": m.group(1)})
        # drop references whose path a call in the same file already covers
        covered = {(c["file"], p) for c in self.calls for p in c["paths"]}
        uniq = {}
        for r in self.references:
            if (r["file"], r["path"]) in covered:
                continue
            uniq.setdefault((r["file"], r["path"]), r)
        self.references = list(uniq.values())

    FUNC_DEF_RES = [
        re.compile(r"(?:const|let)\s+(\w+)\s*=\s*(?:async\s*)?\(([^)]*)\)\s*(?::\s*[^=]+?)?\s*=>\s*\{"),
        re.compile(r"(?:async\s+)?function\s+(\w+)\s*\(([^)]*)\)\s*(?::[^{]+)?\{"),
        re.compile(r"(?m)^[ \t]*(?:async\s+)?(?!if\b|for\b|while\b|switch\b|catch\b|return\b)(\w+)\s*\(([^)]*)\)\s*(?::[^{=;]+)?\{"),
    ]

    def enclosing_function(self, ff, pos):
        best = None
        for rx in self.FUNC_DEF_RES:
            for m in rx.finditer(ff.code, 0, pos):
                brace = m.end() - 1
                end, _ = read_group(ff.code, brace, "js")
                if brace < pos < end and (best is None or brace > best["body_start"]):
                    params = []
                    for p in split_top(m.group(2), ",", "js"):
                        pm = re.match(r"^\s*\.{0,3}([A-Za-z_$][\w$]*)", p)
                        if pm:
                            params.append(pm.group(1))
                    best = {"name": m.group(1), "params": params, "def_start": m.start(), "body_start": brace, "body_end": end}
        return best

    def wrapper_of(self, ff, pos, args):
        """A $fetch whose URL ends in a parameter of the enclosing function makes that function a wrapper."""
        u = args[0].strip()
        if not u.startswith("`"):
            return None
        fn = self.enclosing_function(ff, pos)
        if not fn:
            return None
        inters = [x.strip() for x in re.findall(r"\$\{([^}]*)\}", u)]
        tail = inters[-1] if inters else None
        if not tail or tail not in fn["params"] or not u.rstrip("`").endswith("${" + tail + "}"):
            return None
        prefix_tpl = u[:u.rfind("${" + tail + "}")] + "`"
        # `${base}/things/${id}` is an ordinary call; a wrapper takes the path itself: `${base}${endpoint}`
        if not prefix_tpl[:-1].endswith("}"):
            return None
        prefixes = self.eval_expr(prefix_tpl, ff, pos)
        opt = args[1].strip() if len(args) > 1 else ""
        opt_idx = fn["params"].index(opt) if opt in fn["params"] else None
        return dict(fn, url_idx=fn["params"].index(tail), opt_idx=opt_idx, prefixes=prefixes)

    # -- evaluation of URL expressions -----------------------------------------
    def methods_of(self, opts, ff, pos):
        if not opts:
            return ["GET"]
        m = re.search(r"\bmethod\s*:\s*([^,}\n]+)", opts)
        if not m:
            if re.search(r"^\s*\w+\s*$", opts):
                # options object built elsewhere: look for `method:` in the preceding definition
                d = find_definition(ff, opts.strip(), pos)
                if d:
                    return self.methods_of(d, ff, pos)
            return ["GET"]
        val = m.group(1)
        verbs = re.findall(r"['\"`](GET|POST|PUT|PATCH|DELETE|get|post|put|patch|delete)['\"`]", val)
        if verbs:
            return sorted({v.upper() for v in verbs}, key=lambda v: METHOD_ORDER.get(v, 9))
        ident = val.strip()
        if re.fullmatch(r"[\w.]+", ident):
            d = find_definition(ff, ident, pos)
            if d:
                verbs = re.findall(r"['\"`](GET|POST|PUT|PATCH|DELETE)['\"`]", d, re.I)
                if verbs:
                    return sorted({v.upper() for v in verbs}, key=lambda v: METHOD_ORDER.get(v, 9))
        return ["?"]

    def eval_expr(self, expr, ff, pos, depth=0):
        e = expr.strip()
        if depth > 6 or not e:
            return [UNKNOWN]
        e = re.sub(r"\s+as\s+[\w<>\[\]|. ]+$", "", e).rstrip("!").strip()
        while e.startswith("(") and e.endswith(")"):
            end, _ = read_group(e, 0, "js")
            if end != len(e):
                break
            e = e[1:-1].strip()
        if e.startswith("`"):
            end = skip_js_string(e, 0)
            if end == len(e):
                return self.eval_template(e, ff, pos, depth)
        lit = js_literal_value(e)
        if lit is not None:
            return [lit]
        parts = split_top(e, "+", "js")
        if len(parts) > 1:
            acc = [""]
            for p in parts:
                vals = self.eval_expr(p, ff, pos, depth + 1)
                acc = [a + v for a in acc for v in vals][:8]
            return acc
        for op in ("||", "??"):
            alts = split_top(e, op, "js")
            if len(alts) > 1:
                out = []
                for a in alts:
                    out += self.eval_expr(a, ff, pos, depth + 1)
                return out
        q = split_top(e, "?", "js")
        if len(q) == 2:
            alts = split_top(q[1], ":", "js")
            if len(alts) == 2:
                return self.eval_expr(alts[0], ff, pos, depth + 1) + self.eval_expr(alts[1], ff, pos, depth + 1)
        m = re.fullmatch(r"(?:this\.)?(\w+)\s*\(\s*\)", e)
        if m:
            name = m.group(1)
            if name in BASE_HELPERS:
                return [BASE_HELPERS[name]]
            d = find_function_return(ff, name, pos)
            if d:
                return self.eval_expr(d, ff, pos, depth + 1)
            return [UNKNOWN]
        m = re.fullmatch(r"(?:\w+\??\.)*public\??\.(\w+)", e) or re.fullmatch(r"useRuntimeConfig\(\)\.public\??\.(\w+)", e)
        if m and m.group(1) in CONFIG_KEYS:
            return [CONFIG_KEYS[m.group(1)]]
        if re.fullmatch(r"[A-Za-z_$][\w$]*", e):
            d = find_definition(ff, e, pos)
            if d is not None:
                return self.eval_expr(d, ff, pos, depth + 1)
        return [UNKNOWN]

    def eval_template(self, t, ff, pos, depth):
        body = t[1:-1]
        out = [""]
        i = 0
        first = True
        while i < len(body):
            if body.startswith("${", i):
                end = skip_js_braces(body, i + 1)
                inner = body[i + 2:end - 1]
                if first and i == 0:
                    vals = self.eval_expr(inner, ff, pos, depth + 1)
                else:
                    vals = [UNKNOWN]
                    if re.fullmatch(r"[A-Za-z_$][\w$]*", inner.strip()):
                        d = find_definition(ff, inner.strip(), pos)
                        if d is not None and js_literal_value(d) is not None and "/" not in js_literal_value(d)[:1]:
                            vals = [js_literal_value(d)]
                        elif d is not None and js_literal_value(d) is not None:
                            vals = [js_literal_value(d)]
                out = [o + v for o in out for v in vals][:8]
                i = end
            else:
                j = body.find("${", i)
                j = len(body) if j < 0 else j
                out = [o + body[i:j] for o in out]
                i = j
            first = False
        return out


def skip_generic(s, i):
    depth = 0
    j = i
    while j < len(s):
        c = s[j]
        if c in "\"'`":
            j = skip_js_string(s, j)
            continue
        if c == "<":
            depth += 1
        elif c == ">" and s[j - 1] != "=":
            depth -= 1
            if depth == 0:
                return j + 1
        elif c in "{[(":
            end, _ = read_group(s, j, "js")
            j = end
            continue
        j += 1
    return j


def iter_js_literals(code):
    i = 0
    n = len(code)
    while i < n:
        c = code[i]
        if c in "\"'`":
            end = skip_js_string(code, i)
            yield i, end, code[i:end]
            i = end
        else:
            i += 1


def find_definition(ff, name, pos):
    """Right-hand side of the nearest `const|let|var name = ...` (or `name = ...`) before pos."""
    code = ff.code
    best = None
    for m in re.finditer(rf"(?:\b(?:const|let|var)\s+{re.escape(name)}\b\s*(?::\s*[^=\n]+)?=(?!=)|(?<![\w.$]){re.escape(name)}\s*=(?![=>]))", code[:pos]):
        best = m
    if best is None:
        # module-level definition after pos (helpers declared lower in the file)
        m = re.search(rf"\b(?:const|let|var)\s+{re.escape(name)}\b\s*(?::\s*[^=\n]+)?=(?!=)", code)
        if not m:
            return None
        best = m
    return read_expression(code, best.end())


def find_function_return(ff, name, pos):
    code = ff.code
    m = None
    for mm in re.finditer(rf"\b(?:const|let)\s+{re.escape(name)}\s*=\s*(?:async\s*)?\([^)]*\)\s*(?::\s*[^=]+)?=>\s*", code):
        m = mm
    if m:
        k = m.end()
        if code[k] == "{":
            end, _ = read_group(code, k, "js")
            body = code[k + 1:end - 1]
            r = re.search(r"\breturn\s+", body)
            return read_expression(body, r.end()) if r else None
        return read_expression(code, k)
    m = re.search(rf"\bfunction\s+{re.escape(name)}\s*\([^)]*\)\s*(?::\s*[^{{]+)?\{{", code)
    if m:
        end, _ = read_group(code, m.end() - 1, "js")
        body = code[m.end():end - 1]
        r = re.search(r"\breturn\s+", body)
        return read_expression(body, r.end()) if r else None
    return None


def read_expression(code, k):
    j = k
    depth = 0
    while j < len(code):
        c = code[j]
        if c in "\"'`":
            j = skip_js_string(code, j)
            continue
        if c in "([{":
            depth += 1
        elif c in ")]}":
            if depth == 0:
                break
            depth -= 1
        elif depth == 0 and (c in ";\n" or (c == "," and depth == 0)):
            # a newline ends the expression unless the next line continues it
            if c == "\n":
                rest = code[j + 1:].lstrip()
                if rest.startswith(("+", "?", ":", ".", "||", "&&")) or code[:j].rstrip().endswith(("+", "?", ":", "(", "=", "||", "&&")):
                    j += 1
                    continue
            break
        j += 1
    return code[k:j].strip()


LOCALHOST_RE = re.compile(r"^https?://(localhost|127\.0\.0\.1)(:\d+)?")


def normalize_url(u, relative_api=False):
    if u is None:
        return None
    if LOCALHOST_RE.match(u):
        rest = normalize_url(LOCALHOST_RE.sub("", u), relative_api)
        return ("localhost:" + rest) if rest else None
    if relative_api and not u.startswith("/api") and not u.startswith(UNKNOWN):
        u = "/api/v1/" + u.lstrip("/")
    u = u.split("?")[0].split("#")[0]
    u = re.sub(r"(?<=[^/])\{\*\}$", "", u)
    u = re.sub(r"/{2,}", "/", u)
    if len(u) > 1:
        u = u.rstrip("/")
    if not (u.startswith("/api/") or u.startswith("/uploads/") or u == "/uploads"):
        return None
    return u


# ---------------------------------------------------------------------------
# matching frontend calls to routes
# ---------------------------------------------------------------------------

def seg_match(route_path, call_path):
    """Returns a specificity score, or None. Route params match any segment; call {*} matches any segment."""
    rs = route_path.strip("/").split("/")
    cs = call_path.strip("/").split("/")
    catch_all = rs and re.fullmatch(r"\{(filepath|path|any)\}", rs[-1])
    if catch_all:
        if len(cs) < len(rs):
            return None
    elif len(rs) != len(cs):
        return None
    score = 0
    for i, r in enumerate(rs):
        if catch_all and i == len(rs) - 1:
            break
        c = cs[i]
        if re.fullmatch(r"\{[^}]+\}", r):
            score += 1
            continue
        if c == r:
            score += 3
            continue
        if UNKNOWN in c:
            pat = "^" + re.escape(c).replace(re.escape(UNKNOWN), ".+") + "$"
            if re.match(pat, r):
                score += 1
                continue
        return None
    return score


# ---------------------------------------------------------------------------
# the build
# ---------------------------------------------------------------------------

def svc_short(name):
    return {"auth-service": "auth", "audit-service": "audit", "master-service": "master", "risk-service": "risk",
            "analytics-service": "analytics", "python-ai": "pyai", "nuxt": "nuxt"}[name]


def module_of(public_path):
    p = public_path
    for base in ("/api/v1/", "/api/analytics/", "/api/python-ai/", "/api/"):
        if p.startswith(base):
            rest = p[len(base):]
            seg = rest.split("/")[0]
            if base == "/api/python-ai/" and seg == "predict":
                return "predict"
            return seg or "root"
    return p.strip("/").split("/")[0] or "root"


def op_id(method, path, used):
    p = re.sub(r"^/api/(v1/)?", "", path)
    words = []
    for seg in p.split("/"):
        if not seg:
            continue
        m = re.fullmatch(r"\{(\w+)\}", seg)
        if m:
            words.append("By" + m.group(1)[:1].upper() + m.group(1)[1:])
        else:
            words.append("".join(w[:1].upper() + w[1:] for w in re.split(r"[-_.]", seg)))
    base = method.lower() + "".join(words)
    oid, n = base, 2
    while oid in used:
        oid = f"{base}{n}"
        n += 1
    used.add(oid)
    return oid


def summary_for(route, info):
    if info.swag.get("summary"):
        return info.swag["summary"]
    model = info.model
    if info.kind.startswith("crud.") and model:
        noun = words_from_camel(model)
        return {"crud.list": f"List {noun} records (paged)", "crud.get": f"Get one {noun}", "crud.create": f"Create a {noun}",
                "crud.update": f"Update a {noun} (partial)", "crud.delete": f"Delete a {noun}"}[info.kind]
    if info.name and info.kind == "func":
        return words_from_camel(info.name.split(".")[-1]).capitalize()
    segs = [s for s in route["public_path"].split("/") if s and not s.startswith("{") and s not in ("api", "v1")]
    noun = " ".join(segs[1:] or segs[:1]).replace("-", " ")
    verb = {"GET": "Get" if route["public_path"].endswith("}") else "List", "POST": "Create", "PUT": "Update", "PATCH": "Update", "DELETE": "Delete", "HEAD": "Check", "OPTIONS": "Options for"}[route["method"]]
    return f"{verb} {noun}".strip()


class Build:
    def __init__(self):
        self.cfg = load_yaml(TOOLS / "config.yaml")
        self.common = load_yaml(OVERLAYS / "common.yaml", {})
        self.frontend_overlay = load_yaml(OVERLAYS / "frontend.yaml", {}) or {}
        self.kong = Kong(ROOT / self.cfg["kong"]["prod"])
        self.kong_dev = Kong(ROOT / self.cfg["kong"]["dev"])
        self.components = {}
        self.routes = []
        self.problems = []  # (severity, message)

    def run(self):
        self.collect_go()
        self.collect_python_ai()
        self.collect_nuxt()
        self.map_kong()
        self.scan_frontend()
        self.match_frontend()
        spec = self.assemble()
        return spec

    # -- collection -------------------------------------------------------------
    def collect_go(self):
        for name, scfg in self.cfg["services"].items():
            if "route_files" not in scfg:
                continue
            idx = GoIndex(ROOT / scfg["dir"])
            sb = SchemaBuilder(svc_short(name), idx, self.components)
            an = Analyzer(idx, sb)
            listed = {(ROOT / scfg["dir"] / rf["file"]).resolve() for rf in scfg["route_files"]}
            ignored = {(ROOT / scfg["dir"] / f).resolve() for f in scfg.get("ignore_files", [])}
            for path, (pkg, src, st) in idx.files.items():
                if path.resolve() in listed or path.resolve() in ignored:
                    continue
                if re.search(r"\.(GET|POST|PUT|PATCH|DELETE)\(\s*\"/", st) and not re.search(r"httptest", st):
                    self.problems.append(("error", f"{path.relative_to(ROOT)} registers Gin routes but is not listed in docs/api/tools/config.yaml"))
            for rf in scfg["route_files"]:
                for r in extract_routes(idx, name, rf):
                    r["info"] = an.analyze(r)
                    auth, roles, notes = describe_middleware(r["middleware"], idx)
                    r.update(auth=auth, roles=roles, mw_notes=notes, origin="gin")
                    del r["assign_before"]
                    self.routes.append(r)

    def collect_python_ai(self):
        scfg = self.cfg["services"]["python-ai"]
        fa = scfg["fastapi"]
        main_py = ROOT / scfg["dir"] / fa["source"]
        snap = load_json(ROOT / fa["snapshot"])
        snap_ops = set()
        if snap:
            for p, ops in snap.get("paths", {}).items():
                for m in ops:
                    snap_ops.add((m.upper(), p))
        for method, path, line in fastapi_routes(main_py):
            info = HandlerInfo()
            info.kind = "fastapi"
            info.loc = f"{main_py.relative_to(ROOT).as_posix()}:{line}"
            op = copy.deepcopy(((snap or {}).get("paths", {}).get(path, {}) or {}).get(method.lower()))
            if snap and (method, path) not in snap_ops:
                self.problems.append(("error", f"python-ai {method} {path} is not in the OpenAPI snapshot; run `apidocs.py refresh-python-ai`"))
            self.routes.append({"service": "python-ai", "method": method, "internal_path": path, "gin_path": path,
                                "file": main_py, "line": line, "info": info, "auth": False, "roles": [], "mw_notes": [],
                                "origin": "fastapi", "fastapi_op": op})
        live = {(m, p) for m, p, _ in fastapi_routes(main_py)}
        for m, p in sorted(snap_ops - live):
            self.problems.append(("error", f"python-ai snapshot has {m} {p}, which main.py no longer registers; run `apidocs.py refresh-python-ai`"))
        if snap:
            for k, v in (snap.get("components", {}).get("schemas", {}) or {}).items():
                self.components[f"pyai.{k}"] = rewrite_refs(copy.deepcopy(v), "pyai.")

    def collect_nuxt(self):
        scfg = self.cfg["services"]["nuxt"]
        for method, path, f in nitro_routes(ROOT / scfg["nitro_dir"]):
            info = HandlerInfo()
            info.kind = "nitro"
            info.loc = f.relative_to(ROOT).as_posix()
            src = f.read_text()
            if "readBody" in src:
                info.body_kind = "json"
            self.routes.append({"service": "nuxt", "method": method or "POST", "internal_path": path, "gin_path": path,
                                "file": f, "line": 1, "info": info, "auth": False, "roles": [], "mw_notes": [],
                                "origin": "nitro", "any_method": method is None})

    def map_kong(self):
        for r in self.routes:
            if r["service"] == "nuxt":
                r["public_path"] = r["internal_path"]
                r["routed_prod"] = True
                r["routed_dev"] = True
                r["aliases"] = []
                r["kong_route"] = "(Nuxt/Nitro server route, not Kong)"
                continue
            pubs = self.kong.public_paths(r["service"], r["internal_path"])
            devs = self.kong_dev.public_paths(r["service"], r["internal_path"])
            r["routed_prod"] = bool(pubs)
            r["routed_dev"] = bool(devs)
            if pubs:
                r["public_path"] = pubs[0][0]
                r["aliases"] = [p for p, _ in pubs[1:]]
                r["kong_route"] = pubs[0][1]["route"]
            elif devs:
                r["public_path"] = devs[0][0]
                r["aliases"] = [p for p, _ in devs[1:]]
                r["kong_route"] = devs[0][1]["route"] + " (dev Kong only)"
            else:
                r["public_path"] = r["internal_path"]
                r["aliases"] = []
                r["kong_route"] = None

    def scan_frontend(self):
        self.fe = FrontendScanner(self.cfg["frontend"])
        self.fe.scan()
        ignore = self.frontend_overlay.get("ignore", []) or []
        manual = self.frontend_overlay.get("manual_calls", []) or []
        manual_sites = {(m["file"], m.get("line")) for m in manual}

        def ignored(rec):
            if (rec["file"], rec["line"]) in manual_sites:
                return True
            for ig in ignore:
                if rec["file"] == ig.get("file") and (not ig.get("line") or ig["line"] == rec["line"]):
                    return True
            return False
        self.fe.unresolved = [u for u in self.fe.unresolved if not ignored(u)]
        self.fe.calls = [c for c in self.fe.calls if not ignored(c)]
        # manual resolutions for calls the scanner cannot evaluate
        for man in manual:
            self.fe.calls.append({"file": man["file"], "line": man.get("line", 0), "methods": [man["method"]],
                                  "paths": [man["path"]], "raw": man.get("note", "manual"), "client": "manual"})

    def match_frontend(self):
        for r in self.routes:
            r["used_by"] = []
            r["referenced_by"] = []
        self.mismatches = []
        by_path_cache = {}

        def candidates(path):
            if path in by_path_cache:
                return by_path_cache[path]
            res = []
            for r in self.routes:
                paths = [r["public_path"]] + r.get("aliases", [])
                if r["service"] != "nuxt" and not r["routed_prod"] and not r["routed_dev"]:
                    paths = [r["public_path"]]
                for pp in paths:
                    s = seg_match(pp, path)
                    if s is not None:
                        res.append((s, r, pp))
            by_path_cache[path] = res
            return res

        for c in self.fe.calls:
            for path in c["paths"]:
                if path.startswith("localhost:"):
                    self.mismatches.append(dict(c, path=path[len("localhost:"):], method="/".join(c["methods"]), kind="hardcoded-localhost",
                                                detail="The URL can resolve to a hard-coded http://localhost:<port> address, which in production points at the visitor's own machine"))
                    continue
                cands = candidates(path)
                if not cands:
                    self.mismatches.append(dict(c, path=path, kind="not-registered",
                                                detail="No backend route matches this path"))
                    continue
                best = max(s for s, _, _ in cands)
                top = [(r, pp) for s, r, pp in cands if s == best]
                for meth in c["methods"]:
                    hits = [(r, pp) for r, pp in top if meth == "?" or r["method"] == meth or r.get("any_method")]
                    if not hits:
                        # a less specific route with the right method?
                        hits = [(r, pp) for s, r, pp in sorted(cands, key=lambda x: -x[0]) if r["method"] == meth]
                        if hits:
                            hs = max(seg_match(pp, path) for _, pp in hits)
                            hits = [(r, pp) for r, pp in hits if seg_match(pp, path) == hs]
                    if not hits:
                        have = sorted({r["method"] for r, _ in top}, key=lambda m: METHOD_ORDER.get(m, 9))
                        self.mismatches.append(dict(c, path=path, method=meth, kind="method-not-registered",
                                                    detail=f"Path exists ({top[0][1]}) but only for {', '.join(have)}"))
                        continue
                    for r, pp in hits:
                        r["used_by"].append({"file": c["file"], "line": c["line"], "method": meth, "via": pp})
                        if not r["routed_prod"] and r["service"] != "nuxt":
                            self.mismatches.append(dict(c, path=path, method=meth, kind="not-routed-prod",
                                                        detail=f"{r['service']} registers {r['method']} {r['internal_path']} but production Kong does not route it" + (" (dev Kong does)" if r["routed_dev"] else " (dev Kong neither)")))
                        elif pp != r["public_path"]:
                            self.mismatches.append(dict(c, path=path, method=meth, kind="legacy-alias",
                                                        detail=f"Uses the legacy Kong alias {pp}; canonical path is {r['public_path']}"))
        for ref in self.fe.references:
            if ref["path"].startswith("localhost:"):
                self.mismatches.append(dict(ref, path=ref["path"][len("localhost:"):], methods=["?"], kind="hardcoded-localhost",
                                            detail="Hard-coded http://localhost:<port> URL"))
                continue
            cands = candidates(ref["path"])
            if not cands:
                self.mismatches.append(dict(ref, methods=["?"], kind="not-registered",
                                            detail="Referenced URL (not a $fetch call) matches no backend route"))
                continue
            best = max(s for s, _, _ in cands)
            top = [(r, pp) for s, r, pp in cands if s == best]
            gets = [(r, pp) for r, pp in top if r["method"] == "GET"] or top
            for r, pp in gets:
                r["referenced_by"].append({"file": ref["file"], "line": ref["line"], "via": pp})
                if not r["routed_prod"] and r["service"] != "nuxt":
                    self.mismatches.append(dict(ref, methods=["?"], path=ref["path"], kind="not-routed-prod",
                                                detail=f"{r['service']} registers {r['method']} {r['internal_path']} but production Kong does not route it"))
        # dedupe mismatches
        seen = set()
        out = []
        for m in self.mismatches:
            key = (m["file"], m["line"], m.get("path"), m.get("method"), m["kind"])
            if key in seen:
                continue
            seen.add(key)
            out.append(m)
        self.mismatches = sorted(out, key=lambda m: (m["kind"], m.get("path", ""), m["file"], m["line"]))

    # -- assembly ------------------------------------------------------------------
    def assemble(self):
        base = load_yaml(OVERLAYS / "common.yaml", {})
        spec = OrderedDict()
        spec["openapi"] = "3.1.0"
        spec["info"] = base.get("info", {})
        spec["servers"] = base.get("servers", [])
        spec["tags"] = []
        spec["paths"] = OrderedDict()
        comps = OrderedDict()
        comps["securitySchemes"] = base.get("components", {}).get("securitySchemes", {})
        comps["responses"] = base.get("components", {}).get("responses", {})
        common_schemas = base.get("components", {}).get("schemas", {})
        overlays = {}
        for name in self.cfg["services"]:
            overlays[name] = load_yaml(OVERLAYS / f"{name}.yaml", {}) or {}
        tag_desc = {}
        for name, ov in overlays.items():
            for t in ov.get("tags", []) or []:
                tag_desc[t["name"]] = t.get("description", "")
        used_ids = set()
        tags_seen = OrderedDict()
        ov_keys_used = set()
        self.doc_level = {}
        documented = [r for r in self.routes if r["routed_prod"] or r["routed_dev"]]
        documented.sort(key=lambda r: (r["public_path"], METHOD_ORDER.get(r["method"], 9)))
        svc_cfg = self.cfg["services"]
        for r in documented:
            info: HandlerInfo = r["info"]
            tag = f"{r['service']}: {module_of(r['public_path'])}"
            if r["public_path"].startswith("/uploads"):
                tag = "audit-service: uploads (static files)"
            tags_seen.setdefault(tag, r["service"])
            key = f"{r['method']} {r['public_path']}"
            op = OrderedDict()
            if r["origin"] == "fastapi" and r.get("fastapi_op"):
                fop = rewrite_refs(r["fastapi_op"], "pyai.")
                op.update({k: v for k, v in fop.items() if k not in ("tags", "operationId")})
            op["tags"] = [tag]
            op["summary"] = op.get("summary") or summary_for(r, info)
            op["operationId"] = op_id(r["method"], r["public_path"], used_ids)
            desc = []
            if info.swag.get("description") and info.swag.get("description") != op["summary"]:
                desc.append(info.swag["description"])
            desc += [n for n in info.notes]
            desc += r["mw_notes"]
            if info.perms:
                desc.append("Permission check in the handler (audit-service ATR role matrix): " + ", ".join(f"`{p}`" for p in info.perms) + ".")
            if r.get("aliases"):
                desc.append("Also reachable through the legacy Kong alias " + ", ".join(f"`{a}`" for a in r["aliases"]) + ".")
            if not r["routed_prod"]:
                desc.append("**Not routed by production Kong** (only by the dev Kong config): unreachable on https://auditsphere.app.")
            if desc:
                op["description"] = "\n\n".join(desc)
            # parameters
            params = []
            for pname in re.findall(r"\{(\w+)\}", r["public_path"]):
                ps = {"type": "string"}
                if pname in info.uuid_params:
                    ps["format"] = "uuid"
                p = {"name": pname, "in": "path", "required": True, "schema": ps}
                if r["gin_path"].endswith("*" + pname):
                    p["description"] = "Rest of the path (may contain `/`)"
                params.append(p)
            for qn, q in info.query.items():
                p = {"name": qn, "in": "query", "required": bool(q.get("required")), "schema": q.get("schema", {"type": "string"})}
                if q.get("description"):
                    p["description"] = q["description"]
                params.append(p)
            if params and "parameters" not in op:
                op["parameters"] = params
            elif params:
                op["parameters"] = params + [p for p in op["parameters"] if (p["name"], p["in"]) not in {(x["name"], x["in"]) for x in params}]
            # request body
            if r["origin"] != "fastapi" and r["method"] in ("POST", "PUT", "PATCH", "DELETE"):
                rb = None
                if info.body_kind == "multipart":
                    props = OrderedDict()
                    props.update(info.files)
                    props.update(info.form_fields)
                    rb = {"required": True, "content": {"multipart/form-data": {"schema": {"type": "object", "properties": dict(props)}}}}
                elif info.body_kind == "json":
                    rb = {"required": True, "content": {"application/json": {"schema": info.body_schema or {"type": "object"}}}}
                if rb:
                    op["requestBody"] = rb
            # responses
            if r["origin"] != "fastapi" or "responses" not in op:
                op["responses"] = self.responses_for(r, info)
            else:
                op["responses"] = dict(op["responses"])
                for code, desc_ in self.error_responses(r, info).items():
                    op["responses"].setdefault(code, desc_)
            # security
            op["security"] = self.security_for(r)
            no_body = (r["origin"] in ("gin",) and info.kind != "unknown" and r["method"] in ("POST", "PUT", "PATCH")
                       and "requestBody" not in op)
            # extension with provenance and usage
            ext = OrderedDict()
            ext["service"] = r["service"]
            ext["handler"] = info.name or info.kind
            ext["source"] = info.loc or f"{r['file'].relative_to(ROOT).as_posix()}:{r['line']}"
            ext["registeredAt"] = f"{r['file'].relative_to(ROOT).as_posix()}:{r['line']}"
            ext["internalPath"] = r["internal_path"]
            ext["kongRoute"] = r["kong_route"]
            if no_body:
                ext["noRequestBody"] = True
            if r["roles"]:
                ext["requiredRoles"] = [{"rule": k, "roles": v} for k, v in r["roles"]]
            if info.perms:
                ext["requiredPermissions"] = info.perms
            usage = sorted({f"{u['file']}:{u['line']}" for u in r["used_by"]} | {f"{u['file']}:{u['line']} (reference)" for u in r["referenced_by"]})
            ext["usedByWebsite"] = bool(usage)
            if usage:
                ext["websiteCallSites"] = usage[:12] + ([f"... {len(usage) - 12} more"] if len(usage) > 12 else [])
            op["x-auditsphere"] = ext
            # overlay
            ov_ops = overlays.get(r["service"], {}).get("operations", {}) or {}
            if key in ov_ops:
                ov_keys_used.add((r["service"], key))
                op = apply_op_overlay(op, ov_ops[key])
            level = doc_level(r, op)
            ext = op["x-auditsphere"]
            ext["docLevel"] = level
            self.doc_level[(r["service"], key)] = level
            spec["paths"].setdefault(r["public_path"], OrderedDict())[r["method"].lower()] = op
        # overlay keys that no longer exist
        for name, ov in overlays.items():
            for key in (ov.get("operations") or {}):
                if (name, key) not in ov_keys_used:
                    self.problems.append(("error", f"overlays/{name}.yaml documents `{key}`, which is not a routed {name} endpoint"))
            for k, v in ((ov.get("components") or {}).get("schemas") or {}).items():
                self.components[k] = v
        for tag, svc in tags_seen.items():
            spec["tags"].append({"name": tag, "description": tag_desc.get(tag, f"{svc_cfg[svc]['title']} ({svc})")})
        spec["tags"].sort(key=lambda t: t["name"])
        spec["x-tagGroups"] = [{"name": svc_cfg[s]["title"], "tags": sorted(t for t, sv in tags_seen.items() if sv == s)}
                               for s in svc_cfg if any(sv == s for sv in tags_seen.values())]
        schemas = OrderedDict()
        schemas.update(common_schemas)
        used = collect_refs(spec["paths"]) | collect_refs(common_schemas)
        # transitive closure
        pending = list(used)
        while pending:
            n = pending.pop()
            s = self.components.get(n) or common_schemas.get(n)
            if s is None:
                continue
            for m in collect_refs(s):
                if m not in used:
                    used.add(m)
                    pending.append(m)
        for n in sorted(self.components):
            if n in used and n not in schemas:
                schemas[n] = self.components[n]
        missing = sorted(n for n in used if n not in schemas)
        for n in missing:
            self.problems.append(("error", f"schema {n} is referenced but not defined"))
        comps["schemas"] = schemas
        spec["components"] = comps
        return spec

    def responses_for(self, r, info):
        out = OrderedDict()
        for code in sorted(info.responses):
            resp = info.responses[code]
            out[str(code)] = response_obj(resp, code)
        if not out:
            if r["service"] == "audit-service" and r["public_path"].startswith("/uploads"):
                out["200"] = {"description": "The file", "content": {"application/octet-stream": {"schema": {"type": "string", "format": "binary"}}}}
            else:
                out["200"] = {"description": "Success (shape not inferred; see the handler)", "content": {"application/json": {"schema": {}}}}
        out.update(self.error_responses(r, info))
        return out

    def error_responses(self, r, info):
        out = OrderedDict()
        codes = set(info.error_statuses)
        if r["auth"]:
            codes.add(401)
        if r["roles"] or info.perms:
            codes.add(403)
        if r["service"] in ("auth-service", "audit-service", "master-service", "risk-service", "analytics-service") and r["routed_prod"]:
            codes.add(429)
        for code in sorted(codes):
            if code < 400:
                continue
            if code == 429:
                out["429"] = {"$ref": "#/components/responses/RateLimited"}
            elif code == 401:
                out["401"] = {"$ref": "#/components/responses/Unauthorized"}
            else:
                out[str(code)] = {"description": ERROR_TEXT.get(code, "Error"),
                                  "content": {"application/json": {"schema": {"$ref": "#/components/schemas/common.ErrorEnvelope"}}}}
        return out

    def security_for(self, r):
        if r["service"] == "python-ai":
            return [{}, {"pythonAiApiKey": []}]
        if not r["auth"]:
            return []
        if r["service"] == "audit-service":
            return [{"bearerAuth": []}, {"cookieAuth": []}, {"tokenQuery": []}]
        if r["service"] in ("risk-service", "analytics-service"):
            return [{"bearerAuth": []}, {"cookieAuth": []}]
        return [{"bearerAuth": []}]


def response_obj(resp, code):
    if resp["kind"] == "binary":
        ct = resp.get("content_type") or "application/octet-stream"
        o = {"description": "File content", "content": {ct: {"schema": {"type": "string", "format": "binary"}}}}
        if resp.get("alt") and resp["alt"].get("kind") == "json":
            o["content"]["application/json"] = {"schema": resp["alt"].get("schema") or {}}
        return o
    if resp["kind"] == "redirect":
        return {"description": "Redirect to the file location", "headers": {"Location": {"schema": {"type": "string"}}}}
    if resp["kind"] == "empty":
        return {"description": "No content"}
    return {"description": "Created" if code == 201 else "Success",
            "content": {"application/json": {"schema": resp.get("schema") or {}}}}


def apply_op_overlay(op, ov):
    op = copy.deepcopy(op)
    for k, v in ov.items():
        if k == "description_append":
            op["description"] = (op.get("description", "") + "\n\n" + v).strip()
        elif k == "parameters":
            existing = {(p["name"], p["in"]): i for i, p in enumerate(op.get("parameters", []))}
            params = list(op.get("parameters", []))
            for p in v:
                if (p["name"], p["in"]) in existing:
                    params[existing[(p["name"], p["in"])]] = deep_merge(params[existing[(p["name"], p["in"])]], p)
                else:
                    params.append(p)
            op["parameters"] = params
        elif k == "responses":
            resps = OrderedDict(op.get("responses", {}))
            for code, rv in v.items():
                resps[str(code)] = rv
            op["responses"] = OrderedDict(sorted(resps.items(), key=lambda kv: kv[0]))
        elif k == "x-auditsphere":
            op["x-auditsphere"] = deep_merge(op["x-auditsphere"], v)
        else:
            op[k] = copy.deepcopy(v)
    return op


def doc_level(r, op):
    ok_req = True
    if r["method"] in ("POST", "PUT", "PATCH"):
        rb = op.get("requestBody")
        ok_req = bool(rb) or bool(op.get("x-auditsphere", {}).get("noRequestBody"))
        if rb:
            for media in rb.get("content", {}).values():
                if media.get("schema") in ({}, {"type": "object"}):
                    ok_req = False
    ok_resp = False
    for code, resp in op.get("responses", {}).items():
        if not str(code).startswith("2"):
            continue
        if "$ref" in resp:
            ok_resp = True
            continue
        content = resp.get("content")
        if not content:
            ok_resp = True
            continue
        for ct, media in content.items():
            s = media.get("schema")
            if ct != "application/json" or (s and not is_generic_envelope(s) and s != {}):
                ok_resp = True
    if ok_req and ok_resp:
        return "full"
    if ok_req or ok_resp:
        return "partial"
    return "skeleton"


def collect_refs(obj):
    out = set()
    if isinstance(obj, dict):
        for k, v in obj.items():
            if k == "$ref" and isinstance(v, str) and v.startswith("#/components/schemas/"):
                out.add(v.split("/")[-1])
            else:
                out |= collect_refs(v)
    elif isinstance(obj, list):
        for v in obj:
            out |= collect_refs(v)
    return out


def rewrite_refs(obj, prefix):
    if isinstance(obj, dict):
        return {k: (("#/components/schemas/" + prefix + v.split("/")[-1]) if k == "$ref" and isinstance(v, str) and v.startswith("#/components/schemas/") else rewrite_refs(v, prefix)) for k, v in obj.items()}
    if isinstance(obj, list):
        return [rewrite_refs(v, prefix) for v in obj]
    return obj


def load_json(path: Path):
    if not path.exists():
        return None
    return json.loads(path.read_text())


# ---------------------------------------------------------------------------
# outputs
# ---------------------------------------------------------------------------

class _Dumper(yaml.SafeDumper):
    pass


def _dict_rep(dumper, data):
    return dumper.represent_mapping("tag:yaml.org,2002:map", data.items())


def _str_rep(dumper, data):
    if "\n" in data:
        return dumper.represent_scalar("tag:yaml.org,2002:str", data, style="|")
    return dumper.represent_scalar("tag:yaml.org,2002:str", data)


_Dumper.add_representer(OrderedDict, _dict_rep)
_Dumper.add_representer(str, _str_rep)
_Dumper.ignore_aliases = lambda self, data: True


def dump_yaml(obj):
    return yaml.dump(obj, Dumper=_Dumper, sort_keys=False, allow_unicode=True, width=120)


def status_of(r):
    if r["service"] == "nuxt":
        return "website" if (r["used_by"] or r["referenced_by"]) else "unused"
    if not r["routed_prod"] and not r["routed_dev"]:
        return "internal"
    if r["used_by"] or r["referenced_by"]:
        return "website"
    return "unused"


def inventory(build: Build):
    rows = []
    for r in sorted(build.routes, key=lambda r: (r["service"], r["public_path"], METHOD_ORDER.get(r["method"], 9))):
        st = status_of(r)
        key = f"{r['method']} {r['public_path']}"
        rows.append(OrderedDict([
            ("service", r["service"]), ("method", r["method"]), ("public_path", r["public_path"] if (r["routed_prod"] or r["routed_dev"] or r["service"] == "nuxt") else None),
            ("internal_path", r["internal_path"]), ("status", st),
            ("routed_prod", r["routed_prod"]), ("routed_dev", r["routed_dev"]),
            ("auth", "jwt" if r["auth"] else ("api-key-if-configured" if r["service"] == "python-ai" else "none")),
            ("roles", [{"rule": k, "roles": v} for k, v in r["roles"]]),
            ("permissions", r["info"].perms),
            ("handler", r["info"].name or r["info"].kind), ("source", r["info"].loc),
            ("registered_at", f"{r['file'].relative_to(ROOT).as_posix()}:{r['line']}"),
            ("doc_level", build.doc_level.get((r["service"], key))),
            ("website_call_sites", sorted({f"{u['file']}:{u['line']}" for u in r["used_by"]})),
            ("website_references", sorted({f"{u['file']}:{u['line']}" for u in r["referenced_by"]})),
        ]))
    stats = OrderedDict()
    for row in rows:
        s = stats.setdefault(row["service"], OrderedDict([("registered", 0), ("documented", 0), ("website", 0), ("unused", 0), ("internal", 0),
                                                           ("website_full", 0), ("website_partial", 0), ("website_skeleton", 0)]))
        s["registered"] += 1
        if row["status"] != "internal":
            s["documented"] += 1
        s[row["status"]] += 1
        if row["status"] == "website" and row["doc_level"]:
            s[f"website_{row['doc_level']}"] += 1
    return rows, stats


def kong_findings(build: Build):
    out = []
    for kong, label in ((build.kong, "prod"), (build.kong_dev, "dev")):
        for kr in kong.routes:
            svc = Kong.target_service(kr["host"])
            hits = []
            for r in build.routes:
                if r["service"] != svc:
                    continue
                for p in kong.public_paths(svc, r["internal_path"]):
                    if kong.match(re.sub(r"\{[^}]+\}", "x", p[0])) is kr:
                        hits.append(r)
            if not hits:
                out.append(f"{label} Kong path `{kr['prefix']}` (service `{kr['kong_service']}`) matches no route that {svc} registers")
    for r in sorted(build.routes, key=lambda r: (r["service"], r["internal_path"])):
        if r["service"] == "nuxt":
            continue
        if r["routed_dev"] and not r["routed_prod"]:
            out.append(f"`{r['method']} {r['internal_path']}` ({r['service']}) is routed by dev Kong but not by prod Kong")
        elif r["routed_prod"] and not r["routed_dev"]:
            out.append(f"`{r['method']} {r['internal_path']}` ({r['service']}) is routed by prod Kong but not by dev Kong")
        elif not r["routed_prod"] and re.match(r"^/api/", r["internal_path"]):
            out.append(f"`{r['method']} {r['internal_path']}` ({r['service']}) is registered under /api but neither Kong config routes it")
    return out


def render_inventory_md(build: Build, rows, stats, known):
    L = []
    L.append("# AuditSphere API inventory")
    L.append("")
    L.append("Generated by `docs/api/tools/apidocs.py` from the Gin route registrations, the production Kong config, "
             "python-ai's FastAPI app, the Nuxt server routes and a static scan of the website's API calls. Do not edit by hand.")
    L.append("")
    L.append("Status: **website** = the frontend calls it; **unused** = registered and routed by Kong but no frontend call found; "
             "**internal** = registered but not routed by Kong (reachable only inside the Docker network or on the service's own port).")
    L.append("")
    L.append("## Counts")
    L.append("")
    L.append("| Service | Registered | Documented (routed) | Used by website | Registered, unused | Internal only | Website ops: full / partial / skeleton docs |")
    L.append("|---|---:|---:|---:|---:|---:|---|")
    tot = defaultdict(int)
    for svc, s in stats.items():
        L.append(f"| {svc} | {s['registered']} | {s['documented']} | {s['website']} | {s['unused']} | {s['internal']} | {s['website_full']} / {s['website_partial']} / {s['website_skeleton']} |")
        for k, v in s.items():
            tot[k] += v
    L.append(f"| **total** | {tot['registered']} | {tot['documented']} | {tot['website']} | {tot['unused']} | {tot['internal']} | {tot['website_full']} / {tot['website_partial']} / {tot['website_skeleton']} |")
    L.append("")
    L.append("## Website calls that hit unregistered or unrouted endpoints")
    L.append("")
    bad = [m for m in build.mismatches if m["kind"] != "legacy-alias"]
    if not bad:
        L.append("None.")
    else:
        L.append("| Kind | Call | Where | Detail | Known issue |")
        L.append("|---|---|---|---|---|")
        for m in bad:
            meth = m.get("method") or "/".join(m.get("methods", ["?"]))
            k = mismatch_key(m)
            L.append(f"| {m['kind']} | `{meth} {m['path']}` | `{m['file']}:{m['line']}` | {m['detail']} | {known.get(k, '') or 'NEW'} |")
    alias = [m for m in build.mismatches if m["kind"] == "legacy-alias"]
    if alias:
        L.append("")
        L.append("Calls through legacy Kong aliases (work, but should move to the canonical path):")
        L.append("")
        for m in alias:
            L.append(f"- `{m.get('method', '?')} {m['path']}` at `{m['file']}:{m['line']}`: {m['detail']}")
    L.append("")
    L.append("## Website calls the scanner could not resolve")
    L.append("")
    if build.fe.unresolved:
        L.append("These build the URL at runtime. Resolve them in `overlays/frontend.yaml` (`manual_calls` or `ignore`).")
        L.append("")
        for u in build.fe.unresolved:
            L.append(f"- `{u['file']}:{u['line']}` `{'/'.join(u['methods'])}` `{u['raw']}`")
    else:
        L.append("None.")
    L.append("")
    L.append("## Kong observations")
    L.append("")
    for f in kong_findings(build) or ["None."]:
        L.append(f"- {f}")
    L.append("")
    L.append("## All routes")
    for svc in stats:
        L.append("")
        L.append(f"### {svc}")
        L.append("")
        L.append("| Method | Public path | Status | Auth / roles | Docs | Handler | Website call sites |")
        L.append("|---|---|---|---|---|---|---|")
        for row in rows:
            if row["service"] != svc:
                continue
            path = row["public_path"] or f"(internal) {row['internal_path']}"
            auth = row["auth"]
            if row["roles"]:
                auth += "; " + "; ".join(f"{x['rule']}: {', '.join(x['roles'])}" for x in row["roles"])
            if row["permissions"]:
                auth += "; perms: " + ", ".join(row["permissions"])
            sites = row["website_call_sites"] + [s + " (ref)" for s in row["website_references"]]
            site_txt = ", ".join(f"`{s}`" for s in sites[:3]) + (f" +{len(sites) - 3}" if len(sites) > 3 else "")
            L.append(f"| {row['method']} | `{path}` | {row['status']} | {auth} | {row['doc_level'] or '-'} | `{row['handler']}` | {site_txt} |")
    L.append("")
    return "\n".join(L)


def mismatch_key(m):
    return f"{m['kind']} {m.get('method') or '/'.join(m.get('methods', ['?']))} {m['path']} @ {m['file']}"


def build_all():
    b = Build()
    spec = b.run()
    rows, stats = inventory(b)
    known = {k["key"]: k.get("note", "known") for k in (b.frontend_overlay.get("known_issues") or [])}
    md = render_inventory_md(b, rows, stats, known)
    inv = OrderedDict([("generated_by", "docs/api/tools/apidocs.py"), ("counts", stats),
                       ("frontend_mismatches", [OrderedDict([("kind", m["kind"]), ("method", m.get("method") or "/".join(m.get("methods", ["?"]))),
                                                             ("path", m["path"]), ("file", m["file"]), ("line", m["line"]), ("detail", m["detail"]),
                                                             ("key", mismatch_key(m))]) for m in b.mismatches]),
                       ("frontend_unresolved", b.fe.unresolved), ("kong_observations", kong_findings(b)), ("routes", rows)])
    for m in b.mismatches:
        if m["kind"] != "legacy-alias" and mismatch_key(m) not in known:
            b.problems.append(("error", f"website call {mismatch_key(m)} (line {m['line']}): {m['detail']}. Fix it or add it to overlays/frontend.yaml known_issues"))
    for k in known:
        if k not in {mismatch_key(m) for m in b.mismatches}:
            b.problems.append(("warning", f"known issue `{k}` no longer occurs; remove it from overlays/frontend.yaml"))
    return b, spec, md, inv


def to_json(obj):
    return json.dumps(obj, indent=2, default=str, ensure_ascii=False) + "\n"


OUT_FILES = {"openapi.yaml": None, "inventory.md": None, "inventory.json": None}


def cmd_generate(args):
    b, spec, md, inv = build_all()
    (DOCS / "openapi.yaml").write_text(dump_yaml(spec))
    (DOCS / "inventory.md").write_text(md)
    (DOCS / "inventory.json").write_text(to_json(inv))
    report(b, inv)
    return 0


def cmd_check(args):
    b, spec, md, inv = build_all()
    stale = []
    for name, content in (("openapi.yaml", dump_yaml(spec)), ("inventory.md", md), ("inventory.json", to_json(inv))):
        p = DOCS / name
        if not p.exists() or p.read_text() != content:
            stale.append(name)
    for s in stale:
        b.problems.append(("error", f"docs/api/{s} is out of date: run `python3 docs/api/tools/apidocs.py generate`"))
    report(b, inv)
    return 1 if any(sev == "error" for sev, _ in b.problems) else 0


def report(b, inv):
    for svc, s in inv["counts"].items():
        print(f"{svc:18} registered={s['registered']:3} documented={s['documented']:3} website={s['website']:3} unused={s['unused']:3} internal={s['internal']:2} "
              f"website-docs full/partial/skeleton={s['website_full']}/{s['website_partial']}/{s['website_skeleton']}")
    print(f"frontend calls: {len(b.fe.calls)} resolved, {len(b.fe.references)} extra references, {len(b.fe.unresolved)} unresolved; "
          f"{len([m for m in b.mismatches if m['kind'] != 'legacy-alias'])} mismatches")
    for sev, msg in b.problems:
        print(f"{sev.upper()}: {msg}")


def cmd_refresh_python_ai(args):
    cfg = load_yaml(TOOLS / "config.yaml")
    dest = ROOT / cfg["services"]["python-ai"]["fastapi"]["snapshot"]
    with urllib.request.urlopen(args.url, timeout=10) as resp:
        data = json.loads(resp.read())
    dest.parent.mkdir(parents=True, exist_ok=True)
    dest.write_text(json.dumps(data, indent=2, sort_keys=False) + "\n")
    print(f"wrote {dest.relative_to(ROOT)} ({len(data.get('paths', {}))} paths)")
    return 0


def main():
    ap = argparse.ArgumentParser(description=__doc__, formatter_class=argparse.RawDescriptionHelpFormatter)
    sub = ap.add_subparsers(dest="cmd")
    sub.add_parser("generate")
    sub.add_parser("check")
    rp = sub.add_parser("refresh-python-ai")
    rp.add_argument("--url", default="http://localhost:8000/openapi.json")
    args = ap.parse_args()
    if args.cmd == "check":
        return cmd_check(args)
    if args.cmd == "refresh-python-ai":
        return cmd_refresh_python_ai(args)
    return cmd_generate(args)


if __name__ == "__main__":
    sys.exit(main())
