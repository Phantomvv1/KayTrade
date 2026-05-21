-- +goose Up
create table if not exists r_tokens(token uuid primary key default gen_random_uuid(),
user_id uuid references authentication(id) on delete cascade, expiration timestamp default current_timestamp + '5 days'::interval, valid bool);

-- +goose Down
drop table r_tokens;
