# Webserver and Bank App

## Project Overview

This project simulates a simple bank system with accounts management, which consists of two servers:
- **HTTP Server:** Exposes a RESTful API for basic CRUD operations on bank accounts. Interactions can be performed using API clients such as Postman.
 - **gRPC Server:** Provides a gRPC service for executing payment transactions. The HTTP Webserver acts as the gRPC client.
---

## Setup Instructions

### Requirements

- **protoc** for protobuf compilation (Maybe not needed). To compile protobuf, run:
```bash
$ cd operations
$ protoc --go_out=paths=source_relative:. --go-grpc_out=paths=source_relative:. payment.proto
```

- **docker engine** to install MySQL docker container
- **Postman** to use as API client

### Setup the bank application

- **Database Setup:**
    - Set up mysql database using docker container (the credentials are in bank-app/app.go)
    ```bash
    $ docker run --name bank-app-mysql -e MYSQL_ROOT_PASSWORD=fakebank1234 -p 3306:3306 -d mysql
    $ docker ps
    CONTAINER ID   IMAGE     COMMAND                  CREATED       STATUS         PORTS                               NAMES
    5c0c043fd3eb   mysql     "docker-entrypoint.s…"   5 weeks ago   Up 8 seconds   0.0.0.0:3306->3306/tcp, 33060/tcp   bank-app-mysql

    # use the same password used in the docker run command 
    $ docker exec -it bank-app-mysql /bin/bash
    bash-4.4# mysql -u root -p
    Enter password:
    Welcome to the MySQL monitor.  Commands end with ; or \g.
    Your MySQL connection id is 8
    Server version: 8.2.0 MySQL Community Server - GPL

    Copyright (c) 2000, 2023, Oracle and/or its affiliates.

    Oracle is a registered trademark of Oracle Corporation and/or its
    affiliates. Other names may be trademarks of their respective
    owners.

    Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

    mysql> CREATE SCHEMA `bankOfAmerica`;
    Query OK, 1 row affected (0.03 sec)

    mysql> USE `bankOfAmerica`;
    Database changed

    mysql> CREATE TABLE `accounts` (
        ->     `id` INT AUTO_INCREMENT PRIMARY KEY,
        ->     `name` VARCHAR(255),
        ->     `balance` DECIMAL(10, 2),
        ->     `card_number` VARCHAR(16),
        ->     `is_card_active` BOOLEAN
        -> );
    Query OK, 0 rows affected (0.04 sec)

    mysql> SELECT * FROM accounts;
    Empty set (0.04 sec)

    mysql> INSERT INTO `accounts` (`name`, `balance`, `card_number`, `is_card_active`)
        -> VALUES ('John Doe', 1000.00, '1234567890123456', true);
    Query OK, 1 row affected (0.02 sec)

    mysql> SELECT * FROM accounts;
    +----+----------+---------+------------------+----------------+
    | id | name     | balance | card_number      | is_card_active |
    +----+----------+---------+------------------+----------------+
    |  1 | John Doe | 1000.00 | 1234567890123456 |              1 |
    +----+----------+---------+------------------+----------------+
    1 row in set (0.00 sec)
    ```
    Use the similar way to operate on the bank records directly from the DB.

- Run the bank application 
    ```bash
    $ cd bank-app
    $ go mod tidy
    $ go build 
    $ ./bank-app
    ```

    When the bank-app starts running, to manipulate the database through the bank-app use the ```bank-application-api.postman_collection.json``` file.

---
### API Documentation

The API server provides endpoints to perform CRUD operations on user accounts. The Base URL
for the API is http://127.0.0.1:8001 with no authentication needed.

#### 1. Get All Accounts

- **Endpoint:** `/account`
- **Method:** `GET`
- **Description:** Retrieve a list of all user accounts.
- **Example Request:**
  ```http
  GET http://127.0.0.1:8001/account
  ```

#### 2. Get Account by ID

- **Endpoint:** `/account/{id}`
- **Method:** `GET`
- **Description:** Retrieve details of a specific user account by providing its unique identifier.
- **Parameters:**
  - `{id}`: The unique identifier of the account.
- **Example Request:**
  ```http
  GET http://127.0.0.1:8001/account/123
  ```

#### 3. Create Account

- **Endpoint:** `/account`
- **Method:** `POST`
- **Description:** Create a new user account.
- **Request Body:**
  - JSON payload containing account details.
- **Example Request:**
  ```http
  POST http://127.0.0.1:8001/account
  Content-Type: application/json

  {
    "name": "Cameron Dias",
    "balance": 35000.00,
    "cardnumber": "267864311444",
    "iscardactive": true
  }
  ```

#### 4. Update Account

- **Endpoint:** `/account/{id}`
- **Method:** `PUT`
- **Description:** Update an existing user account by providing its unique identifier.
- **Parameters:**
  - `{id}`: The unique identifier of the account.
- **Request Body:**
  - JSON payload containing updated account details.
- **Example Request:**
  ```http
  PUT http://127.0.0.1:8001/account/123
  Content-Type: application/json

  {
    "name": "Cameron Dias",
    "balance": 15000.00,
    "cardnumber": "267864311444",
    "iscardactive": true
  }
  ```

#### 5. Delete Account

- **Endpoint:** `/account/{id}`
- **Method:** `DELETE`
- **Description:** Delete an existing user account by providing its unique identifier.
- **Parameters:**
  - `{id}`: The unique identifier of the account.
- **Example Request:**
  ```http
  DELETE http://127.0.0.1:8001/account/123
  ```