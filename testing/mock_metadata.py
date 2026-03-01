from http.server import BaseHTTPRequestHandler, HTTPServer
import sys

class MetadataHandler(BaseHTTPRequestHandler):
    def do_GET(self):
        # Verify Metadata-Flavor header is present
        flavor = self.headers.get('Metadata-Flavor')
        if flavor != 'Google':
            self.send_error(403, "Forbidden: Missing Metadata-Flavor header")
            return

        # Check for the specific identity path
        # The tool requests: /computeMetadata/v1/instance/service-accounts/default/identity?audience=...&format=full
        if self.path.startswith("/computeMetadata/v1/instance/service-accounts/default/identity"):
            self.send_response(200)
            self.send_header('Content-type', 'text/plain')
            self.end_headers()
            self.wfile.write(b"mock-oidc-token-for-testing")
            return
        
        self.send_error(404, "Not Found")

if __name__ == "__main__":
    # Listen on all interfaces on port 80
    server = HTTPServer(('0.0.0.0', 80), MetadataHandler)
    print("Mock Metadata Server listening on port 80...", file=sys.stderr)
    server.serve_forever()
