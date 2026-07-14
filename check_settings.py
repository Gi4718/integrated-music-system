import json
import urllib.request

url = "http://localhost:33550/api/settings"
try:
    with urllib.request.urlopen(url) as resp:
        data = json.loads(resp.read().decode("utf-8"))
        s = data["settings"]
        print(f"ssl_redirect: {s['ssl_redirect']}")
        print(f"ssl_mode: {s['ssl_mode']}")
        print(f"download_path: {s['download_path']}")
except Exception as e:
    print(f"Error: {e}")
