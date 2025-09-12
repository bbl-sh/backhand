loginResponse=$(curl -s -X POST http://localhost:3000/login \
  -H "Content-Type: application/json" \
  -d '{"email": "bhushanbharat6958@gmail.com", "password": "1"}')

echo "$loginResponse"

challengeToken=$(echo "$loginResponse" | jq -r '.token')

challengeEmail="bhushanbharat6958@gmail.com"

# Execute challenge
response=$(curl -s -X POST http://localhost:3000/challenge01 \
  -H "Authorization: Bearer $challengeToken" \
  -H "X-User-Email: $challengeEmail" \
  -F "challengeId=1" \
  -F "challengeName=Sum" \
  -F "problemId=1" \
  -F "code=@solution.py")

echo "Challenge response:" "$response"
