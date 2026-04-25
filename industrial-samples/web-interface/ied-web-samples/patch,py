import urllib.request
import os

with open("configurations/services/http-80.yaml", "r") as f:
    lines = f.readlines()

with open("industrial-samples/web-interface/ied-web-samples/login.html", "r") as f:
    login_html = f.read()

with open("industrial-samples/web-interface/ied-web-samples/dashboard.html", "r") as f:
    dashboard_html = f.read()

import yaml
with open("configurations/services/http-80.yaml", "r") as f:
    doc = yaml.safe_load(f)

# we will just use pyyaml and write it out, comments will be lost but that's ok
for cmd in doc.get("commands", []):
    if cmd.get("regex") == "^(/login|/login.html)$" and cmd.get("method") == "GET":
        cmd["handler"] = login_html
    if cmd.get("regex") == "^(/dashboard|/dashboard/)$" and cmd.get("method") == "GET":
        cmd["handler"] = dashboard_html

# Remove the POST /login
new_cmds = []
for cmd in doc.get("commands", []):
    if cmd.get("regex") == "^(/login|/login.html)$" and cmd.get("method") == "POST":
        continue
    new_cmds.append(cmd)
doc["commands"] = new_cmds

with open("configurations/services/http-80.yaml", "w") as f:
    yaml.dump(doc, f, allow_unicode=True, default_flow_style=False, sort_keys=False)

print("patched http-80.yaml successfully")
