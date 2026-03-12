#!/usr/bin/env bash

# Very simple test for agg + multiple real feeds
# Just adds 3 feeds and runs agg briefly to see if it scrapes

echo "=== Reset database (if needed) ==="
go run . reset || true

echo ""
echo "=== Create/login test user ==="
go run . register tester || true
go run . login tester

echo ""
echo "=== Add 3 real feeds ==="
go run . addfeed "TechCrunch"        "https://techcrunch.com/feed/"
go run . addfeed "Hacker News"       "https://news.ycombinator.com/rss"
go run . addfeed "Boot.dev Blog"     "https://www.boot.dev/blog/index.xml"

echo ""
echo "=== Show what you're following ==="
go run . following

echo ""
echo "=== Run agg briefly (8 seconds, 1s interval) ==="
echo "Look for:"
echo "  - Collecting feeds every 0m1s"
echo "  - Scraping messages or post titles from the 3 feeds"

# Run in background → capture output → stop after ~8s
timeout 8s go run . agg 1s > agg_test.txt 2>&1 &

sleep 7
kill $! 2>/dev/null || true

echo ""
echo "=== agg output (last 20 lines or so) ==="
tail -n 20 agg_test.txt

echo ""
echo "=== Quick check ==="
if grep -q "Collecting feeds every" agg_test.txt; then
    echo "→ agg started correctly"
else
    echo "→ No starting message → check handlerAgg"
fi

if grep -qi "title" agg_test.txt || grep -qi "scrap" agg_test.txt || grep -qi "fetched" agg_test.txt; then
    echo "→ Looks like it scraped at least something (good!)"
else
    echo "→ No scraping/posts visible → most likely scrapeFeeds is not printing or not finding posts"
fi

rm -f agg_test.txt

echo ""
echo "In real life just run in one terminal:"
echo "   go run . agg 30s"
echo "and watch titles appear every ~30 seconds from all 3 feeds."
