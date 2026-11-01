"""Behavioral tests using real Markdown files and a temporary HTTP server."""

import io
import json
import tempfile
import threading
import unittest
import urllib.error
import urllib.request
import zipfile
from pathlib import Path
from http.server import ThreadingHTTPServer
from server import Problem, Vault, make_handler, SessionAuth


class WorkspaceTests(unittest.TestCase):
    def setUp(self):
        self.temporary = tempfile.TemporaryDirectory()
        self.vault = Vault(Path(self.temporary.name) / "vault")

    def tearDown(self):
        self.temporary.cleanup()

    def create(self, title, content="", kind="concept"):
        return self.vault.write({"title": title, "kind": kind, "content": content, "tags": ["测试"]})

    def test_persistent_markdown_and_bidirectional_links(self):
        project = self.create("订单项目", kind="project")
        incident = self.create("上传故障", "所属：[[订单项目|项目]]\n另见 [[订单项目#运行]]\n[[未整理知识]]", "incident")
        # A fresh process can recover all relationships from files alone.
        recovered = {n["title"]: n for n in Vault(self.vault.path).read()}
        self.assertEqual(recovered["上传故障"]["outgoing"], [project["id"]])
        self.assertEqual(recovered["订单项目"]["incoming"], [incident["id"]])
        self.assertEqual(recovered["上传故障"]["missing"], ["未整理知识"])
        self.assertIn('kind: "incident"', (self.vault.path / (incident["id"] + ".md")).read_text())

    def test_code_examples_do_not_create_false_links(self):
        self.create("实际目标")
        note = self.create("链接示例", "`[[行内示例]]`\n\n```markdown\n[[代码示例]]\n```\n\n[[实际目标]]")
        self.assertEqual(note["links"], ["实际目标"])
        self.assertEqual(note["missing"], [])

    def test_external_edit_conflict_does_not_lose_either_version(self):
        original = self.create("并发修改", "初始正文")
        path = self.vault.path / (original["id"] + ".md")
        path.write_text(path.read_text().replace("初始正文", "外部编辑器修改"))
        with self.assertRaises(Problem) as result:
            self.vault.write({**original, "content": "浏览器修改"}, original["id"])
        self.assertEqual(result.exception.status, 409)
        self.assertIn("外部编辑器修改", path.read_text())

    def test_remove_is_recoverable_and_references_remain(self):
        target = self.create("可恢复笔记", "重要内容")
        self.create("引用者", "[[可恢复笔记]]")
        self.vault.remove(target["id"], target["version"])
        self.assertEqual(self.vault.read()[0]["missing"], ["可恢复笔记"])
        archived = list((self.vault.path / ".trash").glob("*.md"))
        self.assertEqual(len(archived), 1)
        self.assertIn("重要内容", archived[0].read_text())
        destination = self.vault.path / (target["id"] + ".md")
        archived[0].rename(destination)
        recovered = {n["title"]: n for n in self.vault.read()}
        self.assertEqual(recovered["引用者"]["outgoing"], [target["id"]])

    def test_ambiguous_titles_require_a_path(self):
        for folder in ["a", "b"]:
            directory = self.vault.path / folder
            directory.mkdir()
            (directory / "同名.md").write_text("# 文档")
        note = self.create("关系", "[[同名]] [[a/同名]]")
        self.assertEqual(note["ambiguous"], ["同名"])
        self.assertEqual(note["outgoing"], ["a/同名"])

    def test_no_overwrite_on_duplicate_or_invalid_filename(self):
        self.create("保留", "原文")
        for title in ["保留", "../逃逸", "路径/逃逸", "无效#标题"]:
            with self.assertRaises(Problem):
                self.create(title, "覆盖")
        self.assertEqual(self.vault.read()[0]["content"].strip(), "原文")

    def test_symlinks_do_not_escape_the_vault(self):
        outside = Path(self.temporary.name) / "private.md"
        outside.write_text("秘密")
        (self.vault.path / "泄露.md").symlink_to(outside)
        (self.vault.path / "01-项目").symlink_to(Path(self.temporary.name), target_is_directory=True)
        self.assertEqual(self.vault.read(), [])
        with self.assertRaises(Problem):
            self.create("写入", kind="project")
        self.assertFalse((Path(self.temporary.name) / "写入.md").exists())
        archive = zipfile.ZipFile(io.BytesIO(self.vault.export()))
        self.assertFalse(any("泄露" in name for name in archive.namelist()))

    def test_backup_contains_notes_and_recoverable_content(self):
        keep = self.create("保留文档", "保留")
        remove = self.create("回收文档", "可恢复")
        self.vault.remove(remove["id"], remove["version"])
        archive = zipfile.ZipFile(io.BytesIO(self.vault.export()))
        self.assertIn("vault/" + keep["id"] + ".md", archive.namelist())
        self.assertTrue(any("vault/.trash/" in name for name in archive.namelist()))
        self.assertTrue(any(name.startswith("templates/") for name in archive.namelist()))

    def test_invalid_external_metadata_does_not_break_the_library(self):
        (self.vault.path / "外部笔记.md").write_text('---\nkind: ["unexpected"]\ncategory: ["unexpected"]\ntags: null\n---\n\n内容')
        note = self.vault.read()[0]
        self.assertEqual(note["kind"], "concept")
        self.assertEqual(note["category"], "")
        self.assertEqual(note["tags"], [])

    def test_category_survives_restart_and_legacy_edits_without_breaking_links(self):
        target = self.create("函数传参")
        referring = self.create("学习记录", "[[函数传参]]")
        categorized = self.vault.write({**target, "category": " 前端 "}, target["id"])
        self.assertEqual(categorized["id"], target["id"])
        recovered = next(n for n in Vault(self.vault.path).read() if n["id"] == target["id"])
        self.assertEqual(recovered["category"], "前端")
        self.assertEqual(recovered["incoming"], [referring["id"]])
        legacy_payload = {key: recovered[key] for key in ["title", "kind", "tags", "content", "version"]}
        updated = self.vault.write({**legacy_payload, "content": "补充说明"}, recovered["id"])
        self.assertEqual(updated["category"], "前端")
        cleared = self.vault.write({**updated, "category": ""}, updated["id"])
        self.assertEqual(cleared["category"], "")

    def test_invalid_categories_do_not_create_notes(self):
        for category in [None, ["前端"], 42, "前端\n后端", "a" * 61]:
            with self.subTest(category=category), self.assertRaises(Problem):
                self.vault.write({"title": "分类验证", "category": category})
        self.assertEqual(self.vault.read(), [])

    def test_obsidian_tags_and_custom_metadata_survive_editing(self):
        path = self.vault.path / "外部资料.md"
        path.write_text('---\ntitle: 外部资料\nkind: concept\ntags:\n  - 后端\n  - 性能\ncustom:\n  owner: me\n  status: reviewed\n---\n\n原文')
        note = self.vault.read()[0]
        self.assertEqual(note["tags"], ["后端", "性能"])
        updated = self.vault.write({**note, "content": "已更新"}, note["id"])
        self.assertEqual(updated["tags"], ["后端", "性能"])
        self.assertIn('custom:\n  owner: me\n  status: reviewed', path.read_text())
        self.assertIn("已更新", path.read_text())


class HttpTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.temporary = tempfile.TemporaryDirectory()
        cls.vault = Vault(cls.temporary.name)
        cls.server = ThreadingHTTPServer(("127.0.0.1", 0), make_handler(cls.vault))
        cls.worker = threading.Thread(target=cls.server.serve_forever, daemon=True)
        cls.worker.start()
        cls.base = f"http://127.0.0.1:{cls.server.server_port}"

    @classmethod
    def tearDownClass(cls):
        cls.server.shutdown()
        cls.server.server_close()
        cls.worker.join()
        cls.temporary.cleanup()

    def request(self, path, method="GET", data=None, headers=None):
        request = urllib.request.Request(self.base + path, data=json.dumps(data).encode() if data is not None else None, method=method,
                                         headers={"Content-Type": "application/json", "X-DevGraph-Request": "1", **(headers or {})})
        try:
            with urllib.request.urlopen(request) as response:
                return response.status, response.headers, response.read()
        except urllib.error.HTTPError as error:
            return error.code, error.headers, error.read()

    def test_create_edit_and_refresh_end_to_end(self):
        status, _, body = self.request("/api/notes", "POST", {"title": "HTTP创建", "kind": "incident", "content": "第一版", "tags": []})
        self.assertEqual(status, 201)
        note = json.loads(body)
        path = "/api/notes/" + urllib.parse.quote(note["id"], safe="")
        status, _, body = self.request(path, "PUT", {**note, "content": "第二版"})
        self.assertEqual(status, 200)
        status, _, body = self.request("/api/notes")
        self.assertTrue(any(n["content"].strip() == "第二版" for n in json.loads(body)["notes"]))

    def test_import_reports_partial_failure_without_overwriting(self):
        status, _, body = self.request("/api/import", "POST", {"files": [{"name": "新笔记.md", "content": "正文"}, {"name": "新笔记.md", "content": "冲突正文"}]})
        self.assertEqual(status, 200)
        result = json.loads(body)
        self.assertEqual(len(result["imported"]), 1)
        self.assertEqual(len(result["errors"]), 1)
        self.assertEqual(next(n for n in self.vault.read() if n["title"] == "新笔记")["content"].strip(), "正文")

    def test_import_preserves_category(self):
        markdown = '---\ntitle: "前端导入笔记"\ncategory: "前端"\ntags: ["JavaScript"]\n---\n\n正文'
        status, _, body = self.request("/api/import", "POST", {"files": [{"name": "分类.md", "content": markdown}]})
        self.assertEqual(status, 200)
        self.assertEqual(json.loads(body)["errors"], [])
        note = next(n for n in self.vault.read() if n["title"] == "前端导入笔记")
        self.assertEqual(note["category"], "前端")
        self.assertEqual(note["tags"], ["JavaScript"])

    def test_local_boundary_and_static_allowlist(self):
        status, _, _ = self.request("/api/notes", "POST", {"title": "不应写入"}, {"Origin": "https://unrelated.example"})
        self.assertEqual(status, 403)
        self.assertEqual(self.request("/api/notes", headers={"Host": "unrelated.example"})[0], 403)
        self.assertEqual(self.request("/../server.py")[0], 404)
        self.assertEqual(self.request("/vault/01-项目/秘密.md".encode('ascii', 'ignore').decode())[0], 404)
        status, headers, body = self.request("/")
        self.assertEqual(status, 200)
        self.assertIn(b"DevGraph", body)
        self.assertIn("frame-ancestors 'none'", headers["Content-Security-Policy"])


class PublicOriginTests(unittest.TestCase):
    def test_reverse_proxy_origin_accepts_edits_and_rejects_foreign_origins(self):
        with tempfile.TemporaryDirectory() as directory:
            vault = Vault(directory)
            server = ThreadingHTTPServer(("127.0.0.1", 0), make_handler(vault, ["https://knowledge.example"]))
            worker = threading.Thread(target=server.serve_forever, daemon=True)
            worker.start()
            url = f"http://127.0.0.1:{server.server_port}/api/notes"
            try:
                def send(origin, title):
                    request = urllib.request.Request(url, data=json.dumps({"title": title, "kind": "concept", "tags": [], "content": "正文"}).encode(),
                                                     headers={"Host": "knowledge.example", "Origin": origin,
                                                              "Content-Type": "application/json", "X-DevGraph-Request": "1"})
                    return urllib.request.urlopen(request)
                with send("https://knowledge.example", "远程保存") as response:
                    self.assertEqual(response.status, 201)
                with self.assertRaises(urllib.error.HTTPError) as error:
                    send("https://unrelated.example", "不应保存")
                self.assertEqual(error.exception.code, 403)
                with urllib.request.urlopen(urllib.request.Request(url, headers={"Host": "knowledge.example"})) as response:
                    data = json.load(response)
                self.assertEqual(data["storage"], "server")
                self.assertEqual([note["title"] for note in data["notes"]], ["远程保存"])
            finally:
                server.shutdown()
                server.server_close()
                worker.join()


class SessionTests(unittest.TestCase):
    def setUp(self):
        self.temporary = tempfile.TemporaryDirectory()
        self.salt = "0123456789abcdef0123456789abcdef"
        self.password = "a-long-generated-test-password"
        self.auth = SessionAuth("devgraph", SessionAuth.digest(self.password, self.salt), self.salt, "/graph/")
        self.server = ThreadingHTTPServer(("127.0.0.1", 0), make_handler(Vault(self.temporary.name), ["https://knowledge.example"], self.auth))
        self.worker = threading.Thread(target=self.server.serve_forever, daemon=True)
        self.worker.start()
        self.base = f"http://127.0.0.1:{self.server.server_port}"

    def tearDown(self):
        self.server.shutdown()
        self.server.server_close()
        self.worker.join()
        self.temporary.cleanup()

    def request(self, path, data=None, cookie=""):
        headers = {"Host": "knowledge.example", "Origin": "https://knowledge.example", "X-DevGraph-Request": "1", "Cookie": cookie}
        request = urllib.request.Request(self.base + path, data=json.dumps(data).encode() if data is not None else None, headers=headers)
        try:
            with urllib.request.urlopen(request) as response:
                return response.status, response.headers, response.read()
        except urllib.error.HTTPError as error:
            return error.code, error.headers, error.read()

    def login(self):
        status, headers, _ = self.request("/api/login", {"username": "devgraph", "password": self.password})
        self.assertEqual(status, 200)
        return headers["Set-Cookie"]

    def test_login_page_does_not_expose_notes(self):
        self.assertEqual(self.request("/api/notes")[0], 401)
        status, _, body = self.request("/")
        self.assertEqual(status, 200)
        self.assertIn(b'login-form', body)
        self.assertEqual(self.request("/login.js")[0], 200)
        self.assertEqual(self.request("/api/export")[0], 401)

    def test_cookie_login_and_logout_protect_read_and_write(self):
        cookie = self.login()
        for attribute in ["HttpOnly", "Secure", "SameSite=Strict", "Path=/graph/"]:
            self.assertIn(attribute, cookie)
        token = cookie.split(";", 1)[0]
        self.assertEqual(self.request("/api/notes", cookie=token)[0], 200)
        status, _, _ = self.request("/api/notes", {"title": "登录后创建", "kind": "concept", "tags": [], "content": "保存"}, token)
        self.assertEqual(status, 201)
        self.assertEqual(self.request("/api/logout", {}, token)[0], 200)
        self.assertEqual(self.request("/api/notes", cookie=token)[0], 401)

    def test_invalid_sessions_and_expired_sessions_are_rejected(self):
        self.assertEqual(self.request("/api/notes", cookie="devgraph_session=forged")[0], 401)
        self.auth.ttl = -1
        cookie = self.login().split(";", 1)[0]
        self.assertEqual(self.request("/api/notes", cookie=cookie)[0], 401)

    def test_repeated_failed_passwords_are_limited(self):
        for _ in range(6):
            self.assertEqual(self.request("/api/login", {"username": "devgraph", "password": "wrong"})[0], 401)
        self.assertEqual(self.request("/api/login", {"username": "devgraph", "password": self.password})[0], 429)


if __name__ == "__main__":
    unittest.main()
