aws dynamodb put-item \
    --table-name ShortCodes \
    --item '{
        "Shortcode": {"S": "example"},
        "SortKey": {"S": "META"},
        "Description" : {"S": "Example URL"},
        "URL": {"S": "https://example.com"}
    }'


aws dynamodb put-item \
    --table-name ShortCodes \
    --item '{
        "Shortcode": {"S": "example"},
        "SortKey": {"S": "META"},
        "Description" : {"S": "Example URL"},
        "URL": {"S": "https://example.com"}
    }'