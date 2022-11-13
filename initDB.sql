CREATE TABLE IF NOT EXISTS servicies (
     id bigserial primary key not null,
     name varchar(60) not null unique
    );

CREATE TABLE IF NOT EXISTS user_accounts (
    user_id bigserial primary key not null,
    balance integer not null CHECK (balance >= 0),
    reserved_balance integer not null CHECK (reserved_balance >= 0)
    );

CREATE TABLE IF NOT EXISTS transactions (
    id  bigserial primary key not null,
    user_id integer REFERENCES user_accounts (user_id) not null,
    amount integer not null	,
    description text,
    order_id integer,
    service_id integer REFERENCES servicies (id),
    closed_date date,
    success_flg boolean not null default false,
    type varchar(30) not null CHECK (type in ('add', 'reserve')),

    UNIQUE (user_id,amount,order_id,service_id)
    );

INSERT INTO servicies (id, name) VALUES (1, 'услуга 1');
INSERT INTO servicies (id, name) VALUES (2, 'услуга 2');