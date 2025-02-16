# Swift Code API

This project provides a RESTful API for managing SWIFT codes using Go, Gin, and BoltDB.

## Technologies Used

- **Go**: Programming language used for developing the application.
- **Gin**: Web framework used for building the API.
- **BoltDB**: Embedded key/value database used for storing SWIFT codes.
- **Docker**: Containerization platform used for running the application and database.
- **Docker Compose**: Tool for defining and running multi-container Docker applications.

## Repository Structure

- `main.go`: The main entry point of the application.
- `handlers/handlers.go`: Contains the handler functions for the API endpoints.
- `tests/handlers_test.go`: Contains the test cases for the handlers.
- `swiftCodes.csv`: The CSV file containing the initial SWIFT codes data.
- `docker-compose.yml`: Docker Compose configuration for running the application and database.

## Prerequisites

- Go 1.16 or higher
- Docker and Docker Compose
- Git

## Setup

1. Clone the repository:

    ```sh
    git clone https://github.com/yourusername/swiftTask.git
    cd swiftTask
    ```

2. Install dependencies:

    ```sh
    go mod tidy
    ```

3. Ensure the `swiftCodes.csv` file is in the project root directory.

## Running the Application

### Locally

1. Build and run the application:

    ```sh
    go run main.go
    ```

2. The server will start on port 8080. You can access the API at `http://localhost:8080`.

### Using Docker

1. Build and start the application using Docker Compose:

    ```sh
    docker-compose up --build
    ```

2. The server will start on port 8080. You can access the API at `http://localhost:8080`.

## API Endpoints

- **GET /v1/swift-codes/:swift-code**: Retrieve details of a specific SWIFT code.
- **DELETE /v1/swift-codes/:swift-code**: Delete a specific SWIFT code.
- **GET /v1/swift-codes/country/:countryISO2code**: Retrieve all SWIFT codes for a specific country.
- **POST /v1/swift-codes**: Add a new SWIFT code.

### Endpoint Details

#### GET /v1/swift-codes/:swift-code

- **Description**: Retrieve details of a specific SWIFT code.
- **Response**:
  - If the SWIFT code is a headquarters, it returns the headquarters details along with its branches.
  - If the SWIFT code is a branch, it returns only the branch details.

#### DELETE /v1/swift-codes/:swift-code

- **Description**: Delete a specific SWIFT code.
- **Response**:
  - Success message if the SWIFT code is deleted.
  - Error message if the SWIFT code is not found or there is a database error.

#### GET /v1/swift-codes/country/:countryISO2code

- **Description**: Retrieve all SWIFT codes for a specific country.
- **Response**:
  - List of SWIFT codes for the specified country.
  - Error message if no SWIFT codes are found for the country.

#### POST /v1/swift-codes

- **Description**: Add a new SWIFT code.
- **Request Body**: JSON object containing the SWIFT code details.
- **Response**:
  - Success message if the SWIFT code is added.
  - Error message if there is a database error or invalid request body.

## Testing

1. To run the tests, navigate to the `tests` directory:

    ```sh
    cd tests
    ```

2. Run the tests using `go test`:

    ```sh
    go test
    ```

### Test Details

- **handlers_test.go**: Contains test cases for the handler functions. It tests the following scenarios:
  - Retrieving a SWIFT code.
  - Deleting a SWIFT code.
  - Retrieving SWIFT codes by country.
  - Adding a new SWIFT code.

## Notes

- The database file `swift_codes.db` will be created in the project root directory.
- The CSV file should use semicolons (`;`) as the delimiter.
