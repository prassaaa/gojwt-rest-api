#!/bin/bash

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

BASE_URL="http://localhost:8080/api/v1"

echo -e "${YELLOW}=== Testing Refresh Token & Logout Implementation ===${NC}\n"

# Step 1: Register a test user
echo -e "${YELLOW}1. Registering test user...${NC}"
REGISTER_RESPONSE=$(curl -s -X POST "$BASE_URL/auth/register" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Refresh Test User",
    "email": "refresh.test@example.com",
    "password": "password123"
  }')
echo "$REGISTER_RESPONSE" | jq '.'

# Step 2: Login to get tokens
echo -e "\n${YELLOW}2. Logging in to get tokens...${NC}"
LOGIN_RESPONSE=$(curl -s -X POST "$BASE_URL/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "refresh.test@example.com",
    "password": "password123"
  }')
echo "$LOGIN_RESPONSE" | jq '.'

ACCESS_TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.data.access_token')
REFRESH_TOKEN=$(echo "$LOGIN_RESPONSE" | jq -r '.data.refresh_token')

echo -e "\n${GREEN}Access Token: ${ACCESS_TOKEN:0:50}...${NC}"
echo -e "${GREEN}Refresh Token: ${REFRESH_TOKEN:0:50}...${NC}"

# Step 3: Use access token to access protected endpoint
echo -e "\n${YELLOW}3. Testing access token on protected endpoint...${NC}"
curl -s -X GET "$BASE_URL/profile" \
  -H "Authorization: Bearer $ACCESS_TOKEN" | jq '.'

# Step 4: Wait a bit and refresh the token
echo -e "\n${YELLOW}4. Refreshing access token...${NC}"
sleep 2
REFRESH_RESPONSE=$(curl -s -X POST "$BASE_URL/auth/refresh" \
  -H "Content-Type: application/json" \
  -d "{
    \"refresh_token\": \"$REFRESH_TOKEN\"
  }")
echo "$REFRESH_RESPONSE" | jq '.'

NEW_ACCESS_TOKEN=$(echo "$REFRESH_RESPONSE" | jq -r '.data.access_token')
NEW_REFRESH_TOKEN=$(echo "$REFRESH_RESPONSE" | jq -r '.data.refresh_token')

echo -e "\n${GREEN}New Access Token: ${NEW_ACCESS_TOKEN:0:50}...${NC}"
echo -e "${GREEN}New Refresh Token: ${NEW_REFRESH_TOKEN:0:50}...${NC}"

# Step 5: Try to use old refresh token (should fail - token rotation)
echo -e "\n${YELLOW}5. Testing old refresh token (should fail - token rotation)...${NC}"
curl -s -X POST "$BASE_URL/auth/refresh" \
  -H "Content-Type: application/json" \
  -d "{
    \"refresh_token\": \"$REFRESH_TOKEN\"
  }" | jq '.'

# Step 6: Use new access token on protected endpoint
echo -e "\n${YELLOW}6. Testing new access token on protected endpoint...${NC}"
curl -s -X GET "$BASE_URL/profile" \
  -H "Authorization: Bearer $NEW_ACCESS_TOKEN" | jq '.'

# Step 7: Logout
echo -e "\n${YELLOW}7. Logging out...${NC}"
LOGOUT_RESPONSE=$(curl -s -X POST "$BASE_URL/auth/logout" \
  -H "Authorization: Bearer $NEW_ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"refresh_token\": \"$NEW_REFRESH_TOKEN\"
  }")
echo "$LOGOUT_RESPONSE" | jq '.'

# Step 8: Try to use refresh token after logout (should fail)
echo -e "\n${YELLOW}8. Testing refresh token after logout (should fail)...${NC}"
curl -s -X POST "$BASE_URL/auth/refresh" \
  -H "Content-Type: application/json" \
  -d "{
    \"refresh_token\": \"$NEW_REFRESH_TOKEN\"
  }" | jq '.'

echo -e "\n${GREEN}=== Test Complete ===${NC}"
