import importlib.util
from pathlib import Path
import unittest

spec = importlib.util.spec_from_file_location(
    "split_nginx", Path(__file__).parents[1] / "split-nginx.py"
)
module = importlib.util.module_from_spec(spec)
spec.loader.exec_module(module)


class SplitNginxTests(unittest.TestCase):
    def test_preserves_blog_and_ip_routes_while_detaching_api_hosts(self):
        shared = """server {
    listen 80;
    server_name henryweb.site www.henryweb.site api.henryweb.site knowledge.henryweb.site;
    location / { return 308 https://$host$request_uri; }
}

"""
        blog = """server {
    listen 443 ssl;
    server_name henryweb.site;
    location / { proxy_pass http://127.0.0.1:3100; }
}

"""
        ip = """server {
    listen 443 ssl;
    server_name 101.34.211.31;
    location /graph/ {
        proxy_pass http://127.0.0.1:8181/;
        proxy_set_header X-Forwarded-Prefix /graph/;
    }
}

"""
        api = """server {
    listen 443 ssl;
    server_name api.henryweb.site;
    location / { proxy_pass http://127.0.0.1:8180; }
}

"""
        result = module.detach(shared + blog + api + ip)
        self.assertIn(blog, result)
        self.assertIn(ip, result)
        self.assertNotIn("api.henryweb.site", result)
        self.assertNotIn("knowledge.henryweb.site", result)
        self.assertIn("server_name henryweb.site www.henryweb.site;", result)
        self.assertEqual(module.detach(result), result)

    def test_unrecognized_config_fails_before_rewrite(self):
        with self.assertRaises(ValueError):
            module.detach("server { server_name knowledge.henryweb.site; }\n")


if __name__ == "__main__":
    unittest.main()
