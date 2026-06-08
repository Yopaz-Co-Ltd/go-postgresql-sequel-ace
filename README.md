# PostgreSQL Web Client

Dark-mode PostgreSQL client with a Vue frontend and Go backend. The browser UI talks to the Go API, and Go owns the PostgreSQL connection.

## Structure

- `backend/`: Go API using `pgxpool`
- `frontend/`: Vue 3 + Vite web app

## Run

Start the sample PostgreSQL database:

```bash
cd postgresql
docker compose up -d
```

Sample connection:

- Host: `127.0.0.1`
- Port: `5432`
- Username: `postgres`
- Password: `postgres`
- Database: `sample_store`
- SSL Mode: `disable`

```bash
cd backend
go run .
```

In another terminal:

```bash
cd frontend
npm install
npm run dev
```

Open the Vite URL shown in the terminal. The frontend proxies `/api` to `http://localhost:8080`.

## Production Build

```bash
cd frontend
npm run build
cd ../backend
go run .
```

When `frontend/dist` exists, the Go server serves it at `http://localhost:8080`.

## Reset Sample Database

```bash
cd postgresql
docker compose down -v
docker compose up -d
```
