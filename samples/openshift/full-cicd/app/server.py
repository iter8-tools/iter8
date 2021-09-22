#!/usr/bin/env python3

import os
import http.server
import socketserver
from http import HTTPStatus

version = os.getenv('version')

class Handler(http.server.SimpleHTTPRequestHandler):
    def do_GET(self):
        self.send_response(HTTPStatus.OK)
        self.end_headers()
        resp = 'Hello my world!!! (new version: %s)\n' % (version)
        self.wfile.write(str.encode(resp))

httpd = socketserver.TCPServer(('0.0.0.0', 8000), Handler)
httpd.serve_forever()
