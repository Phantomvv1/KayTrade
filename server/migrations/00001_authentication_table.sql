-- +goose Up
create table if not exists authentication (id uuid primary key, full_name text, 
email text, password text, type int check (type in (1, 2)), 
created_at timestamp default current_timestamp, updated_at timestamp default current_timestamp);

-- +goose Down
drop table authentication;
