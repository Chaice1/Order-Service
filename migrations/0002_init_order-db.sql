-- +goose Up
CREATE TABLE IF NOT EXISTS goods(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name_of_good TEXT NOT NULL,
    count INT NOT NULL,
    price DECIMAL(10,2) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
CREATE TABLE IF NOT EXISTS orders(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    total_price DECIMAL(10,2) NOT NULL,
    status TEXT NOT NULL DEFAULT 'CREATED',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS order_items(
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    good_id UUID NOT NULL REFERENCES goods(id) ON DELETE CASCADE,
    name_of_good TEXT NOT NULL, 
    count INT NOT NULL,
    price DECIMAL(10,2) NOT NULL
);

INSERT INTO goods (name_of_good,count,price) VALUES ('snickers',20,403);
INSERT INTO goods (name_of_good,count,price) VALUES ('cap',10,100);
INSERT INTO goods (name_of_good,count,price) VALUES ('t-shirt',25,800);
INSERT INTO goods (name_of_good,count,price) VALUES ('coat',3,10000);

CREATE INDEX IF NOT EXISTS idx_order_items ON order_items(order_id);



-- +goose Down
DROP TABLE IF EXISTS order_items;
DROP TABLE IF EXISTS orders;
DROP TABLE IF EXISTS goods;
