CREATE TABLE IF NOT EXISTS users (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    role ENUM('landlord', 'tenant') NOT NULL,
    display_name VARCHAR(80) NOT NULL,
    email VARCHAR(190) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uk_users_email (email)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS houses (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    landlord_id BIGINT UNSIGNED NOT NULL,
    title VARCHAR(160) NOT NULL,
    description TEXT NOT NULL,
    city VARCHAR(80) NOT NULL,
    district VARCHAR(80) NOT NULL,
    address VARCHAR(255) NOT NULL,
    monthly_rent INT UNSIGNED NOT NULL,
    bedrooms TINYINT UNSIGNED NOT NULL,
    bathrooms TINYINT UNSIGNED NOT NULL DEFAULT 1,
    area_sqm DECIMAL(8, 2) UNSIGNED NOT NULL DEFAULT 0,
    amenities JSON NOT NULL,
    image_urls JSON NOT NULL,
    status ENUM('draft', 'active', 'rented', 'archived') NOT NULL DEFAULT 'active',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    KEY idx_houses_search (city, district, monthly_rent, bedrooms, status),
    FULLTEXT KEY ft_houses_content (title, description, address),
    CONSTRAINT fk_houses_landlord FOREIGN KEY (landlord_id) REFERENCES users (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS favorites (
    tenant_id BIGINT UNSIGNED NOT NULL,
    house_id BIGINT UNSIGNED NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (tenant_id, house_id),
    CONSTRAINT fk_favorites_tenant FOREIGN KEY (tenant_id) REFERENCES users (id),
    CONSTRAINT fk_favorites_house FOREIGN KEY (house_id) REFERENCES houses (id) ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE TABLE IF NOT EXISTS messages (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    house_id BIGINT UNSIGNED NOT NULL,
    sender_id BIGINT UNSIGNED NOT NULL,
    content VARCHAR(1000) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    KEY idx_messages_house_created (house_id, created_at),
    CONSTRAINT fk_messages_house FOREIGN KEY (house_id) REFERENCES houses (id) ON DELETE CASCADE,
    CONSTRAINT fk_messages_sender FOREIGN KEY (sender_id) REFERENCES users (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

INSERT INTO users (id, role, display_name, email) VALUES
    (1, 'landlord', '林先生', 'landlord@example.com'),
    (2, 'tenant', '陈同学', 'tenant@example.com')
ON DUPLICATE KEY UPDATE display_name = VALUES(display_name);

INSERT INTO houses (
    id, landlord_id, title, description, city, district, address,
    monthly_rent, bedrooms, bathrooms, area_sqm, amenities, image_urls, status
) VALUES
    (
        1, 1, '徐汇滨江明亮两居', '朝南客厅，步行可达地铁站，适合两人合租或小家庭。',
        '上海', '徐汇区', '龙腾大道附近', 6200, 2, 1, 68.00,
        JSON_ARRAY('近地铁', '电梯', '可做饭', '独立阳台'),
        JSON_ARRAY('https://images.unsplash.com/photo-1522708323590-d24dbb6b0267d?auto=format&fit=crop&w=1200&q=80'),
        'active'
    ),
    (
        2, 1, '张江园区精装一居', '带独立书桌和充足收纳，通勤张江科学城方便。',
        '上海', '浦东新区', '张江路附近', 5100, 1, 1, 46.00,
        JSON_ARRAY('近地铁', '精装修', '智能门锁', '可养猫'),
        JSON_ARRAY('https://images.unsplash.com/photo-1505693416388-ac5ce068fe85?auto=format&fit=crop&w=1200&q=80'),
        'active'
    ),
    (
        3, 1, '静安老洋房安静单间', '安静内街，采光好，公共区域定期保洁。',
        '上海', '静安区', '愚园路附近', 3900, 1, 1, 28.00,
        JSON_ARRAY('市中心', '定期保洁', '拎包入住'),
        JSON_ARRAY('https://images.unsplash.com/photo-1560448204-e02f11c3d0e2?auto=format&fit=crop&w=1200&q=80'),
        'active'
    )
ON DUPLICATE KEY UPDATE title = VALUES(title);
