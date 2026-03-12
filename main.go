
package main

import (
  "log"
  "fmt"
  "internal/config"
  "encoding/xml"
  "errors"
  "os"
  "context"
  "time"
  "net/http"
  "io"
  "html"
  "strconv"
  "internal/database"
  "database/sql"
  "github.com/lib/pq"
  "github.com/google/uuid"
)

type state struct {
  db *database.Queries
  cfg *config.Config
}

type command struct {
  name string
  args []string
}

type commands struct {
  Items map[string]func(*state, command) error
}

type RSSFeed struct {
	Channel struct {
		Title       string    `xml:"title"`
		Link        string    `xml:"link"`
		Description string    `xml:"description"`
		Item        []RSSItem `xml:"item"`
	} `xml:"channel"`
}

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

func fetchFeed(ctx context.Context, feedURL string) (*RSSFeed, error) {
  var feed RSSFeed
  req, err := http.NewRequestWithContext(ctx, "GET", feedURL, nil)
  if err != nil {
    return &feed, err
	}
  req.Header.Set("User-Agent", "gator")
  client := &http.Client{}
	resp, err := client.Do(req)
  if err != nil {
		return &feed, err
	}
	defer resp.Body.Close()

  data, err := io.ReadAll(resp.Body)
  if err != nil {
    return &feed, err
  }

  err = xml.Unmarshal(data, &feed)
  if err != nil {
    return &feed, err
  }
  for _, item := range feed.Channel.Item {
    item.Title = html.UnescapeString(item.Title)
    item.Description = html.UnescapeString(item.Description)
  }
  return &feed, nil
}

func parseRSSDate(dateStr string) (time.Time, error) {
	formats := []string{
		time.RFC1123Z,                // "Mon, 02 Jan 2006 15:04:05 -0700"
		time.RFC1123,                 // "Mon, 02 Jan 2006 15:04:05 MST"
		time.RFC3339,                 // "2006-01-02T15:04:05Z07:00"
		"Mon, 2 Jan 2006 15:04:05 -0700", // No leading zero on day
		"Mon, 2 Jan 2006 15:04:05 MST",   // No leading zero on day
		"02 Jan 2006 15:04:05 -0700", // Missing weekday
		"02 Jan 2006 15:04:05 MST",   // Missing weekday
		"2006-01-02 15:04:05",        // Plain ISO-like
	}

	for _, format := range formats {
		t, err := time.Parse(format, dateStr)
		if err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("could not parse date: %s", dateStr)
}

func scrapeFeeds(ctx context.Context, db *database.Queries) error{
  feed, err := db.GetNextFeedToFetch(ctx)
  if err != nil {
    return err
  }
  lastFetched := "null"
  if feed.LastFetchedAt.Valid {
      lastFetched = feed.LastFetchedAt.Time.Format("2006-01-02 15:04:05")
  }
  fmt.Print("Fetching the feed...\n")
  fmt.Printf(
      "ID: %v,\nCreatedAt: %v,\nUpdatedAt: %v,\nLastFetchAt: %v,\nName: %v,\nUrl: %v,\nUserID: %v\n",
      feed.ID,
      feed.CreatedAt,
      feed.UpdatedAt,
      lastFetched,
      feed.Name,
      feed.Url,
      feed.UserID,
  )
  err = db.MarkFeedFetched(ctx, feed.ID)
  if err != nil {
    return err
  }
  fmt.Print("Marked the feed as fetched...\n")
  feedData, err := fetchFeed(ctx, feed.Url)
  if err != nil {
    return err
  }
  fmt.Printf("--- Feed: %s ---\n", feedData.Channel.Title)
  for _, item := range feedData.Channel.Item {
    var publishedAt sql.NullTime

    pubDate, err := parseRSSDate(item.PubDate)
    if err != nil {
        publishedAt = sql.NullTime{Valid: false}
    } else {
        publishedAt = sql.NullTime{
            Time:  pubDate,
            Valid: true,
        }
    }
    post, err := db.CreatePost(
      ctx,
      database.CreatePostParams {
        ID: uuid.New(),
        CreatedAt: time.Now(),
        UpdatedAt: time.Now(),
        Title: item.Title,
        Url: item.Link,
        Description: sql.NullString{String: item.Description, Valid: item.Description != ""},
        PublishedAt: publishedAt,
        FeedID: feed.ID,
      },
    )
    if pgErr, ok := err.(*pq.Error); ok && pgErr.Code == "23505" {
      fmt.Printf("Error: a post with that url already exists.\n")
      continue
    }
    if err != nil {
      fmt.Printf("Error: %v", err)
      continue
    }
    fmt.Printf("Post created with title %v and url %v\n", post.Title, post.Url)
  }
  fmt.Println("--------------------")
  return nil
}

func handlerLogin(s *state, cmd command) error {
  if len(cmd.args) == 0 {
    return errors.New("The login handler expects a single argument, the username.\n")
  }
  _, err := s.db.GetUser(
    context.Background(),
    cmd.args[0],
  )
  if err != nil {
    if errors.Is(err, sql.ErrNoRows) {
        fmt.Println("A user with that name does not exist")
        os.Exit(1)
    }
    fmt.Printf("Database error: %v\n", err)
    os.Exit(1)
  }
  err = s.cfg.SetUser(cmd.args[0])
  if err != nil {
    return err
  }
  fmt.Print("The user has been set.\n")
  return nil
}

func handlerRegister (s *state, cmd command) error {
  if len(cmd.args) == 0 {
    return errors.New("The register handler expects a single argument, the username.\n")
  }
  user, err := s.db.CreateUser(
    context.Background(),
    database.CreateUserParams {
      ID: uuid.New(),
      CreatedAt: time.Now(),
      UpdatedAt: time.Now(),
      Name: cmd.args[0],
    },
  )
  if pgErr, ok := err.(*pq.Error); ok && pgErr.Code == "23505" {
    fmt.Print("Error: a user with that name already exists.")
    os.Exit(1)
  }
  if err != nil {
    return err
  }
  err = s.cfg.SetUser(cmd.args[0])
  if err != nil {
    return err
  }
  fmt.Print("The user was created.\n")
  fmt.Printf(
    "ID: %v,\nCreatedAt: %v,\nUpdatedAt: %v,\nName: %v\n",
    user.ID,
    user.CreatedAt,
    user.UpdatedAt,
    user.Name,
  )
  return nil
}

func handlerReset (s *state, cmd command) error {
  err := s.db.DeleteUsers(
    context.Background(),
  )
  if err != nil {
    fmt.Printf("Error: %v", err)
    os.Exit(1)
  }
  return nil
}

func handlerUsers (s *state, cmd command) error {
  users, err := s.db.GetUsers(
    context.Background(),
  )
  if err != nil {
    fmt.Printf("Error: %v", err)
    os.Exit(1)
  }
  for _, u := range users {
    if u == s.cfg.CurrentUserName{
      fmt.Printf("* %v (current)\n", u)
    } else {
      fmt.Printf("* %v\n", u)
    }
  }
  return nil
}

func handlerAgg (s *state, cmd command) error {
  if len(cmd.args) == 0 {
    return errors.New("The agg handler expects a single argument, the time_between_reqs like 1s, 1m, 1h, etc.\n")
  }
  timeBetweenRequests, err := time.ParseDuration(cmd.args[0])
  if err != nil {
    return err
  }
  mins := int(timeBetweenRequests.Minutes())
  secs := int(timeBetweenRequests.Seconds()) % 60
  fmt.Printf("Collecting feeds every %dm%ds\n", mins, secs)

  ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
  ticker := time.NewTicker(timeBetweenRequests)
  for ; ; <-ticker.C {
    _ = scrapeFeeds(ctx, s.db)
  }
  return nil
}

func handlerAddFeed (s *state, cmd command, user database.User) error {
  if len(cmd.args) < 2 {
    return errors.New("The addfeed handler expects two arguments, the feed name and url.\n")
  }

  feed, err := s.db.CreateFeed(
    context.Background(),
    database.CreateFeedParams {
      ID: uuid.New(),
      CreatedAt: time.Now(),
      UpdatedAt: time.Now(),
      Name: cmd.args[0],
      Url: cmd.args[1],
      UserID: user.ID,
    },
  )
  if pgErr, ok := err.(*pq.Error); ok && pgErr.Code == "23505" {
    fmt.Print("Error: a feed with that URL already exists.")
    os.Exit(1)
  }
  if err != nil {
    return err
  }
  fmt.Print("The feed was created.\n")
  fmt.Printf(
      "ID: %v,\nCreatedAt: %v,\nUpdatedAt: %v,\nName: %v,\nUrl: %v,\nUserID: %v\n",
      feed.ID,
      feed.CreatedAt,
      feed.UpdatedAt,
      feed.Name,
      feed.Url,
      feed.UserID,
  )
  feedFollow, err := s.db.CreateFeedFollow(
      context.Background(),
      database.CreateFeedFollowParams {
        ID: uuid.New(),
        CreatedAt: time.Now(),
        UpdatedAt: time.Now(),
        UserID: user.ID,
        FeedID: feed.ID,
      },
  )
  if err != nil {
    return err
  }
  fmt.Print("The feed follow was created.\n")
  fmt.Printf(
    "ID: %v,\nCreatedAt: %v,\nUpdatedAt: %v,\nUserName: %v\nFeedName: %v\n",
    feedFollow.ID,
    feedFollow.CreatedAt,
    feedFollow.UpdatedAt,
    feedFollow.UserName,
    feedFollow.FeedName,
  )
  return nil
}

func handlerFeeds(s *state, cmd command) error {
	feeds, err := s.db.GetFeeds(context.Background())
	if err != nil {
		return err
	}

	for _, f := range feeds {
		fmt.Printf("* Name:          %s\n", f.Name)
		fmt.Printf("  URL:           %s\n", f.Url)
		fmt.Printf("  Created By:    %s\n", f.UserName)
		fmt.Println("  --------------------")
	}
	return nil
}

func handlerFollow (s *state, cmd command, user database.User) error {
  if len(cmd.args) == 0 {
    return errors.New("The follow handler expects a single argument, the feed url.\n")
  }
	feed, err := s.db.GetFeed(
    context.Background(),
    cmd.args[0],
  )
	if err != nil {
		return err
	}
  feedFollow, err := s.db.CreateFeedFollow(
    context.Background(),
    database.CreateFeedFollowParams {
      ID: uuid.New(),
      CreatedAt: time.Now(),
      UpdatedAt: time.Now(),
      UserID: user.ID,
      FeedID: feed.ID,
    },
  )
  if pgErr, ok := err.(*pq.Error); ok && pgErr.Code == "23505" {
    fmt.Print("Error: You have already followed this feed.")
    os.Exit(1)
  }
  if err != nil {
    return err
  }
  fmt.Print("The feed follow was created.\n")
  fmt.Printf(
    "ID: %v,\nCreatedAt: %v,\nUpdatedAt: %v,\nUserName: %v\nFeedName: %v\n",
    feedFollow.ID,
    feedFollow.CreatedAt,
    feedFollow.UpdatedAt,
    feedFollow.UserName,
    feedFollow.FeedName,
  )
  return nil
}

func handlerFollowing(s *state, cmd command, user database.User) error {
	feeds, err := s.db.GetFeedFollowsForUser(
    context.Background(),
    user.Name,
  )
	if err != nil {
		return err
	}

  for _, f := range feeds {
      fmt.Printf("* Feed Name:     %s\n", f.FeedName)
      fmt.Printf("  User Name:     %s\n", f.UserName)
      fmt.Printf("  Feed ID:       %s\n", f.FeedID)
      fmt.Printf("  Follow ID:     %s\n", f.ID)
      fmt.Printf("  Followed At:   %s\n", f.CreatedAt.Format("2006-01-02 15:04:05"))
      fmt.Println("  --------------------")
  }
  return nil
}


func handlerUnfollow (s *state, cmd command, user database.User) error {
  if len(cmd.args) == 0 {
    return errors.New("The unfollow handler expects a single argument, the feed url.\n")
  }
	feed, err := s.db.GetFeed(
    context.Background(),
    cmd.args[0],
  )
	if err != nil {
		return err
	}
  err = s.db.DeleteFeedFollow(
    context.Background(),
    database.DeleteFeedFollowParams {
      UserID: user.ID,
      FeedID: feed.ID,
    },
  )
  if err != nil {
    return err
  }
  fmt.Print("The feed follow was deleted.\n")
  fmt.Printf(
    "UserName: %v\nFeedName: %v\n",
    user.Name,
    feed.Name,
  )
  return nil
}


func handlerBrowse (s *state, cmd command, user database.User) error {
  postLimit := 2
  if len(cmd.args) > 0 {
      limit, err := strconv.Atoi(cmd.args[0])
      if err != nil {
          fmt.Printf("invalid limit: %v (must be a number)", cmd.args[0])
      }
      if limit <= 0 {
          fmt.Printf("limit must be a positive integer")
      }
      postLimit = limit
  }
	posts, err := s.db.GetPostsForUser(
    context.Background(),
    database.GetPostsForUserParams {
      ID: user.ID,
      Limit: int32(postLimit),
    },
  )
	if err != nil {
		return err
	}
  fmt.Printf("Geting the latest %d posts...\n", postLimit)
  for i, p := range posts {
      description := "No description"
      if p.Description.Valid {
          description = p.Description.String
      }

      pubDate := "Unknown date"
      if p.PublishedAt.Valid {
          pubDate = p.PublishedAt.Time.Format("2006-01-02 15:04")
      }

      fmt.Printf("%d. [%s] %s\n", i+1, pubDate, p.Title)
      fmt.Printf("   URL:  %s\n", p.Url)
      fmt.Printf("   Desc: %s\n", description)
      fmt.Println("--------------------------------------")
  }
  return nil
}

func middlewareLoggedIn(handler func(s *state, cmd command, user database.User) error) func(*state, command) error{
  return func (s *state, cmd command) error {
    if s.cfg.CurrentUserName == "" {
      return errors.New("you must be logged in to perform this action")
    }
    user, err := s.db.GetUser(
      context.Background(),
      s.cfg.CurrentUserName,
    )
    if err != nil {
      return err
    }
    return handler(s, cmd, user)
  }
}

func (c *commands) run(s *state, cmd command) error {
  f, found := c.Items[cmd.name]
  if !found {
    return errors.New("The given command not exists\n")
  }
  err := f(s, cmd)
  if err != nil {
    return err
  }
  return nil
}

func (c *commands) register(name string, f func(*state, command) error) error {
  c.Items[name] = f
  return nil
}

func main() {
  cfg, err := config.Read()
  if err != nil {
      log.Fatalf("error: could not read config file: %v\n", err)
  }
  db, err := sql.Open("postgres", cfg.DBURL)
  dbQueries := database.New(db)

  s := state {
    db: dbQueries,
    cfg: &cfg,
  }
  cmds := commands{
    Items: make(map[string]func(*state, command) error),
  }
  cmds.register("login", handlerLogin)
  cmds.register("register", handlerRegister)
  cmds.register("reset", handlerReset)
  cmds.register("users", handlerUsers)
  cmds.register("agg", handlerAgg)
  cmds.register("addfeed", middlewareLoggedIn(handlerAddFeed))
  cmds.register("feeds", handlerFeeds)
  cmds.register("follow", middlewareLoggedIn(handlerFollow))
  cmds.register("following", middlewareLoggedIn(handlerFollowing))
  cmds.register("unfollow", middlewareLoggedIn(handlerUnfollow))
  cmds.register("browse", middlewareLoggedIn(handlerBrowse))

  args := os.Args
  if len(args) < 2 {
    log.Fatalf("error: command name must be provided\n")
  }
  cmd := command{
    name: args[1],
    args: args[2:],
  }
  err = cmds.run(&s, cmd)
  if err != nil {
    log.Fatalf("error: %v", err)
  }
}
