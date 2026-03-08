# Password Generator

Generate secure random passwords and update auth configuration.

## Usage

```bash
# Update all users (default)
go run ./cmd/genpass/main.go

# Update specific user
go run ./cmd/genpass/main.go --user admin

# Custom length
go run ./cmd/genpass/main.go --length 24

# Custom config path
go run ./cmd/genpass/main.go --config path/to/auth.yml
```

## Options

| Flag | Description | Default |
|------|-------------|---------|
| `--user` | Specific user to update | all users |
| `--length` | Password length | 16 |
| `--config` | Path to auth.yml | internal/config/auth.yml |

## Notes

- When password is updated, the user's session is automatically cleared
- Old and new passwords are displayed in the output
- Uses `crypto/rand` for cryptographically secure random generation
