#!/bin/bash

echo "🧪 Testing React Todos OAuth Flow"
echo "=================================="

echo ""
echo "1. 📱 Testing frontend accessibility..."
FRONTEND_STATUS=$(curl -s -o /dev/null -w "%{http_code}" "http://localhost:3000")
if [ "$FRONTEND_STATUS" = "200" ]; then
    echo "✅ Frontend accessible (HTTP $FRONTEND_STATUS)"
else
    echo "❌ Frontend not accessible (HTTP $FRONTEND_STATUS)"
fi

echo ""
echo "2. 🔧 Testing backend accessibility..."
BACKEND_STATUS=$(curl -s -o /dev/null -w "%{http_code}" "http://localhost:8080/health")
if [ "$BACKEND_STATUS" = "200" ]; then
    echo "✅ Backend accessible (HTTP $BACKEND_STATUS)"
else
    echo "❌ Backend not accessible (HTTP $BACKEND_STATUS)"
fi

echo ""
echo "3. 🔗 Testing OAuth login endpoint..."
LOGIN_URL="http://localhost:8080/api/auth/google/login?redirect=http%3A%2F%2Flocalhost%3A3000%2Foauth%2Fcallback"
LOGIN_RESPONSE=$(curl -s -I "$LOGIN_URL")
if echo "$LOGIN_RESPONSE" | grep -q "302\|Found"; then
    echo "✅ Login endpoint redirects correctly"
    GOOGLE_URL=$(echo "$LOGIN_RESPONSE" | grep -i location | cut -d' ' -f2)
    echo "   Redirects to: ${GOOGLE_URL:0:100}..."
else
    echo "❌ Login endpoint not redirecting"
fi

echo ""
echo "4. 🔄 Testing OAuth callback endpoint..."
CALLBACK_RESPONSE=$(curl -s "http://localhost:8080/api/auth/callback/google")
if [ "$CALLBACK_RESPONSE" = "unauthorized" ]; then
    echo "✅ Callback endpoint responds correctly (unauthorized without code)"
else
    echo "❌ Callback endpoint unexpected response: $CALLBACK_RESPONSE"
fi

echo ""
echo "5. 🍪 Testing rate limits..."
echo "   Making multiple rapid requests to test rate limiting..."
for i in {1..5}; do
    STATUS=$(curl -s -o /dev/null -w "%{http_code}" "http://localhost:8080/api/auth/me")
    echo "   Request $i: HTTP $STATUS"
    sleep 0.1
done

echo ""
echo "6. 🌐 Checking CORS headers..."
CORS_HEADERS=$(curl -s -I "http://localhost:8080/api/auth/me" | grep -i "access-control")
if [ -n "$CORS_HEADERS" ]; then
    echo "✅ CORS headers present:"
    echo "$CORS_HEADERS" | sed 's/^/   /'
else
    echo "❌ No CORS headers found"
fi

echo ""
echo "=================================="
echo "🎯 Manual Testing Instructions:"
echo "1. Open browser: http://localhost:3000"
echo "2. Open DevTools (F12) → Console tab"
echo "3. Click 'Login with Google'"
echo "4. Complete Google OAuth"
echo "5. Check console logs for debugging info"
echo "6. Verify you're redirected to home page with todos"
echo ""
echo "🔍 If still failing, check browser console for errors!"
