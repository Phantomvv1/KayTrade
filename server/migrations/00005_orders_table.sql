-- +goose Up
create table if not exists orders(id uuid primary key, user_id uuid references authentication(id) on delete cascade, 
symbol text, side text, created_at timestamp, updated_at timestamp);

-- +goose Down
drop table orders;
