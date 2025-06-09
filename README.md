# Cross Competition API

This API provides endpoints for managing cross-country running competitions.

## Features

- User authentication and authorization
- Competition management with zones and participants
- Live ranking system
- Rate limiting for security
- Email notifications
- File upload support (CSV and Excel for participants)

## Configuration

### Environment Variables

#### Database
```env
DB_NAME=cross_competition
DB_URI=user:password@tcp(localhost:3306)/cross_competition
```

#### JWT
```env
JWT_SECRET_KEY=your-secret-key
```

#### Email (SMTP)
```env
EMAIL_HOST=smtp.example.com
EMAIL_PORT=587
EMAIL_USERNAME=your-email@example.com
EMAIL_PASSWORD=your-password
EMAIL_FROM=noreply@example.com
```

#### Rate Limiting (Optional)
```env
# Login endpoint: 5 attempts per 5 minutes (default)
LOGIN_RATE_LIMIT_ATTEMPTS=5
LOGIN_RATE_LIMIT_WINDOW=5m

# Forgot password endpoint: 3 attempts per hour (default)
FORGOT_PASSWORD_RATE_LIMIT_ATTEMPTS=3
FORGOT_PASSWORD_RATE_LIMIT_WINDOW=1h
```

#### CORS and Security
```env
ALLOW_ORIGINS=http://localhost:3000,https://yourdomain.com
SECURE_MODE=true
```

## Rate Limiting

The API includes built-in rate limiting to prevent brute force attacks:

### Protected Endpoints
- **POST /login**: 5 attempts per 5 minutes per IP
- **POST /auth/forgot-password**: 3 attempts per hour per IP

### Features
- IP-based tracking with support for proxy headers (X-Forwarded-For)
- Sliding window approach for accurate rate limiting
- Automatic memory cleanup to prevent leaks
- Configurable limits via environment variables
- Proper HTTP 429 responses with Retry-After headers

### Trusted Proxies
The API is pre-configured to work with OVH SSL Gateway:
- 213.32.4.0/24
- 54.39.240.0/24  
- 144.217.9.0/24

## API Endpoints

### Authentication
- `PUT /login` - User login (rate limited)
- `POST /logout` - User logout
- `PUT /auth/password` - Change password (authenticated)
- `POST /auth/forgot-password` - Reset forgotten password (rate limited)

### Competition Management
- `POST /competition` - Create competition
- `GET /competition` - List competitions
- `POST /competition/zone` - Add zone to competition
- `PUT /competition/zone` - Update zone in competition
- `DELETE /competition/zone` - Delete zone from competition

### Participants
- `POST /competition/participants` - Add participants (bulk CSV/Excel upload)
- `POST /participant` - Create single participant
- `GET /competition/{id}/participants` - List participants by category
- `GET /competition/{id}/participant/{dossard}` - Get specific participant

### Live Results
- `POST /run` - Record a run result
- `GET /competition/{id}/liveranking` - Get live rankings

### Run Management (Admin Only)
- `GET /competition/{competitionID}/participant/{dossard}/runs` - Get all runs for a participant with referee details
- `PUT /run` - Update an existing run
- `DELETE /run` - Delete a run

## Security Features

- JWT-based authentication with refresh tokens
- Password hashing using bcrypt
- Rate limiting on authentication endpoints
- CORS protection
- Secure cookie handling
- Input validation and sanitization

## Development

```bash
# Install dependencies
go mod download

# Run the application
go run cmd/api/main.go rest

# Build
go build ./...
```

## API Documentation

Visit `/swagger/index.html` when the server is running to access the interactive API documentation.

## How to build

### Submodules initialization

`git submodule update --init`

### Install required packages

```bash
sudo apt  install golang-go
go install github.com/swaggo/swag/cmd/swag@latest
```

### Export path

```bash
export PATH="$PATH:$(go env GOPATH)/bin"
```

### Generate documentation
Run `make doc`

### Start the application
Run `make start`

### Access the documentation

Follow [this link](http://localhost:9000/swagger/index.html)

