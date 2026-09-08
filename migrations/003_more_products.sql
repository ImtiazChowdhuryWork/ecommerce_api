-- Extra product catalog for local practice / testing.
-- Safe to re-run: each product name is unique per this file and guarded with NOT EXISTS.
-- Categories referenced here come from 002_seed.sql (Electronics, Clothing, Books, Home & Garden, Sports).

INSERT INTO products (name, description, price, stock, category_id, image_url)
SELECT v.name, v.description, v.price, v.stock, c.id, v.image_url
FROM (
    VALUES
    -- Electronics
    ('4K Action Camera',        'Waterproof 4K/60fps action camera with stabilisation', 199.99, 40, 'Electronics', 'https://picsum.photos/seed/actioncam/600'),
    ('Bluetooth Speaker',       'Portable 20W speaker, 24-hour battery, IPX7',          59.99,  120, 'Electronics', 'https://picsum.photos/seed/speaker/600'),
    ('USB-C 7-in-1 Hub',        'HDMI, SD, 3x USB-A, USB-C PD, ethernet',               39.99,  85,  'Electronics', 'https://picsum.photos/seed/hub/600'),
    ('Wireless Mouse',          'Silent-click ergonomic wireless mouse, 2.4GHz + BT',   24.99,  200, 'Electronics', 'https://picsum.photos/seed/mouse/600'),
    ('27-inch 1440p Monitor',   'IPS 165Hz gaming monitor, HDR400, FreeSync',           289.99, 25,  'Electronics', 'https://picsum.photos/seed/monitor/600'),
    ('Smartwatch Series 6',     'AMOLED, GPS, heart-rate + SpO2, 7-day battery',        149.99, 60,  'Electronics', 'https://picsum.photos/seed/smartwatch/600'),
    ('Power Bank 20000mAh',     '65W USB-C PD power bank, laptop-capable',              49.99,  150, 'Electronics', 'https://picsum.photos/seed/powerbank/600'),
    ('Webcam 1080p',            'Full HD webcam with dual mics and privacy shutter',    34.99,  90,  'Electronics', 'https://picsum.photos/seed/webcam/600'),

    -- Clothing
    ('Classic Denim Jacket',    'Unisex mid-wash denim jacket, 100% cotton',            64.99,  70,  'Clothing', 'https://picsum.photos/seed/denim/600'),
    ('Merino Wool Socks 3-Pack','Cushioned merino hiking socks, temperature regulating', 27.99, 180, 'Clothing', 'https://picsum.photos/seed/socks/600'),
    ('Cotton Crew T-Shirt',     'Pre-shrunk heavyweight crew neck tee',                 19.99,  300, 'Clothing', 'https://picsum.photos/seed/tshirt/600'),
    ('Rain Shell Jacket',       'Packable 2.5-layer waterproof shell, taped seams',     89.99,  55,  'Clothing', 'https://picsum.photos/seed/rainshell/600'),
    ('Leather Belt',            'Full-grain leather belt with matte buckle',            34.99,  110, 'Clothing', 'https://picsum.photos/seed/belt/600'),
    ('Beanie Hat',              'Ribbed acrylic-wool blend beanie',                     16.99,  240, 'Clothing', 'https://picsum.photos/seed/beanie/600'),

    -- Books
    ('Designing Data-Intensive Applications', 'Martin Kleppmann — the definitive systems book', 44.99, 40, 'Books', 'https://picsum.photos/seed/ddia/600'),
    ('Clean Architecture',      'Robert C. Martin on software structure and boundaries', 32.99, 50,  'Books', 'https://picsum.photos/seed/cleanarch/600'),
    ('The Pragmatic Programmer','20th Anniversary Edition — Hunt & Thomas',             39.99,  45,  'Books', 'https://picsum.photos/seed/pragprog/600'),
    ('SQL Performance Explained','Markus Winand — indexing and query tuning',           29.99,  35,  'Books', 'https://picsum.photos/seed/sqlperf/600'),
    ('Go in Action',            'Practical guide to building Go applications',          34.99,  30,  'Books', 'https://picsum.photos/seed/goinaction/600'),

    -- Home & Garden
    ('Cast Iron Skillet 12"',   'Pre-seasoned cast iron skillet, oven-safe',            34.99,  95,  'Home & Garden', 'https://picsum.photos/seed/skillet/600'),
    ('LED Desk Lamp',           'Dimmable desk lamp, 5 colour temps, USB charging port', 29.99, 130, 'Home & Garden', 'https://picsum.photos/seed/desklamp/600'),
    ('Stainless Cookware Set',  '10-piece tri-ply stainless steel cookware set',        219.99, 20,  'Home & Garden', 'https://picsum.photos/seed/cookware/600'),
    ('Ceramic Planter Set',     'Set of 3 mid-century ceramic planters with trays',     42.99,  75,  'Home & Garden', 'https://picsum.photos/seed/planter/600'),
    ('Pruning Shears',          'Bypass pruning shears, titanium-coated blades',        24.99,  140, 'Home & Garden', 'https://picsum.photos/seed/shears/600'),
    ('Weighted Blanket 15lb',   'Cooling weighted blanket, glass beads, washable cover', 69.99, 50,  'Home & Garden', 'https://picsum.photos/seed/blanket/600'),

    -- Sports
    ('Adjustable Dumbbell 25kg','Single adjustable dumbbell, 2.5-25kg dial',            179.99, 30,  'Sports', 'https://picsum.photos/seed/dumbbell/600'),
    ('Resistance Bands Set',    '5-band set with door anchor and handles',              22.99,  220, 'Sports', 'https://picsum.photos/seed/bands/600'),
    ('Insulated Water Bottle',  '1L vacuum-insulated stainless bottle, 24h cold',       28.99,  260, 'Sports', 'https://picsum.photos/seed/bottle/600'),
    ('Foam Roller',             'High-density textured foam roller, 33cm',              21.99,  150, 'Sports', 'https://picsum.photos/seed/foamroller/600'),
    ('Jump Rope Pro',           'Ball-bearing speed rope with adjustable cable',        14.99,  300, 'Sports', 'https://picsum.photos/seed/jumprope/600'),
    ('Trail Running Backpack',  '12L hydration-compatible running vest',                74.99,  45,  'Sports', 'https://picsum.photos/seed/runvest/600')
) AS v(name, description, price, stock, category_name, image_url)
JOIN categories c ON c.name = v.category_name
WHERE NOT EXISTS (SELECT 1 FROM products p WHERE p.name = v.name);
