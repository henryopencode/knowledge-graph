#!/usr/bin/env python3
"""A dependency-free, Markdown-backed knowledge workspace. Python 3.10+."""

import argparse
import csv
import hashlib
import hmac
import io
import json
import mimetypes
import os
import re
import secrets
import shutil
import tempfile
import threading
import time
import zipfile
from datetime import datetime, timezone
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
from http.cookies import SimpleCookie, CookieError
from pathlib import Path
from urllib.parse import unquote, urlsplit

ROOT = Path(__file__).resolve().parent
KINDS = {"project": "01-项目", "incident": "02-故障", "decision": "03-决策", "concept": "04-知识"}
TEMPLATES = {
    "project": "## 项目目标\n\n这个项目解决什么问题？\n\n## 模块与边界\n\n主要模块、职责和外部依赖。\n\n## 本地运行\n\n记录启动命令、配置位置和验证方式，勿记录密钥。\n\n## 关键约定\n\n业务规则、接口约定和相关决策。\n\n## 相关资料\n\n用 [[笔记标题]] 连接故障、决策或知识。\n",
    "incident": "## 现象与环境\n\n错误信息、触发条件、影响范围、版本。\n\n## 排查过程\n\n记录证据，以及排除过的假设。\n\n## 根因\n\n证据指向什么原因？尚未确认的内容请注明。\n\n## 修复方案\n\n具体修改与回滚方式。\n\n## 验证结果\n\n复现步骤、测试结果和适用范围。\n\n## 来源与关联\n\nIssue、修复提交、脱敏日志；用 [[笔记标题]] 关联项目和知识。\n",
    "decision": "## 背景与约束\n\n要解决什么问题？有哪些已知约束？\n\n## 候选方案\n\n比较收益、成本、复杂度。\n\n## 最终选择\n\n选择与理由。\n\n## 代价与复审条件\n\n接受了什么代价？何时需要重新评估？\n\n## 来源与关联\n\n讨论记录、实验、[[关联项目]]。\n",
    "concept": "## 我的理解\n\n用自己的话解释这个概念。\n\n## 适用场景\n\n什么时候使用，什么时候需要谨慎？\n\n## 实验与例子\n\n记录可验证的例子。\n\n## 在项目中的应用\n\n连接真实项目、故障或决策：[[笔记标题]]。\n\n## 参考资料\n\n官方文档与适用版本。\n",
}


class Problem(Exception):
    def __init__(self, status, message):
        self.status, self.message = status, message


def stamp():
    return datetime.now(timezone.utc).isoformat(timespec="seconds")


def clean_name(value):
    if not isinstance(value, str) or not value.strip():
        raise Problem(400, "请填写笔记标题。")
    value = value.strip()
    if len(value) > 100 or re.search(r'[\x00-\x1f<>:"/\\|?*#\[\]^]', value) or value in {".", ".."} or value.endswith("."):
        raise Problem(400, "标题过长或包含文件名、链接不支持的字符。")
    return value


def split_markdown(raw):
    raw = raw.lstrip("\ufeff").replace("\r\n", "\n")
    metadata = {}
    if raw.startswith("---\n"):
        end = raw.find("\n---\n", 4)
        if end >= 0:
            lines = raw[4:end].splitlines()
            for index, line in enumerate(lines):
                if not re.match(r"^[A-Za-z_][\w-]*:", line):
                    continue
                key, sep, value = line.partition(":")
                if sep:
                    value = value.strip()
                    try:
                        metadata[key.strip()] = json.loads(value)
                    except (ValueError, TypeError):
                        if value.startswith("[") and value.endswith("]"):
                            metadata[key.strip()] = [x.strip().strip("'\"") for x in next(csv.reader([value[1:-1]], skipinitialspace=True))]
                        elif not value and index + 1 < len(lines) and re.match(r"^\s*-\s", lines[index + 1]):
                            items = []
                            for following in lines[index + 1:]:
                                item = re.match(r"^\s*-\s+(.*)$", following)
                                if not item:
                                    break
                                items.append(item.group(1).strip().strip("'\""))
                            metadata[key.strip()] = items
                        else:
                            metadata[key.strip()] = value.strip("'\"")
            return metadata, raw[end + 5:].lstrip("\n")
    return metadata, raw


def links_in(body):
    # Code examples are not actual document links.
    text = re.sub(r"(?ms)^\s*(`{3,}|~{3,})[^\n]*\n.*?^\s*\1\s*$", "", body)
    text = re.sub(r"`+[^`\n]*`+", "", text)
    return list(dict.fromkeys(m.group(1).split("|", 1)[0].split("#", 1)[0].removesuffix(".md").strip()
                              for m in re.finditer(r"(?<!!)\[\[([^\]\n]+)\]\]", text)
                              if m.group(1).split("#", 1)[0].strip()))


class Vault:
    def __init__(self, path):
        self.path = Path(path).resolve()
        self.path.mkdir(parents=True, exist_ok=True)
        self.lock = threading.RLock()

    def safe(self, path):
        resolved = path.resolve()
        if not resolved.is_relative_to(self.path) or path.is_symlink():
            raise Problem(400, "文件路径超出知识库。")
        return resolved

    def read(self):
        with self.lock:
            notes = []
            for path in sorted(self.path.rglob("*.md")):
                relative = path.relative_to(self.path)
                if any(part.startswith(".") for part in relative.parts) or not path.is_file():
                    continue
                try:
                    self.safe(path)
                    data = path.read_bytes()
                    meta, body = split_markdown(data.decode("utf-8"))
                except (UnicodeDecodeError, OSError, Problem):
                    continue
                kind = meta.get("kind", "concept")
                if not isinstance(kind, str) or kind not in KINDS:
                    kind = "concept"
                category = meta.get("category", "")
                category = category.strip() if isinstance(category, str) else ""
                tags = meta.get("tags", [])
                if isinstance(tags, str):
                    tags = [x.strip() for x in tags.split(",") if x.strip()]
                if not isinstance(tags, list):
                    tags = []
                title = str(meta.get("title") or path.stem)
                modified = datetime.fromtimestamp(path.stat().st_mtime, timezone.utc).isoformat(timespec="seconds")
                notes.append({"id": relative.as_posix()[:-3], "title": title, "kind": kind, "category": category,
                              "tags": [str(x) for x in tags], "created": str(meta.get("created", modified)),
                              "updated": str(meta.get("updated", modified)), "content": body,
                              "sample": meta.get("sample") is True,
                              "version": hashlib.sha256(data).hexdigest(), "links": links_in(body),
                              "outgoing": [], "incoming": [], "missing": [], "ambiguous": []})
            exact = {note["id"]: note for note in notes}
            aliases = {}
            for note in notes:
                for alias in {note["title"], Path(note["id"]).name}:
                    aliases.setdefault(alias, []).append(note)
            for note in notes:
                for target in note["links"]:
                    found = exact.get(target)
                    if not found:
                        candidates = aliases.get(target, [])
                        if len(candidates) > 1:
                            note["ambiguous"].append(target)
                            continue
                        found = candidates[0] if candidates else None
                    if not found:
                        note["missing"].append(target)
                    elif found["id"] != note["id"] and found["id"] not in note["outgoing"]:
                        note["outgoing"].append(found["id"])
                        found["incoming"].append(note["id"])
            return notes

    def validate(self, payload, existing=None):
        if not isinstance(payload, dict):
            raise Problem(400, "请求内容必须是对象。")
        title = clean_name(payload.get("title"))
        if any(n["title"].casefold() == title.casefold() and (not existing or n["id"] != existing["id"]) for n in self.read()):
            raise Problem(409, "已有同名笔记，请使用不同标题。")
        kind = payload.get("kind", "concept")
        if not isinstance(kind, str) or kind not in KINDS:
            raise Problem(400, "未知笔记类型。")
        category = payload.get("category", existing.get("category", "") if existing else "")
        if not isinstance(category, str) or len(category.strip()) > 60 or re.search(r"[\x00-\x1f]", category):
            raise Problem(400, "分类必须是 60 字以内的单行文本。")
        content = payload.get("content", "")
        if not isinstance(content, str) or len(content.encode("utf-8")) > 512_000:
            raise Problem(400, "笔记正文必须是文本，且不能超过 500 KB。")
        tags = payload.get("tags", [])
        if not isinstance(tags, list) or len(tags) > 30 or any(not isinstance(t, str) or len(t) > 60 for t in tags):
            raise Problem(400, "标签格式不正确。")
        return {"title": title, "kind": kind, "category": category.strip(), "tags": list(dict.fromkeys(t.strip() for t in tags if t.strip())),
                "content": content.replace("\r\n", "\n")}

    def write(self, payload, note_id=None):
        with self.lock:
            existing = None
            notes = self.read()
            if note_id:
                existing = next((n for n in notes if n["id"] == note_id), None)
                if not existing:
                    raise Problem(404, "笔记不存在，可能已被移动。")
                if payload.get("version") != existing["version"]:
                    raise Problem(409, "文件已被其他窗口或编辑器修改。请先复制你的修改，再重新加载。")
                if payload.get("title", "").strip() != existing["title"]:
                    raise Problem(400, "为保留已有链接，请保持笔记标题不变；重命名可在 Obsidian 中完成。")
            values = self.validate(payload, existing)
            path = self.path / (existing["id"] + ".md") if existing else self.path / KINDS[values["kind"]] / (values["title"] + ".md")
            self.safe(path)
            if not existing and path.exists():
                raise Problem(409, "目标文件已存在。")
            path.parent.mkdir(parents=True, exist_ok=True)
            metadata = {"title": values["title"], "kind": values["kind"], "category": values["category"], "tags": values["tags"],
                        "created": existing["created"] if existing else stamp(), "updated": stamp()}
            if existing and existing["sample"]:
                metadata["sample"] = True
            # Preserve custom frontmatter verbatim, including nested YAML which
            # this app does not interpret. Only our managed fields are updated.
            preserved = []
            if existing:
                original = path.read_text(encoding="utf-8").lstrip("\ufeff").replace("\r\n", "\n")
                end = original.find("\n---\n", 4) if original.startswith("---\n") else -1
                skip = False
                if end >= 0:
                    for line in original[4:end].splitlines():
                        key = re.match(r"^([A-Za-z_][\w-]*):", line)
                        if key:
                            skip = key.group(1) in metadata
                        if not skip:
                            preserved.append(line)
            fields = [f"{key}: {json.dumps(value, ensure_ascii=False)}" for key, value in metadata.items()]
            text = "---\n" + "\n".join(fields + preserved) + "\n---\n\n" + values["content"].rstrip() + "\n"
            handle, temporary = tempfile.mkstemp(prefix=".saving-", dir=path.parent)
            try:
                with os.fdopen(handle, "w", encoding="utf-8") as stream:
                    stream.write(text)
                    stream.flush()
                    os.fsync(stream.fileno())
                os.replace(temporary, path)
            finally:
                if os.path.exists(temporary):
                    os.unlink(temporary)
            return next(n for n in self.read() if n["id"] == path.relative_to(self.path).as_posix()[:-3])

    def remove(self, note_id, version):
        with self.lock:
            note = next((n for n in self.read() if n["id"] == note_id), None)
            if not note:
                raise Problem(404, "笔记不存在。")
            if note["version"] != version:
                raise Problem(409, "文件已被修改，请重新加载后再移除。")
            path = self.safe(self.path / (note_id + ".md"))
            trash = self.safe(self.path / ".trash")
            trash.mkdir(exist_ok=True)
            destination = trash / f"{datetime.now().strftime('%Y%m%d-%H%M%S-%f')}-{path.name}"
            shutil.move(str(path), str(destination))

    def export(self):
        with self.lock:
            output = io.BytesIO()
            with zipfile.ZipFile(output, "w", zipfile.ZIP_DEFLATED) as archive:
                for path in sorted(self.path.rglob("*.md")):
                    if path.is_file():
                        try:
                            self.safe(path)
                        except Problem:
                            continue
                        archive.write(path, "vault/" + path.relative_to(self.path).as_posix())
                for path in sorted((ROOT / "templates").glob("*.md")):
                    archive.write(path, "templates/" + path.name)
            return output.getvalue()


class SessionAuth:
    """Single-owner login; random, revocable sessions stay in server memory."""

    def __init__(self, username, password_hash, salt, cookie_path="/", ttl=43200):
        self.username, self.password_hash, self.salt = username, password_hash, salt
        self.cookie_path, self.ttl = cookie_path, ttl
        self.sessions, self.failures = {}, {}
        self.lock = threading.Lock()

    @staticmethod
    def digest(password, salt):
        return hashlib.pbkdf2_hmac("sha256", password.encode(), bytes.fromhex(salt), 260000).hex()

    def login(self, username, password, address):
        if not isinstance(username, str) or not isinstance(password, str) or len(password) > 1024:
            raise Problem(400, "请输入用户名和密码。")
        now = time.time()
        with self.lock:
            self.failures = {key: value for key, value in self.failures.items() if value[1] > now - 900}
            if self.failures.get(address, (0, 0))[0] >= 6:
                raise Problem(429, "尝试次数过多，请 15 分钟后重试。")
        valid = hmac.compare_digest(self.digest(password, self.salt), self.password_hash)
        valid = hmac.compare_digest(username.encode(), self.username.encode()) and valid
        with self.lock:
            if not valid:
                count, _ = self.failures.get(address, (0, now))
                if len(self.failures) >= 1024:
                    self.failures.pop(next(iter(self.failures)))
                self.failures[address] = (count + 1, now)
                raise Problem(401, "用户名或密码不正确。")
            self.failures.pop(address, None)
            self.sessions = {key: expiry for key, expiry in self.sessions.items() if expiry > now}
            if len(self.sessions) >= 1024:
                self.sessions.pop(next(iter(self.sessions)))
            token = secrets.token_urlsafe(32)
            self.sessions[hashlib.sha256(token.encode()).hexdigest()] = now + self.ttl
        return token

    def session_key(self, cookie_header):
        try:
            cookie = SimpleCookie()
            cookie.load(cookie_header or "")
            token = cookie.get("devgraph_session")
            return hashlib.sha256(token.value.encode()).hexdigest() if token else ""
        except CookieError:
            return ""

    def authenticated(self, cookie_header):
        with self.lock:
            return self.sessions.get(self.session_key(cookie_header), 0) > time.time()

    def logout(self, cookie_header):
        with self.lock:
            self.sessions.pop(self.session_key(cookie_header), None)

    def cookie(self, token="", clear=False):
        return f"devgraph_session={token}; Path={self.cookie_path}; Max-Age={0 if clear else self.ttl}; HttpOnly; Secure; SameSite=Strict"


def make_handler(vault, public_origins=(), auth=None):
    public_origins = frozenset(public_origins)
    public_hosts = {urlsplit(origin).netloc for origin in public_origins}

    class Handler(BaseHTTPRequestHandler):
        def log_message(self, format, *args):
            if args and str(args[1] if len(args) > 1 else "").startswith("5"):
                super().log_message(format, *args)

        def send(self, status, data, mime="application/json; charset=utf-8", extra=None):
            if isinstance(data, (dict, list)):
                data = json.dumps(data, ensure_ascii=False).encode("utf-8")
            elif isinstance(data, str):
                data = data.encode("utf-8")
            self.send_response(status)
            self.send_header("Content-Type", mime)
            self.send_header("Content-Length", str(len(data)))
            self.send_header("Cache-Control", "no-store")
            self.send_header("X-Content-Type-Options", "nosniff")
            self.send_header("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self'; img-src 'self' data:; connect-src 'self'; frame-ancestors 'none'; base-uri 'self'; form-action 'self'")
            for key, value in (extra or {}).items():
                self.send_header(key, value)
            self.end_headers()
            self.wfile.write(data)

        def body(self):
            try:
                length = int(self.headers.get("Content-Length", "0"))
            except ValueError:
                raise Problem(400, "无效请求长度。")
            if not 0 < length <= 2_000_000:
                raise Problem(413, "请求为空或超过 2 MB，请分批导入。")
            try:
                return json.loads(self.rfile.read(length))
            except (ValueError, UnicodeDecodeError):
                raise Problem(400, "无效 JSON。")

        def dispatch(self):
            host = self.headers.get("Host", "")
            allowed_hosts = {f"127.0.0.1:{self.server.server_port}", f"localhost:{self.server.server_port}", *public_hosts}
            if host not in allowed_hosts:
                raise Problem(403, "请求的访问地址未配置。")
            if self.command != "GET":
                origin = self.headers.get("Origin")
                if origin and origin not in {f"http://{host}", *public_origins}:
                    raise Problem(403, "请求来源不匹配。")
                if self.headers.get("X-DevGraph-Request") != "1":
                    raise Problem(403, "缺少本地请求标识。")
            path = unquote(urlsplit(self.path).path)
            if auth:
                if self.command == "POST" and path == "/api/login":
                    payload = self.body()
                    if not isinstance(payload, dict):
                        raise Problem(400, "无效的登录请求。")
                    token = auth.login(payload.get("username"), payload.get("password"),
                                       self.headers.get("X-Real-IP") or self.client_address[0])
                    return self.send(200, {"ok": True}, extra={"Set-Cookie": auth.cookie(token)})
                if self.command == "POST" and path == "/api/logout":
                    auth.logout(self.headers.get("Cookie"))
                    return self.send(200, {"ok": True}, extra={"Set-Cookie": auth.cookie(clear=True)})
                login_assets = {"/login.css": "login.css", "/login.js": "login.js", "/favicon.svg": "favicon.svg"}
                if self.command == "GET" and path in login_assets:
                    file = ROOT / "web" / login_assets[path]
                    mime = mimetypes.guess_type(str(file))[0] or "application/octet-stream"
                    return self.send(200, file.read_bytes(), mime + "; charset=utf-8")
                if not auth.authenticated(self.headers.get("Cookie")):
                    if self.command == "GET" and path in {"/", "/index.html"}:
                        return self.send(200, (ROOT / "web/login.html").read_bytes(), "text/html; charset=utf-8")
                    raise Problem(401, "请先登录知识库。")
            if self.command == "GET" and path == "/api/health":
                return self.send(200, {"ok": True})
            if self.command == "GET" and path == "/api/notes":
                return self.send(200, {"notes": vault.read(), "vault": str(vault.path), "templates": TEMPLATES,
                                       "storage": "server" if public_origins else "local", "authenticated": bool(auth)})
            if self.command == "GET" and path == "/api/export":
                return self.send(200, vault.export(), "application/zip", {"Content-Disposition": 'attachment; filename="devgraph-backup.zip"'})
            if self.command == "POST" and path == "/api/notes":
                return self.send(201, vault.write(self.body()))
            if path.startswith("/api/notes/") and self.command in {"PUT", "DELETE"}:
                payload = self.body()
                if not isinstance(payload, dict):
                    raise Problem(400, "无效请求内容。")
                note_id = path[len("/api/notes/"):]
                if self.command == "PUT":
                    return self.send(200, vault.write(payload, note_id))
                vault.remove(note_id, payload.get("version"))
                return self.send(200, {"ok": True})
            if self.command == "POST" and path == "/api/import":
                payload = self.body()
                files = payload.get("files") if isinstance(payload, dict) else None
                if not isinstance(files, list) or not 1 <= len(files) <= 50:
                    raise Problem(400, "每次可导入 1–50 个 Markdown 文件。")
                imported, errors = [], []
                for item in files:
                    try:
                        if not isinstance(item, dict) or not isinstance(item.get("content"), str):
                            raise Problem(400, "无效文件内容。")
                        meta, body = split_markdown(item["content"])
                        title = meta.get("title") or Path(str(item.get("name", "导入笔记"))).stem
                        imported.append(vault.write({"title": title, "kind": meta.get("kind", "concept"), "category": meta.get("category", ""), "tags": meta.get("tags", []), "content": body})["id"])
                    except Problem as error:
                        errors.append({"name": str(item.get("name", "未知文件")) if isinstance(item, dict) else "未知文件", "error": error.message})
                return self.send(200, {"imported": imported, "errors": errors})
            if self.command != "GET":
                raise Problem(404, "接口不存在。")
            static = {"/": "index.html", "/index.html": "index.html", "/app.js": "app.js", "/styles.css": "styles.css", "/favicon.svg": "favicon.svg"}
            if path not in static:
                raise Problem(404, "页面不存在。")
            target = ROOT / "web" / static[path]
            mime = mimetypes.guess_type(str(target))[0] or "application/octet-stream"
            return self.send(200, target.read_bytes(), mime + ("; charset=utf-8" if mime.startswith("text/") or "javascript" in mime else ""))

        def handle_request(self):
            try:
                self.dispatch()
            except Problem as error:
                self.send(error.status, {"error": error.message})
            except (BrokenPipeError, ConnectionResetError):
                pass
            except Exception:
                import traceback
                traceback.print_exc()
                self.send(500, {"error": "操作未完成，请查看终端日志；原文件不会因写入中断而被截断。"})

        do_GET = do_POST = do_PUT = do_DELETE = handle_request

    return Handler


def main():
    parser = argparse.ArgumentParser(description="DevGraph — 本地程序员知识库")
    parser.add_argument("--port", type=int, default=8765)
    parser.add_argument("--vault", type=Path, default=ROOT / "vault")
    parser.add_argument("--public-origin", action="append", default=[],
                        help="允许的公网访问来源，例如 https://example.com；仅用于有访问控制的反向代理之后。")
    args = parser.parse_args()
    origins = []
    for origin in args.public_origin:
        parsed = urlsplit(origin)
        if parsed.scheme not in {"http", "https"} or not parsed.netloc or parsed.username or parsed.password or parsed.path not in {"", "/"} or parsed.query or parsed.fragment:
            parser.error("--public-origin 必须是完整的 HTTP(S) 来源，不能包含路径或凭据。")
        origins.append(origin.rstrip("/"))
    vault = Vault(args.vault)
    auth = None
    if os.environ.get("DEVGRAPH_AUTH_ENABLED") == "1":
        password_hash = os.environ.get("DEVGRAPH_PASSWORD_HASH", "")
        salt = os.environ.get("DEVGRAPH_PASSWORD_SALT", "")
        cookie_path = os.environ.get("DEVGRAPH_COOKIE_PATH", "/")
        if not re.fullmatch(r"[a-f0-9]{64}", password_hash) or not re.fullmatch(r"[a-f0-9]{32}", salt):
            parser.error("服务器登录配置缺失或无效。")
        if not cookie_path.startswith("/") or not cookie_path.endswith("/") or re.search(r"[;\s]", cookie_path):
            parser.error("会话 Cookie 路径无效。")
        auth = SessionAuth(os.environ.get("DEVGRAPH_USERNAME", "devgraph"), password_hash, salt, cookie_path)
    try:
        server = ThreadingHTTPServer(("127.0.0.1", args.port), make_handler(vault, origins, auth))
    except OSError as error:
        parser.exit(1, f"启动失败：{error}。可用 --port 8766 更换端口。\n")
    print(f"DevGraph 已启动：http://127.0.0.1:{server.server_port}\n知识库：{vault.path}\n按 Ctrl+C 停止。", flush=True)
    try:
        server.serve_forever()
    except KeyboardInterrupt:
        print("\nDevGraph 已停止。")
    finally:
        server.server_close()


if __name__ == "__main__":
    main()
