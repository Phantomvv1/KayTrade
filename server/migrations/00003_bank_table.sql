-- +goose Up
create table if not exists bank(id uuid primary key, user_id uuid references authentication(id) on delete cascade,
type text, created_at timestamp default current_timestamp, updated_at timestamp default current_timestamp);

-- +goose Down
drop table bank;
