# PT. Dana Sejahtera - Fintech Loan Management System

A simulated fintech system for loan management built with Go backend and Next.js frontend.

## Architecture

- **Backend**: Go with Gin framework, Clean Architecture
- **Frontend**: Next.js with TypeScript and Tailwind CSS
- **Database**: PostgreSQL
- **Containerization**: Docker & Docker Compose

## Quick Start

1. Clone the repository
2. Run `docker-compose up --build`
3. Access the application at:
   - Frontend: http://localhost:3000
   - Backend API: http://localhost:8080

## Development

### Backend
```bash
cd backend
go mod tidy
go run cmd/main.go
```

### Frontend
```bash
cd frontend
npm install
npm run dev
```

### Database
```bash
cd database
# Run migrations
```

## Security Notes

This application is designed as a vulnerable target for grey box penetration testing. It contains intentional security vulnerabilities marked with TODO comments for educational purposes.

## API Endpoints

- `POST /api/auth/login` - User authentication
- `POST /api/auth/register` - User registration
- `GET /api/nasabah` - List customers (vulnerable)
- `GET /api/nasabah/:id` - Get customer by ID (BOLA vulnerable)
- `POST /api/nasabah` - Create customer
- `GET /api/loans` - List loans (excessive data exposure)
- `GET /api/loans/:id` - Get loan by ID (BOLA vulnerable)
- `POST /api/loans` - Create loan

## License

This project is for educational purposes only.