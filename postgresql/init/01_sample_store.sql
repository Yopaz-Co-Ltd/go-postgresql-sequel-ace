create schema if not exists app;

create table app.customers (
  id bigserial primary key,
  full_name text not null,
  email text not null unique,
  city text not null,
  created_at timestamptz not null default now()
);

create table app.products (
  id bigserial primary key,
  sku text not null unique,
  name text not null,
  category text not null,
  price numeric(12, 2) not null,
  active boolean not null default true
);

create table app.orders (
  id bigserial primary key,
  customer_id bigint not null references app.customers(id),
  status text not null check (status in ('draft', 'paid', 'shipped', 'cancelled')),
  ordered_at timestamptz not null default now()
);

create table app.order_items (
  id bigserial primary key,
  order_id bigint not null references app.orders(id) on delete cascade,
  product_id bigint not null references app.products(id),
  quantity integer not null check (quantity > 0),
  unit_price numeric(12, 2) not null
);

insert into app.customers (full_name, email, city, created_at) values
  ('Nguyen Minh Anh', 'minh.anh@example.com', 'Ho Chi Minh City', now() - interval '12 days'),
  ('Tran Bao Long', 'bao.long@example.com', 'Da Nang', now() - interval '9 days'),
  ('Le Thu Ha', 'thu.ha@example.com', 'Ha Noi', now() - interval '7 days'),
  ('Pham Quoc Viet', 'quoc.viet@example.com', 'Can Tho', now() - interval '3 days');

insert into app.products (sku, name, category, price, active) values
  ('KB-MECH-01', 'Mechanical Keyboard', 'Accessories', 89.90, true),
  ('MS-WL-02', 'Wireless Mouse', 'Accessories', 34.50, true),
  ('MN-27-4K', '27 inch 4K Monitor', 'Displays', 319.00, true),
  ('USB-C-HUB', 'USB-C Hub 8-in-1', 'Adapters', 49.00, true),
  ('HDMI-2M', 'HDMI Cable 2m', 'Cables', 9.99, false);

insert into app.orders (customer_id, status, ordered_at) values
  (1, 'paid', now() - interval '10 days'),
  (1, 'shipped', now() - interval '8 days'),
  (2, 'paid', now() - interval '5 days'),
  (3, 'draft', now() - interval '2 days');

insert into app.order_items (order_id, product_id, quantity, unit_price) values
  (1, 1, 1, 89.90),
  (1, 2, 1, 34.50),
  (2, 3, 1, 319.00),
  (3, 4, 2, 49.00),
  (4, 2, 1, 34.50),
  (4, 5, 3, 9.99);

create view app.order_summary as
select
  o.id as order_id,
  c.full_name as customer,
  c.city,
  o.status,
  o.ordered_at,
  sum(oi.quantity * oi.unit_price) as total_amount,
  count(*) as item_count
from app.orders o
join app.customers c on c.id = o.customer_id
join app.order_items oi on oi.order_id = o.id
group by o.id, c.full_name, c.city, o.status, o.ordered_at
order by o.ordered_at desc;
