# Steppe Way - Comprehensive Travel Platform
## Front - https://github.com/tenajuro12/Steppe-way-Front
## Overview
This Travel App is a modern, microservices-based platform designed to provide users with a complete travel planning and exploration experience. The application allows users to discover accommodations, attractions, local food places, events, and create personalized travel plans.
Front - https://github.com/tenajuro12/Steppe-way-Front

## Architecture
The application follows a microservices architecture with the following key components:

- **Gateway Service**: Central API gateway that routes requests to appropriate microservices
- **Auth Service**: Handles user authentication and authorization
- **Profile Service**: Manages user profiles and social features
- **Accommodation Service**: Manages listings, bookings, and reviews for accommodations
- **Attraction Service**: Provides information about tourist attractions
- **Event Service**: Manages events and event bookings
- **Food Service**: Features restaurants, cafes, and food reviews
- **Blog Service**: Supports travel blogs and discussions
- **Plan Service**: Enables users to create and customize travel itineraries
- **Favorites Service**: Allows users to save and manage favorite places
- **Review Service**: Handles user reviews for various services

## Technology Stack

### Backend
- **Language**: Golang (versions 1.23-1.24)
- **Database**: PostgreSQL
- **ORM**: GORM
- **API Framework**: Gorilla Mux
- **Container**: Docker and Docker Compose

### Key Features

#### User Management
- Registration and login
- Profile creation and customization
- Following other users

#### Accommodations
- Detailed accommodation listings with amenities
- Different room types and availability
- Location-based search
- Reviews and ratings

#### Attractions
- Tourist attraction listings
- Categorized attractions
- User reviews

#### Food Places
- Restaurant and café listings
- Menu items and specialties
- Cuisine types
- Food reviews

#### Events
- Local events calendar
- Event details and registration
- Location-based event discovery

#### Travel Planning
- Create personalized travel plans
- Add attractions, restaurants and events to plans
- Route optimization
- Template-based plans

#### Social Features
- Travel blogs
- Commenting and liking
- Follow other travelers

#### Favorites
- Save favorite places, events, and attractions
- Organized favorites management

## Service Details

### Gateway Service (Port: 8080)
Central entry point that routes requests to the appropriate microservices. Handles CORS and basic authentication validation.

### Auth Service (Port: 8082)
Handles user authentication, registration, and session management.

### Profile Service (Port: 8084)
Manages user profiles, profile pictures, and social connections.

### Accommodation Service (Port: 8089)
Comprehensive service for accommodation listings, booking management, and reviews.

### Attraction Service (Port: 8085)
Provides information about tourist attractions, including details, images, and categories.

### Event Service (Port: 8083)
Manages events, event categories, and registrations.

### Food Service (Port: 8090)
Features restaurants, cafes, menus, cuisine types, and food reviews.

### Blog Service (Port: 8081)
Supports travel blogs, comments, and social interactions.

### Plan Service (Port: 8087)
Enables users to create and manage travel itineraries, including route optimization.

### Favorites Service (Port: 8088)
Allows users to save and organize favorite places and attractions.

### Review Service (Port: 8086)
Centralized service for user reviews across different categories.

## Setup and Installation

### Prerequisites
- Docker and Docker Compose
- Go (version 1.23+)
- PostgreSQL (handled by Docker)

### Steps to Run

1. **Clone the repository**
   ```bash
   git clone https://github.com/yourusername/travel-app.git
   cd travel-app
   ```

2. **Start the services with Docker Compose**
   ```bash
   docker-compose up
   ```

3. **Access the application**
   The application will be available at http://localhost:8080

### Development Setup

1. **Install Go dependencies for a specific service**
   ```bash
   cd backend/[service_name]
   go mod download
   ```

2. **Run a specific service locally**
   ```bash
   cd backend/[service_name]
   go run cmd/main.go
   ```

## Database Structure
The application uses PostgreSQL with multiple schemas for different services. The database migrations are handled automatically when the services start.

## API Documentation

### Authentication
- `POST /login`: User login
- `POST /register`: User registration
- `GET /profile`: Get user profile
- `POST /profile`: Update user profile

### Accommodations
- `GET /places`: List accommodations
- `GET /places/{id}`: Get accommodation details
- `POST /admin/places`: Create accommodation (admin)
- `PUT /admin/places/{id}`: Update accommodation (admin)

### Attractions
- `GET /attractions`: List attractions
- `GET /attractions/{id}`: Get attraction details
- `POST /admin/attractions`: Create attraction (admin)

### Events
- `GET /events`: List events
- `GET /events/{id}`: Get event details
- `POST /admin/events`: Create event (admin)

### Food
- `GET /places`: List food places
- `GET /places/{id}`: Get food place details
- `GET /dishes/search`: Search for dishes
- `POST /admin/places`: Create food place (admin)

### Plans
- `GET /api/plans`: List user plans
- `POST /api/plans`: Create plan
- `GET /api/plans/{id}`: Get plan details
- `PUT /api/plans/{id}`: Update plan
- `POST /api/plans/{id}/items`: Add item to plan
- `POST /api/plans/{id}/optimize`: Optimize route

### Blogs
- `GET /blogs`: List blogs
- `POST /blogs`: Create blog
- `GET /blogs/{id}`: Get blog details
- `POST /blogs/{id}/comments`: Comment on blog

### Favorites
- `GET /favorites`: List favorites
- `POST /favorites`: Add to favorites
- `DELETE /favorites/{type}/{id}`: Remove from favorites
- `GET /favorites/check/{type}/{id}`: Check if item is favorited

## Contributing
Guidelines for contributing to the project:

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Submit a pull request
