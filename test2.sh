#!/usr/bin/env bash

# Minimal test: agg inserts posts → browse shows them
# Uses one reliable feed for quick testing

echo "=== Reset database ==="
go run . reset || true

echo ""
echo "=== Register & login test user ==="
go run . register tester || true
go run . login tester

echo ""
echo "=== Add one feed that usually has many items ==="
go run . addfeed "TechCrunch" "https://techcrunch.com/feed/"

echo ""
echo "=== Run agg very briefly (should insert some posts) ==="
echo "Expect to see:"
echo "  Collecting feeds every..."
echo "  Post created with title ..."

timeout 10s go run . agg 2s > agg_temp.txt 2>&1 &

sleep 8
kill $! 2>/dev/null || true

echo ""
echo "=== agg output preview (should show post creations) ==="
tail -n 12 agg_temp.txt

echo ""
echo "=== Now browse latest posts (should show 2 or more) ==="
echo "Expect output like:"
echo "Getting the latest X posts..."
echo "1. [2026-03-..] Some TechCrunch title"
echo "   URL: https://..."
echo "   Desc: ..."

go run . browse

echo ""
echo "=== If you see post titles above → agg + CreatePost + browse works!"
echo ""
echo "Common failure modes:"
echo "  - No 'Post created ...' lines → check scrapeFeeds loop & CreatePost call"
echo "  - browse shows nothing → check getPostsForUser query or postLimit"
echo "  - duplicate url error only → good (your code skips them)"
echo ""
echo "Tip: Increase agg time (e.g. timeout 20s go run . agg 3s) if no posts appear yet."
