import json
import urllib.request

url = "http://localhost:33550/api/settings"
data = json.dumps({"download_path": "/music"}).encode("utf-8")
req = urllib.request.Request(url, data=data, headers={"Content-Type": "application/json"}, method="POST")
try:
    with urllib.request.urlopen(req) as resp:
        print(resp.read().decode("utf-8"))
except Exception as e:
    print(f"Error: {e}")
