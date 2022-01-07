CREATE TABLE UserBalance (
    id BIGINT NOT NULL,
    balance BIGINT NOT NULL -- balance is stored in pennies (kopeks?) and then converted to roubles
);

CREATE TABLE UserTransfers (
    id BIGINT NOT NULL,
    amount BIGINT NOT NULL,
    transferred_at TIMESTAMP,
    purpose TEXT
);