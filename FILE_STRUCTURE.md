nw-guardian/
├── cmd/
│   └── nw-guardian-server/
│       ├── main.go
│       └── config.yaml
│
├── internal/
│   ├── database/
│   │   ├── database.go        # Database connection setup
│   │   ├── users.go           # User-related database operations
│   │   ├── sites.go           # Site-related database operations
│   │   ├── probes.go          # Probe-related database operations
│   │   ├── checks.go          # Check-related database operations
│   │   └── migrations/
│   │       ├── 001_initial.sql
│   │       └── ...
│   │
│   ├── users/               # Users related code
│   │   ├── user.go
│   │   └── repository.go
│   │
│   ├── sites/               # Sites related code
│   │   ├── site.go
│   │   └── repository.go
│   │
│   ├── probes/              # Probes related code
│   │   ├── probe.go
│   │   └── repository.go
│   │
│   ├── checks/              # Checks related code
│   │   ├── check.go
│   │   └── repository.go
│   │
│   └── ...
│
├── web/
│   ├── handlers.go
│   ├── middleware.go
│   └── routes.go
│
├── static/
│   ├── index.html
│   ├── style.css
│   └── ...
│
├── templates/
│   ├── base.html
│   ├── user.html
│   ├── site.html
│   ├── probe.html
│   ├── check.html
│   └── ...
│
├── config/
│   └── config.yaml
│
├── logs/
│   └── app.log
│
├── scripts/
│   └── migrate.sh
│
└── README.md
