#!/bin/bash

# === Colors ===
RED='\033[1;31m'
GREEN='\033[1;32m'
YELLOW='\033[1;33m'
CYAN='\033[1;36m'
BOLD='\033[1m'
RESET='\033[0m'

# === Config ===
#challengeEmail = bhushanbharat6958@gmail.com
#challengeToken = eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJjb2xsZWN0aW9uSWQiOiJfcGJfdXNlcnNfYXV0aF8iLCJleHAiOjE3NTc0OTI1MzQsImlkIjoiaXFpNjJoYXIybnV3dGJwIiwicmVmcmVzaGFibGUiOnRydWUsInR5cGUiOiJhdXRoIn0.B2eQI-zLtCAMp000og-MviDuQ2gsbmJfcqruY58_l-Y
#
challengeEmail=$(git config user.challengeEmail)
challengeToken=$(git config user.challengeToken)
challengeFile="./challenges/challenge.py"
problemId=1   # Set the problem ID you want to test
learnName="mastering-python"
learnId="4"

# === Checks ===
if [[ -z "$challengeEmail" || -z "$challengeToken" ]]; then
  clear
  echo
  echo -e "${YELLOW}💡 No credentials found please login with:${RESET}"
  echo "   git commit -m \"setup\""
  exit 1
fi

if [[ ! -f "$challengeFile" ]]; then
  echo -e "${RED}❌ Error:${RESET} Challenge file not found: $challengeFile"
  exit 1
fi

# === Info ===
echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${RESET}"
echo -e "✅ ${GREEN}Configuration OK${RESET}"
echo -e "📧 User: ${YELLOW}$challengeEmail${RESET}"
echo -e "📌 Problem ID: ${YELLOW}$problemId${RESET}"
echo -e "📝 Code file: ${YELLOW}$challengeFile${RESET}"
echo -e "🚀 Sending submission to ${CYAN}http://*:8080/challenge01${RESET}..."
echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${RESET}"

# === Send request ===
response=$(curl -s  -X POST http://localhost:8080/challenge01 \
  -H "Authorization: Bearer $challengeToken" \
  -H "X-User-Email: $challengeEmail" \
  -F "learnId=$learnId" \
  -F "learnName=$learnName" \
  -F "problemId=$problemId" \
  -F "code=@$challengeFile")


if [[ $? -ne 0 || -z "$response" ]]; then
  echo -e "\n${RED}❌ Error:${RESET} Request failed."
  echo "💡 Check if server is running on :8080 and token is valid."
  exit 1
fi

# === Parse Response ===
status=$(echo "$response" | jq -r '.status')
message=$(echo "$response" | jq -r '.message')
actual_output=$(echo "$response" | jq -r '.actual_output')
expected_output=$(echo "$response" | jq -r '.expected_output')
user_email=$(echo "$response" | jq -r '.user_email')
timestamp=$(echo "$response" | jq -r '.timestamp')

# === Pretty Result ===
echo -e "\n${CYAN}━━━━━━━━━━ RESULT ━━━━━━━━━━${RESET}"
if [[ "$status" == "Accepted" ]]; then
  echo -e "🎉 ${GREEN}${BOLD}Challenge Completed Successfully!${RESET}"
  echo -e "👤 User: $user_email"
  echo -e "⏰ Time: $timestamp"
  echo -e "✅ Status: ${GREEN}$status${RESET}"
else
  echo -e "⚠️  ${RED}${BOLD}Challenge Not Accepted${RESET}"
  echo -e "👤 User: $user_email"
  echo -e "📝 Message: ${YELLOW}$message${RESET}"
  echo -e "🔵 Actual:   $actual_output"
  echo -e "🟢 Expected: $expected_output"
  echo -e "❌ Status: ${RED}$status${RESET}"
fi
echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${RESET}"
