# Database Setup Instructions

## If you're getting enum errors, follow these steps:

### Option 1: Reset Database (Recommended)
1. Connect to your PostgreSQL database:
   ```bash
   psql -U postgres -d motiv
   ```

2. Run the reset script:
   ```sql
   \i scripts/reset-db.sql
   ```

3. Exit psql and restart your Go application:
   ```bash
   \q
   go run main.go
   ```

### Option 2: Manual Reset
1. Drop the database and recreate it:
   ```bash
   psql -U postgres
   DROP DATABASE motiv;
   CREATE DATABASE motiv;
   \q
   ```

2. Restart your Go application:
   ```bash
   go run main.go
   ```

### Option 3: Continue with Warnings
The current migration will work with basic functionality even if enums fail. You'll see warnings but the core features (users, events, tickets) will work.

## Database Connection
Make sure your `.env` file has the correct database settings:
```
DB_HOST=localhost
DB_USER=postgres
DB_PASSWORD=motiv-2025
DB_NAME=motiv
DB_PORT=5432
JWT_SECRET=L8/r9Z3p8qX+v/L8/r9Z3p8qX+v/L8/r9Z3p8qX+v/L8ww/r9Z3p8qX+v/A==
```

## Testing the Setup
After the migration completes, test the API:
```bash
curl http://localhost:8080/health
```

You should see:
```json
{
  "status": "ok",
  "message": "MOTIV Backend is running",
  "version": "1.0.0"
}
```