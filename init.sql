CREATE TABLE IF NOT EXISTS product (
    id SERIAL PRIMARY KEY,
    product_name VARCHAR(255) NOT NULL,
    price DECIMAL(10, 2) NOT NULL
);

INSERT INTO product (product_name, price) VALUES ('Playstation', 3.500);
INSERT INTO product (product_name, price) VALUES ('Fone', 25.5);
INSERT INTO product (product_name, price) VALUES ('Carregador', 30.99);
