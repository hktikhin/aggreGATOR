module github.com/hktikhin/aggreGATOR

go 1.26.1

require internal/config v1.0.0

replace internal/config => ./internal/config

require (
	github.com/google/uuid v1.6.0
	github.com/lib/pq v1.11.2
)

require internal/database v0.0.0-00010101000000-000000000000 // indirect

replace internal/database => ./internal/database
