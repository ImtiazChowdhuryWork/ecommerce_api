-- Seed an admin user (password: Admin@123)
INSERT INTO users (email, password_hash, name, role) VALUES
  ('admin@example.com', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', 'Admin User', 'admin')
ON CONFLICT (email) DO NOTHING;

-- Seed categories
INSERT INTO categories (name, description) VALUES
  ('Electronics', 'Gadgets, devices, and electronic components'),
  ('Clothing',    'Apparel, footwear, and accessories'),
  ('Books',       'Fiction, non-fiction, textbooks, and more'),
  ('Home & Garden', 'Furniture, decor, and gardening tools'),
  ('Sports',      'Equipment and gear for sports and fitness')
ON CONFLICT (name) DO NOTHING;

-- Seed products (requires category IDs — uses subqueries)
INSERT INTO products (name, description, price, stock, category_id, image_url)
SELECT 'Wireless Headphones', 'High-quality noise-cancelling wireless headphones', 149.99, 50,
       id, 'https://example.com/images/headphones.jpg' FROM categories WHERE name = 'Electronics';

INSERT INTO products (name, description, price, stock, category_id, image_url)
SELECT 'Mechanical Keyboard', 'RGB mechanical keyboard with Cherry MX switches', 89.99, 30,
       id, 'https://example.com/images/keyboard.jpg' FROM categories WHERE name = 'Electronics';

INSERT INTO products (name, description, price, stock, category_id, image_url)
SELECT 'Running Shoes', 'Lightweight running shoes with cushioned sole', 79.99, 100,
       id, 'https://example.com/images/shoes.jpg' FROM categories WHERE name = 'Clothing';

INSERT INTO products (name, description, price, stock, category_id, image_url)
SELECT 'The Go Programming Language', 'Comprehensive guide to Go by Donovan & Kernighan', 39.99, 25,
       id, 'https://example.com/images/gobook.jpg' FROM categories WHERE name = 'Books';

INSERT INTO products (name, description, price, stock, category_id, image_url)
SELECT 'Yoga Mat', 'Non-slip premium yoga mat, 6mm thick', 34.99, 75,
       id, 'https://example.com/images/yogamat.jpg' FROM categories WHERE name = 'Sports';
