import yaml
import http.server
import socketserver
import threading
import time
import yaml
import json
import os
from urllib.parse import urlparse, parse_qs

os.environ['LIBROSA_CACHE_DIR'] = '/tmp/librosa_cache'
os.environ['NUMBA_CACHE_DIR'] = '/tmp/numba_cache'
from musicnn.tagger import top_tags

config_path = os.path.join(os.path.dirname(__file__), './config.yaml')
with open(config_path, 'r') as f:
    config = yaml.safe_load(f)

PORT = config['ports']['genre']
ENDPOINT = config['endpoints']['genre']

shutdown_flag = threading.Event()

class SimpleHTTPRequestHandler(http.server.SimpleHTTPRequestHandler):
    def do_GET(self):
        parsed_url = urlparse(self.path)
        query_params = parse_qs(parsed_url.query)
        path = parsed_url.path
        if path == "/genre":
            print("request initiated")
            file_id = query_params['id'][0]
            file_path = os.path.join('./data/', file_id + '.mp3')
            print("file path: " + file_path)
            if not os.path.exists(file_path):
                self.send_response(400)
                self.end_headers()
            print("file exists: " + file_path)
            tops = 5
            t = top_tags(file_path, model='MSD_musicnn', topN=tops, print_tags=False)
            v = top_tags(file_path, model='MSD_vgg', topN=tops, print_tags=False)
            x = top_tags(file_path, model='MTT_musicnn', topN=tops, print_tags=False)
            y = top_tags(file_path, model='MTT_vgg', topN=tops, print_tags=False)
            print(t)
            print(v)
            print(x)
            print(y)
            z = {}
            w = tops
            for l in t:
                z[l] = w
                w -= 1
            w = tops
            for l in v:
                if l in z: 
                    z[l] += w
                else:
                    z[l] = w
                w -= 1
            w = tops
            for l in x:
                if l in z: 
                    z[l] += w
                else:
                    z[l] = w
                w -= 1
            w = tops
            for l in y:
                if l in z: 
                    z[l] += w
                else:
                    z[l] = w
                w -= 1
            s = sorted(z.items(), key=lambda x: x[1], reverse=True)
            z = dict(s)
            print(z)
            g = ['classical', 'techno', 'strings', 'drums', 'electronic', 'rock', 'piano', 'ambient', 'violin', 'vocal', 'synth', 'indian', 'opera', 'harpsichord', 'flute', 'pop', 'sitar', 'classic', 'choir', 'new age', 'dance', 'harp', 'cello', 'country', 'metal', 'choral', 'alternative', 'indie', '00s', 'alternative rock', 'jazz', 'chillout', 'classic rock', 'soul', 'indie rock', 'Mellow', 'electronica', '80s', 'folk', '90s', 'chill', 'instrumental', 'punk', 'oldies', 'blues', 'hard rock', 'acoustic', 'experimental', 'Hip-Hop', '70s', 'party', 'easy listening', 'funk', 'electro', 'heavy metal', 'Progressive rock', '60s', 'rnb', 'indie pop', 'sad', 'House']
            genre = next((k for k in z if k in g), list(z.keys())[0])
            print(genre.title())
            response = {'genre': genre.title()}
            self.send_response(200)
            self.end_headers()
            self.wfile.write(json.dumps(response).encode('utf-8'))
        elif path == '/kill':
            self.send_response(200)
            self.end_headers()
            self.wfile.write(b'Server is shutting down...')
            shutdown_flag.set()
        else:
            self.send_response(404)
            self.end_headers()
            self.wfile.write(b'Not Found')

def run_server():
    with socketserver.TCPServer(("0.0.0.0", PORT), SimpleHTTPRequestHandler) as httpd:
        print(f"Serving on port {PORT}")
        while not shutdown_flag.is_set():
            httpd.handle_request()

# Run the server in a separate thread
server_thread = threading.Thread(target=run_server)
server_thread.start()

# Wait for the shutdown signal
try:
    while not shutdown_flag.is_set():
        time.sleep(1)
finally:
    print("Server has been shut down.")