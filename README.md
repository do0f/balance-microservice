# balance-microservice
Microservice written in go for interacting with user balance using REST API

This microservice can interact with users' balance: getting balance in different currencies, transferring money to other users and change balance explicitly for external money transfers.

This project uses:
- Echo framework for REST API server
- pgx driver for interacting with PostgreSQL through sql package
- testify and gomock for implementing unit tests

Each user's balance and transfer history is stored in PostgreSQL database. Dockerfile for database and schema for creating empty tables are stored in /database folder.
User's balance record is created with first money crediting.

Balance is stored in kopeks to avoid loss of precision and then is converted to primary value (RUB) and secondary value (kopeks).

format of JSONs returned by api:
- balance
{
    "primary value": integer,
    "secondary value": integer,
    "currency": string
}

- transfer
 {
    "primary value": integer,
    "secondary value": integer,
    "transferred at": timestamp,
    "purpose": string
 }
 
 format of JSON required by api:
 - changing balance
 {
    "amount": integer
 }
 - transferring
 {
    "recipient": integer,
    "amount": integer
 }
  
API methods:
- Getting balance:
  - path: /balance/users/{id}?currency={currency} (currency can be omited)
  - output: JSON in "balance" format
  - example:
      - path: localhost:1323/balance/users/1
      - output:  
      {
        "primary value": 1000,
        "secondary value": 50,
        "currency": "RUB"
      }

- Getting history:
  - path: /balance/users/{id}/history
  - output: array of JSONs in "transfer" format
  - example:
      - path: localhost:1323/balance/users/1/history
      - output:  
      [
    {
        "primary value": 10000,
        "secondary value": 0,
        "transferred at": "2022-01-05T17:54:24.708493Z",
        "purpose": "External service operation"
    },
    {
        "primary value": 10000,
        "secondary value": 0,
        "transferred at": "2022-01-07T12:36:20.923268Z",
        "purpose": "External service operation"
    }
    ]
    
- Changing balance:
  - path: /balance/users/{id}
  - input: JSON in "changing balance" format
  - output: JSON in "balance" format
    - example:
      - path: localhost:1323/balance/users/1
      - input:
      {
        "amount": 10000
      }
      - output:  
      {
        "primary value": 1000,
        "secondary value": 50,
        "currency": "RUB"
      }
  
- Transferring:
  - path: /balance/users/{id}/transfer
  - input: JSON in "transferring" format
  - output: JSON in "balance" format with sender's balance
    - example:
      - path: localhost:1323/balance/users/1/transfer
      - input:
      {
        "recipient": 2,
        "amount": 10000
      }
      - output:  
      {
        "primary value": 1000,
        "secondary value": 50,
        "currency": "RUB"
      }
  
  
  
  
  
  
  
  
  
  
  
  
  
 
