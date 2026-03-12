# aggreGATOR

A command-line RSS feed aggregator that lets users subscribe to multiple RSS feeds, track posts, and browse aggregated content in one place.

[![License](https://img.shields.io/badge/license-MIT-blue)]()

## Overview

AggreGATOR is a CLI-based RSS feed aggregator built in Go and PostgreSQL. It allows users to register, subscribe to RSS feeds, and browse aggregated posts from their followed feeds.

## Features

- **User Management**: Register users and switch between accounts
- **Feed Management**: Add and manage RSS feeds
- **Feed Following**: Subscribe to feeds and manage subscriptions
- **Post Aggregation**: Automatically fetch and store posts from subscribed feeds
- **Browse Posts**: View the latest posts from all followed feeds
- **Database-backed**: Persistent storage with PostgreSQL

## Quick Start

### Prerequisites

- Go 1.20+
- PostgreSQL 12+
- Goose (database migration tool)
- SQLC (SQL code generator)

### Installation

First, install the required tools:

1. **Goose** - Database migration tool for managing schema changes:
```bash
go install github.com/pressly/goose/v3/cmd/goose@latest
```

2. **SQLC** - SQL code generator that produces type-safe Go code from your SQL queries:
```bash
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
```

These tools are essential for managing the database schema and generating query methods used throughout the application.

### Setup Database

Before running the application, you need to set up a PostgreSQL database (assume you're in linux ubuntu enviroment):

1. **Install PostgreSQL** (on Linux):
```bash
sudo apt update
sudo apt install postgresql postgresql-contrib
```

2. **Start the PostgreSQL service**:
```bash
sudo service postgresql start
```

3. **Connect to PostgreSQL as the default user**:
```bash
sudo -u postgres psql
```

4. **Create the gator database and set credentials**:
```sql
CREATE DATABASE gator;
\c gator
ALTER USER postgres PASSWORD 'postgres';
```

This creates a new database called `gator` and sets the postgres user password. You'll use these credentials in your configuration file.

5. **Verify the connection** (optional):
```bash
psql -U postgres -d gator -h localhost
```

### Run Migrations

```bash
cd sql/schema/
goose postgres postgres://postgres:postgres@localhost:5432/gator up
```

### Configuration

Create `~/.gatorconfig.json`:

```json
{
  "db_url": "postgres://postgres:postgres@localhost:5432/gator?sslmode=disable"
}
```

## Usage

### User Commands

```bash
# Register a new user
gator register username

# Login as a user
gator login username

# View all users
gator users

# Reset database (delete all users)
gator reset
```

### Feed Commands

```bash
# Add a new feed (requires login)
gator addfeed "Feed Name" "https://example.com/rss"

# View all feeds
gator feeds

# Subscribe to a feed
gator follow "https://example.com/rss"

# View your followed feeds
gator following

# Unsubscribe from a feed
gator unfollow "https://example.com/rss"
```

### Content Commands

```bash
# Fetch new posts from all feeds
gator agg

# Browse latest posts (limit default: 2)
gator browse 10
```

## Architecture

- **SQLc**: Type-safe database queries generated from SQL
- **Goose**: Database schema migrations
- **Middleware**: Authentication layer for logged-in-only commands
- **RSS Parser**: Handles multiple RSS date formats

## Contributing

Contributions welcome. Please ensure database migrations are included.

## Acknowledgments

This project was built with guidance from **Boot.dev**, a hands-on backend development learning platform. The project demonstrates real-world patterns including database migrations, type-safe queries with SQLc, and CLI application design in Go.

## License

MIT