#!/bin/bash

# Use 'git config' to read values from the repository's configuration.

challengeEmail=$(git config user.challengeEmail)
challengeToken=$(git config user.challengeToken)
echo "$challengeEmail"
echo "$challengeToken"

# Check if the configuration values were found.
if [[ -z "$challengeEmail" || -z "$challengeToken" ]]; then
  echo "Error: 'challenge.email' or 'challenge.token' not found in your git config."
  echo "Please set them for this repository by running:"
  echo "git config challenge.email \"your-email@example.com\""
  echo "git config challenge.token \"your-secret-token\""
  exit 1
fi

echo "Configuration found. Sending verification request..."

# Send curl request to the test endpoint.
curl -X GET http://13.232.198.0:8080/testConnection \
-H "Authorization: Bearer $challengeToken" \
-H "X-User-Email: $challengeEmail"

# Check the exit status of curl to see if it failed
if [[ $? -ne 0 ]]; then
  echo "Error: The curl request failed. Please check if the server is running."
  exit 1
fi

echo -e "\nVerification request sent successfully."
