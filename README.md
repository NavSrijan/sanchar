# sanchar
A simple server and client to implement chat.

## Installation

1. Ensure you have Go and Python installed on your machine.
2. Clone the repository.
3. Navigate to the project directory.

## Usage

### Make sure to setup the Database before running the server
The server uses a PostgreSQL database to store user information. To setup the database, look at Database.md.

```sh

### Server

To start the server, navigate to the project directory and run the following command:

```sh
go run main.go
```

### Client

To start the client, navigate to the client directory and run the following command:

```sh
python3 client.py
```

When prompted, enter your username, password and the reciepent's username.
