curl -X POST --location "http://localhost:8081" \
    -H "Accept: application/json" \
    -H "Content-Type: multipart/form-data" \
    -F meta='{
  "url": "/my-endpoint",
  "method": "POST"
}' \
    -F body='@lorem-ipsum.txt;filename=lorem-ipsum.txt'