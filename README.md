# Vault Accounting Application

## Description

A simple application that stores and manages accounting information using immudb Vault. The application includes a backend API and a frontend interface for creating and displaying accounting records.

## Features

- **Store Accounting Information**: Manage accounts with with the following structure: account number (unique), account name, iban, address, amount, type (sending, receiving).
- **API**: Add and retrieve accounting information via RESTful API.
- **Frontend**: User-friendly interface to display and create new records.
- **Docker Compose**: Easy setup and deployment using Docker.

## Installation

1. **Clone the repository:**

    ```bash
    git clone https://github.com/hitsumitomo/49755d8472b1b5b4a0d7d889e78dceab.git
    ```

2. **Navigate to the project directory:**

    ```bash
    cd 49755d8472b1b5b4a0d7d889e78dceab
    ```

3. **Update the credentials in docker-compose.yml, then start the application using Docker Compose:**

    ```bash
    docker compose up
    ```

## API Documentation

### Add Account

- **Endpoint**: `/api/add`
- **Method**: `POST`
- **Request Body**:

    ```json
    {
        "number": "123456789",
        "name": "John Doe",
        "iban": "DE89370400440532013000",
        "address": "123 Main St",
        "amount": 1000.50,
        "type": "sending"
    }
    ```

- **Responses**:
    - `200 OK`: Account added successfully.
    - `400 Bad Request`: Invalid account data.
    - `409 Conflict`: Account already exists.
    - `500 Internal Server Error`: Failed to add account.
- Important: Make sure to provide a `correct IBAN` when making API request.
### Get Accounts

- **Endpoint**: `/api/get`
- **Method**: `POST`
- **Request Body**:

    ```json
    {
        "number": "123456789"
    }
    ```

    or

    ```json
    {
        "type": "sending"
    }
    ```

- **Responses**:
    - `200 OK`: Returns list of accounts.
    - `400 Bad Request`: Invalid query.
    - `404 Not Found`: Account not found.
    - `500 Internal Server Error`: Failed to retrieve accounts.

## Frontend

The frontend displays accounting information and allows users to create new records.
Access it by navigating to `http://yourserver.example.com:8080`

## Docker Compose
- Important: Update the credentials before run.
- A `docker-compose.yml` file is provided to set up the application.

```yaml
services:
  vault-app:
    image: vault-app:latest
    build: .
    ports:
      - "8080:8080"
    environment:
      - API_URL=https://vault.immudb.io/ics/api/v1/...
      - API_KEY=default....
      - API_RO_KEY=defaultro....
      - PORT=8080
    command: ["./app"]
```

