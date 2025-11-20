import requests
try:
    r = requests.get('http://localhost:8000/playlists/public')
    print(f"Status: {r.status_code}")
    print(f"Response: {r.text}")
except Exception as e:
    print(f"Error: {e}")
