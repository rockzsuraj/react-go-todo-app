#!/bin/bash

echo "🚀 Setting up Supabase + GitHub Pages deployment..."

echo ""
echo "📋 Steps to deploy:"
echo ""
echo "1. Copy and run this SQL in Supabase SQL Editor:"
echo "   👉 supabase/migrations/001_create_todos.sql"
echo ""
echo "2. Push to GitHub:"
echo "   git add ."
echo "   git commit -m 'Add Supabase integration'"
echo "   git push origin main"
echo ""
echo "3. Enable GitHub Pages:"
echo "   - Go to repo Settings > Pages"
echo "   - Source: GitHub Actions"
echo "   - The workflow will auto-deploy"
echo ""
echo "4. Your app will be live at:"
echo "   https://YOUR_USERNAME.github.io/react-todo-demo"
echo ""
echo "✅ 100% Free Forever:"
echo "   - Frontend: GitHub Pages (unlimited)"
echo "   - Database: Supabase (500MB free)"
echo "   - API: Supabase REST API (unlimited)"
echo ""

read -p "Ready to proceed? (y/n): " confirm

if [[ $confirm == [yY] || $confirm == [yY][eE][sS] ]]; then
    echo ""
    echo "📄 Opening migration file..."
    cat supabase/migrations/001_create_todos.sql
    echo ""
    echo "👆 Copy this SQL and run it in Supabase SQL Editor"
    echo ""
    echo "🔗 Supabase SQL Editor: https://qnlhgaymddnazecbtsau.supabase.co/project/qnlhgaymddnazecbtsau/sql"
else
    echo "Setup cancelled."
fi