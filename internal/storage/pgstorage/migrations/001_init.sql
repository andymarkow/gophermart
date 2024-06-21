-- +goose Up

--
-- Users
--
CREATE TABLE IF NOT EXISTS users (
    "login" VARCHAR (50) UNIQUE NOT NULL,
    "password_hash" VARCHAR (100) NOT NULL
);

CREATE INDEX IF NOT EXISTS users_login_idx ON users (login);

--
-- User Balance
--
CREATE TABLE IF NOT EXISTS user_balance (
    "login" VARCHAR (50) UNIQUE NOT NULL,
    "current" DOUBLE PRECISION NOT NULL DEFAULT 0,
    "withdrawn" DOUBLE PRECISION NOT NULL DEFAULT 0
);

CREATE INDEX IF NOT EXISTS user_balance_login_idx ON user_balance (login);

--
-- User Withdrawals
--
CREATE TABLE IF NOT EXISTS user_withdrawals (
    "order_id" VARCHAR (32) UNIQUE NOT NULL,
    "login" VARCHAR (50) NOT NULL,
    "amount" DOUBLE PRECISION NOT NULL,
    "processed_at" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS user_withdrawals_order_id_idx ON user_withdrawals (order_id);

--
-- Orders
--
CREATE TABLE IF NOT EXISTS orders (
    "id" VARCHAR (32) UNIQUE NOT NULL,
    "user_login" VARCHAR (50) NOT NULL,
    "status" VARCHAR (32) NOT NULL,
    "accrual" DOUBLE PRECISION NOT NULL DEFAULT 0,
    "uploaded_at" TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS orders_id_idx ON orders (id);


-- +goose Down
DROP INDEX users_login_idx;
DROP TABLE users;

DROP INDEX user_balance_login_idx;
DROP TABLE user_balance;

DROP INDEX user_withdrawals_order_id_idx;
DROP TABLE user_withdrawals;

DROP INDEX orders_id_idx;
DROP TABLE orders;
