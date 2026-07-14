import json
import urllib.request

url = "http://localhost:33550/api/settings"
try:
    with urllib.request.urlopen(url) as resp:
        data = json.loads(resp.read().decode("utf-8"))
        print(json.dumps(data["settings"]["download_path"]))
except Exception as e:
    print(f"Error: {e}")
