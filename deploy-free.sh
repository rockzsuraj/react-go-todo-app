#!/bin/bash

echo "🚀 Free Deployment Options:"
echo ""
echo "1. Vercel (Frontend + Serverless API)"
echo "   - npm install -g vercel"
echo "   - vercel --prod"
echo ""
echo "2. Netlify (Frontend + Functions)"
echo "   - npm install -g netlify-cli" 
echo "   - netlify deploy --prod"
echo ""
echo "3. GitHub Pages + Supabase"
echo "   - Frontend: GitHub Pages"
echo "   - Database: Supabase (free PostgreSQL)"
echo ""
echo "4. Heroku (Free dyno hours)"
echo "   - git push heroku main"
echo ""

read -p "Choose deployment (1-4): " choice

case $choice in
  1)
    echo "Deploying to Vercel..."
    if ! command -v vercel &> /dev/null; then
      npm install -g vercel
    fi
    vercel --prod
    ;;
  2)
    echo "Deploying to Netlify..."
    if ! command -v netlify &> /dev/null; then
      npm install -g netlify-cli
    fi
    netlify deploy --prod
    ;;
  3)
    echo "Setup GitHub Pages + Supabase manually"
    echo "1. Push to GitHub"
    echo "2. Enable GitHub Pages"
    echo "3. Create Supabase project"
    echo "4. Update API URL"
    ;;
  4)
    echo "Deploying to Heroku..."
    git push heroku main
    ;;
  *)
    echo "Invalid choice"
    ;;
esac