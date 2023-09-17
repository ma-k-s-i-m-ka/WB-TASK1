DROP TABLE IF EXISTS "order";
DROP TABLE IF EXISTS Delivery;
DROP TABLE IF EXISTS Payment;
DROP TABLE IF EXISTS Item;

CREATE TABLE IF NOT EXISTS Delivery (
 id         serial primary key,
 name       text,
 phone      text,
 zip        text,
 city       text,
 address    text,
 region     text,
 email      text
);


CREATE TABLE IF NOT EXISTS Payment (
 id              serial primary key,
 transaction     text,
 request_id      text,
 currency        text,
 provider        text,
 amount          int,
 payment_dt      bigint,
 bank            text,
 delivery_cost   int,
 goods_total     int,
 custom_fee      int
);

CREATE TABLE IF NOT EXISTS Item (
 id              serial primary key,
 chrt_id         int,
 track_number    text,
 price           int,
 rid             text,
 name            text,
 sale            int,
 size            text,
 total_price     int,
 nm_id           int,
 brand           text,
 status          int
);

CREATE TABLE IF NOT EXISTS "order" (
 order_uid          text,
 track_number       text,
 entry              text,
 delivery           bigint,
 payment            bigint,
 items              bigint[],
 locale             text,
 internal_signature text,
 customer_id        text,
 delivery_service   text,
 shardkey           text,
 sm_id              int,
 date_created       text,
 oof_shard          text,

 foreign key(delivery) references Delivery(id) on delete cascade,
 foreign key(payment) references Payment(id) on delete cascade
);
