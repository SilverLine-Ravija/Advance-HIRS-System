#!/usr/bin/env python3
"""Assert the production Compose template does not accidentally inherit demo exposure."""
import json
import sys

config = json.load(sys.stdin)
services = config["services"]
assert "postgres" not in services, "production template must use an external database"
assert {"migrate", "api", "web", "agent"}.issubset(services), "expected components missing"

for service in ("api", "agent", "migrate"):
    assert not services[service].get("ports"), f"{service} has published host ports"

ports = services["web"].get("ports") or []
assert ports, "web must have an explicit, loopback-only reverse-proxy binding"
assert all(p.get("host_ip") == "127.0.0.1" for p in ports), "web must be loopback-only"
assert services["api"]["environment"]["SEED_DEMO_DATA"] == "false", "demo seeding must be disabled"
assert services["web"]["environment"]["WEB_COOKIE_SECURE"] == "true", "secure web cookies required"
assert services["agent"]["environment"]["AGENT_DATABASE_URL"] != services["api"]["environment"]["DATABASE_URL"], "agent and API must use distinct DB roles"
assert not services["agent"].get("privileged"), "agent cannot be a privileged container"
assert services["api"].get("read_only"), "API filesystem must be read-only"
assert services["agent"].get("read_only"), "agent filesystem must be read-only"
print("PASS: standalone production Compose isolates database, backend ports, demo users and AI credentials.")
