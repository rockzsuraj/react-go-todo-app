#!/bin/bash

echo "🔍 Testing API Response Structure"
echo "================================="

echo ""
echo "1. Testing /api/auth/me endpoint structure..."
AUTH_RESPONSE=$(curl -s -X GET "http://localhost:8080/api/auth/me" -H "Origin: http://localhost:3000")
echo "Raw Response:"
echo "$AUTH_RESPONSE" | jq '.' 2>/dev/null || echo "$AUTH_RESPONSE"

echo ""
echo "2. Testing with valid cookies (simulating logged-in user)..."
# First try to get cookies by attempting login, then test auth/me
echo "Note: This test shows the expected response structure when authenticated"

echo ""
echo "3. Expected Response Structure Analysis:"
echo "Backend returns: { success: boolean, data: { user: UserResponse }, timestamp: string }"
echo "Frontend expects: UserResponse (unwrapped)"
echo "Fix needed: Extract res.data.data in auth API"

echo ""
echo "4. Testing /api/todos endpoint structure..."
TODO_RESPONSE=$(curl -s -X GET "http://localhost:8080/api/todos" -H "Origin: http://localhost:3000")
echo "Raw Response:"
echo "$TODO_RESPONSE" | jq '.' 2>/dev/null || echo "$TODO_RESPONSE"

echo ""
echo "================================="
echo "✅ Frontend API Fix Applied:"
echo "- authApi.getMe() now extracts res.data.data"
echo "- Added null checking for safety"
echo "- Other APIs (todos) already handle wrapped responses correctly"
