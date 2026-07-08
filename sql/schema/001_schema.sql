-- +goose Up
CREATE TABLE users (
    id UUID PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    email VARCHAR(100) UNIQUE NOT NULL CHECK(char_length(email) BETWEEN 1 AND 100),
    hashed_password TEXT NOT NULL DEFAULT 'unset'
);

CREATE TABLE refresh_tokens (
    token TEXT PRIMARY KEY,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    user_id UUID NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ,
    CONSTRAINT fk_users FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE categories (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE products (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    price_usd NUMERIC(10, 2) NOT NULL,
    quantity INTEGER NOT NULL DEFAULT 0,
    category_id INT, 
    CONSTRAINT fk_category FOREIGN KEY (category_id) REFERENCES categories(id) ON DELETE SET NULL
);

CREATE TABLE shopping_cart (
    id SERIAL PRIMARY KEY,
    user_id UUID NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT fk_cart_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE cart_items (
    id SERIAL PRIMARY KEY,
    cart_id INT REFERENCES shopping_cart(id) ON DELETE CASCADE,
    product_id INT REFERENCES products(id) ON DELETE CASCADE,
    quantity INT DEFAULT 1 CHECK (quantity > 0), 
    added_at TIMESTAMPTZ DEFAULT NOW(), 
    UNIQUE (cart_id, product_id) 
);

CREATE TYPE payment_status AS ENUM (
    'pending',    
    'processing', 
    'succeeded',  
    'failed',     
    'canceled'    
);

CREATE TABLE payments (
    id SERIAL PRIMARY KEY,
    user_id UUID NOT NULL,  
    cart_id INT, 
    stripe_payment_intent_id VARCHAR(255) UNIQUE, 
    amount INT NOT NULL, 
    currency VARCHAR(3) DEFAULT 'usd',
    status payment_status DEFAULT 'pending',
    error_message TEXT, 
    created_at TIMESTAMPTZ DEFAULT NOW(),  
    updated_at TIMESTAMPTZ DEFAULT NOW(),
    CONSTRAINT fk_payment_user FOREIGN KEY (user_id) REFERENCES users(id),
    CONSTRAINT fk_payment_cart FOREIGN KEY (cart_id) REFERENCES shopping_cart(id) ON DELETE SET NULL
);

-- +goose Down
DROP TABLE IF EXISTS payments;
DROP TABLE IF EXISTS cart_items;
DROP TABLE IF EXISTS shopping_cart;
DROP TABLE IF EXISTS products;
DROP TABLE IF EXISTS categories;
DROP TABLE IF EXISTS refresh_tokens;
DROP TABLE IF EXISTS users;
DROP TYPE IF EXISTS payment_status; 
