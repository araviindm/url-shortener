# URL Shortener Service

## DB Setup

- Create a redis database and retrive the REDIS_ADDR, and REDIS_PASSWORD.
- Create a database (URLShortener) in a mongodb cluster and retrive the MONGO_URI, and MONGODB_NAME.

## .env file

```bash
PORT=8080
MONGO_URI=value
MONGODB_NAME=URLShortener
REDIS_ADDR=value
REDIS_PASSWORD=value
```

## To run the app locally

- Git clone the code, go to the root directory

```bash
go run main.go
```

- The go application will be running on http://localhost:8080/

## Test cases

- Set env varibles to run from terminal.

```bash
set ENV_KEY = value
```

```bash
$env:ENV_KEY = value
```

- To run the test cases,

```bash
go test -v ./...
```

## Docker

- Creating a docker container

```bash
docker build -t url-shortener .
docker run -d -p 8081:8081 -e PORT=8081 -e MONGO_URI=value -e MONGODB_NAME=URLShortener -e REDIS_ADDR=value -e REDIS_PASSWORD=value url-shortener
```

- The container will be running on http://localhost:8081/
