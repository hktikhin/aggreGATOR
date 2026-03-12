#!/usr/bin/env bash

# Simple smoke test for gator CLI — now including unfollow
# Run with:  bash test.sh

set -e

echo "=== Resetting database ==="
go run . reset || echo "reset failed (maybe already empty)"

echo ""
echo "=== Register & login test user ==="
go run . register testuser123 || echo "user may already exist — continuing"
go run . login testuser123

echo ""
echo "=== Adding a feed (should auto-follow) ==="
go run . addfeed "Test Blog" "https://www.wagslane.dev/index.xml"

echo ""
echo "=== Listing what you're following (should show the feed) ==="
go run . following

echo ""
echo "=== Unfollowing the feed ==="
go run . unfollow "https://www.wagslane.dev/index.xml"

echo ""
echo "=== Checking following again (should be empty now) ==="
go run . following

echo ""
echo "=== Listing all feeds (should still exist) ==="
go run . feeds

echo ""
echo "=== Done. Quick checklist:"
echo "  1. addfeed → shows 'The feed follow was created'"
echo "  2. following → shows feed name after addfeed"
echo "  3. unfollow → shows 'The feed follow was deleted' + username + feed name"
echo "  4. following after unfollow → should show nothing (or no feeds)"
echo ""
echo "If you see those 4 things → unfollow is probably working correctly."
