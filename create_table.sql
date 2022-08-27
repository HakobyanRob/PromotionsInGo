create table if not exists promotions
(
    id              text          not null        constraint promotions_pk primary key,
    price           numeric(9, 6) not null,
    expiration_date text          not null
);