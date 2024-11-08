# aws dynamodb put-item \
#     --table-name ShortCodes \
#     --item '{
#         "Shortcode": {"S": "example"},
#         "SortKey": {"S": "META"},
#         "Description" : {"S": "Example URL"},
#         "URL": {"S": "https://example.com"}
#     }'

curl -X POST https://o-sp.one/generate \
-H "x-api-key: ltZePsfaHM7ZdAX2C84Zk8Bu96OHJwnN26F9nYiO" \
-H "Content-Type: application/json" \
-d '{
    "url": "https://example.com",
    "description": "Example URL"
}'