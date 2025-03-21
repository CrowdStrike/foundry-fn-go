curl -X POST --location "http://localhost:8081" \
    -H "Content-Type: multipart/form-data" \
    -F meta='{
  "url": "/my-endpoint",
  "method": "POST"
}' \
    -F body1='@lorem-ipsum-1.txt;filename=lorem-ipsum-1.txt' \
    -F body2='@lorem-ipsum-2.txt;filename=lorem-ipsum-2.txt' \
    -F body3='@lorem-ipsum-3.txt;filename=lorem-ipsum-3.txt'