-- +goose Up
create table if not exists wishlist(user_id uuid references authentication(id) on delete cascade, symbol text);

-- +goose Down
drop table wishlist;
