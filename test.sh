#!/bin/bash

# === Config ===
BASE_URL="http://localhost:3000"
PB_URL="http://13.232.198.0:8890"

challengeEmail="bhushanbharat6958@gmail.com"
challengePassword="1"
challengeFile="./challenges/challenge.py"
problemId=1
challengeName="mastering-python"
challengeId="1"

# === Function to Login and Get Token ===
login() {
  loginResponse=$(curl -s -X POST "$PB_URL/api/collections/users/auth-with-password" \
    -H "Content-Type: application/json" \
    -d "{\"identity\":\"$challengeEmail\",\"password\":\"$challengePassword\"}")

  challengeToken=$(echo "$loginResponse" | jq -r '.token')

  if [[ "$challengeToken" == "null" || -z "$challengeToken" ]]; then
    echo "Login failed. Please check your email and password."
    echo "Response: $loginResponse"
    exit 1
  fi

  echo "Login successful. Token received."
}

# === Check if Token is Valid ===
check_token_validity() {
  response=$(curl -s -X POST "$BASE_URL/challenge01" \
    -H "Authorization: Bearer $challengeToken" \
    -H "X-User-Email: $challengeEmail" \
    -F "challengeId=$challengeId" \
    -F "challengeName=$challengeName" \
    -F "problemId=$problemId" \
    -F "code=@$challengeFile")

  if echo "$response" | grep -q "Authentication failed"; then
    echo "Token is invalid or expired. Reauthenticating..."
    login
  fi
}

# === Main Script ===
if [[ -z "$challengeToken" ]]; then
  login
else
  check_token_validity
fi

# === Send Challenge Submission ===
response=$(curl -s -X POST "$BASE_URL/challenge01" \
  -H "Authorization: Bearer $challengeToken" \
  -H "X-User-Email: $challengeEmail" \
  -F "challengeId=$challengeId" \
  -F "challengeName=$challengeName" \
  -F "problemId=$problemId" \
  -F "code=@$challengeFile")

echo "Challenge submission response:"
echo "$response"
