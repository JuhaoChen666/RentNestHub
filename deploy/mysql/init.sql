SET NAMES utf8mb4;

CREATE TABLE IF NOT EXISTS users (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    role ENUM('admin', 'landlord', 'tenant') NOT NULL,
    username VARCHAR(80) NOT NULL,
    display_name VARCHAR(80) NOT NULL,
    email VARCHAR(190) NOT NULL,
    password_hash VARCHAR(100) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uk_users_username (username),
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
    recipient_id BIGINT UNSIGNED NOT NULL,
    content VARCHAR(1000) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    KEY idx_messages_house_created (house_id, created_at),
    KEY idx_messages_participant_created (sender_id, recipient_id, created_at),
    CONSTRAINT fk_messages_house FOREIGN KEY (house_id) REFERENCES houses (id) ON DELETE CASCADE,
    CONSTRAINT fk_messages_sender FOREIGN KEY (sender_id) REFERENCES users (id),
    CONSTRAINT fk_messages_recipient FOREIGN KEY (recipient_id) REFERENCES users (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

INSERT INTO users (id, role, username, display_name, email, password_hash) VALUES
    (1, 'landlord', 'landlord', '林先生', 'landlord@test.example.com', '$2a$10$QOpjNI71YtZPcto6Ez4/6OjEQ3AVuXpRXxr7496GvCmWp5V4dNaUa'),
    (2, 'admin', 'admin', '管理员', 'admin@test.example.com', '$2a$10$QOpjNI71YtZPcto6Ez4/6OjEQ3AVuXpRXxr7496GvCmWp5V4dNaUa'),
    (3, 'tenant', 'tenant', '普通用户', 'tenant@test.example.com', '$2a$10$QOpjNI71YtZPcto6Ez4/6OjEQ3AVuXpRXxr7496GvCmWp5V4dNaUa')
ON DUPLICATE KEY UPDATE
    role = VALUES(role),
    username = VALUES(username),
    display_name = VALUES(display_name),
    email = VALUES(email),
    password_hash = VALUES(password_hash);

INSERT INTO houses (
    id, landlord_id, title, description, city, district, address,
    monthly_rent, bedrooms, bathrooms, area_sqm, amenities, image_urls, status
) VALUES
    (
        1, 2, '徐汇滨江明亮两居', '朝南客厅，步行可达地铁站，适合两人合租或小家庭。',
        '上海', '徐汇区', '龙腾大道附近', 6200, 2, 1, 68.00,
        JSON_ARRAY('近地铁', '电梯', '可做饭', '独立阳台'),
        JSON_ARRAY('https://images.unsplash.com/photo-1522708323590-d24dbb6b0267d?auto=format&fit=crop&w=1200&q=80'),
        'active'
    ),
    (
        2, 2, '张江园区精装一居', '带独立书桌和充足收纳，通勤张江科学城方便。',
        '上海', '浦东新区', '张江路附近', 5100, 1, 1, 46.00,
        JSON_ARRAY('近地铁', '精装修', '智能门锁', '可养猫'),
        JSON_ARRAY('https://images.unsplash.com/photo-1505693416388-ac5ce068fe85?auto=format&fit=crop&w=1200&q=80'),
        'active'
    ),
    (
        3, 2, '静安老洋房安静单间', '安静内街，采光好，公共区域定期保洁。',
        '上海', '静安区', '愚园路附近', 3900, 1, 1, 28.00,
        JSON_ARRAY('市中心', '定期保洁', '拎包入住'),
        JSON_ARRAY('https://images.unsplash.com/photo-1560448204-e02f11c3d0e2?auto=format&fit=crop&w=1200&q=80'),
        'active'
    ),
    (4, 2, '陆家嘴江景一居', '高层采光充足，步行可达地铁与滨江步道，适合陆家嘴通勤。', '上海', '浦东新区', '浦东南路附近', 8200, 1, 1, 52.00, JSON_ARRAY('近地铁', '江景', '电梯', '独立阳台'), JSON_ARRAY('https://images.unsplash.com/photo-1600566753086-00f18fb6b3ea?auto=format&fit=crop&w=1200&q=80'), 'active'),
    (5, 2, '张江科学城阳光一居', '独立书房区和充足收纳，通勤园区方便。', '上海', '浦东新区', '松涛路附近', 5600, 1, 1, 48.00, JSON_ARRAY('近地铁', '可做饭', '智能门锁', '精装修'), JSON_ARRAY('https://images.unsplash.com/photo-1600210492486-724fe5c67fb0?auto=format&fit=crop&w=1200&q=80'), 'active'),
    (6, 2, '世纪公园静谧两居', '南北通透，步行可到世纪公园，客厅空间开阔。', '上海', '浦东新区', '锦绣路附近', 9800, 2, 1, 76.00, JSON_ARRAY('近公园', '电梯', '可做饭', '独立阳台'), JSON_ARRAY('https://images.unsplash.com/photo-1600585154340-be6161a56a0c?auto=format&fit=crop&w=1200&q=80'), 'active'),
    (7, 2, '金桥品质一居', '小区环境安静，精装交付，适合金桥办公区通勤。', '上海', '浦东新区', '金桥路附近', 5300, 1, 1, 45.00, JSON_ARRAY('精装修', '电梯', '可做饭', '门禁'), JSON_ARRAY('https://images.unsplash.com/photo-1600607687920-4e2a09cf159d?auto=format&fit=crop&w=1200&q=80'), 'active'),
    (8, 2, '北蔡地铁口两居', '双卧独立，楼下生活配套完善，家庭合租均合适。', '上海', '浦东新区', '沪南路附近', 7200, 2, 1, 66.00, JSON_ARRAY('近地铁', '电梯', '可做饭', '宠物友好'), JSON_ARRAY('https://images.unsplash.com/photo-1600573472591-ee6b68d14c68?auto=format&fit=crop&w=1200&q=80'), 'active'),
    (9, 2, '三林清爽一居', '全屋新装，厨房和卫浴独立，周边超市齐全。', '上海', '浦东新区', '三林路附近', 4400, 1, 1, 39.00, JSON_ARRAY('精装修', '可做饭', '独立卫浴', '拎包入住'), JSON_ARRAY('https://images.unsplash.com/photo-1600210491892-03d54c0aaf87?auto=format&fit=crop&w=1200&q=80'), 'active'),
    (10, 2, '前滩滨江两居', '客餐厅一体，步行可达商场和滨江，适合品质居住。', '上海', '浦东新区', '前滩大道附近', 11800, 2, 2, 88.00, JSON_ARRAY('近商场', '电梯', '中央空调', '独立阳台'), JSON_ARRAY('https://images.unsplash.com/photo-1613490493576-7fde63acd811?auto=format&fit=crop&w=1200&q=80'), 'active'),
    (11, 2, '康桥通勤一居', '租金友好，通勤张江与周浦方便，房间方正。', '上海', '浦东新区', '康桥路附近', 3900, 1, 1, 36.00, JSON_ARRAY('可做饭', '独立卫浴', '小区安静', '拎包入住'), JSON_ARRAY('https://images.unsplash.com/photo-1600607688969-a5bfcd646154?auto=format&fit=crop&w=1200&q=80'), 'active'),
    (12, 2, '周浦万达舒适两居', '卧室采光好，临近商圈和公交站，适合朋友合租。', '上海', '浦东新区', '年家浜路附近', 5800, 2, 1, 64.00, JSON_ARRAY('近商场', '可做饭', '电梯', '独立阳台'), JSON_ARRAY('https://images.unsplash.com/photo-1600585154526-990dced4db0d?auto=format&fit=crop&w=1200&q=80'), 'active'),
    (13, 2, '洋泾地铁旁一居', '近6号线，通勤陆家嘴便利，配有独立洗烘。', '上海', '浦东新区', '张杨路附近', 6100, 1, 1, 43.00, JSON_ARRAY('近地铁', '洗烘', '可做饭', '精装修'), JSON_ARRAY('https://images.unsplash.com/photo-1600566753190-17f0baa2a6c3?auto=format&fit=crop&w=1200&q=80'), 'active'),
    (14, 2, '塘桥温馨两居', '双卧朝南，社区成熟，周边有菜场和医院。', '上海', '浦东新区', '浦建路附近', 7600, 2, 1, 69.00, JSON_ARRAY('近地铁', '可做饭', '电梯', '小区安静'), JSON_ARRAY('https://images.unsplash.com/photo-1600585154363-67eb9e2e2099?auto=format&fit=crop&w=1200&q=80'), 'active'),
    (15, 2, '花木艺术馆一居', '采光客厅配独立阳台，适合在世纪公园附近上班的租客。', '上海', '浦东新区', '芳甸路附近', 6800, 1, 1, 50.00, JSON_ARRAY('近公园', '独立阳台', '电梯', '可做饭'), JSON_ARRAY('https://images.unsplash.com/photo-1618221195710-dd6b41faaea6?auto=format&fit=crop&w=1200&q=80'), 'active'),
    (16, 2, '御桥精装一居', '家具家电齐全，地铁换乘便利，入住成本低。', '上海', '浦东新区', '御桥路附近', 4700, 1, 1, 41.00, JSON_ARRAY('近地铁', '精装修', '拎包入住', '智能门锁'), JSON_ARRAY('https://images.unsplash.com/photo-1616486338812-3dadae4b4ace?auto=format&fit=crop&w=1200&q=80'), 'active'),
    (17, 2, '川沙古镇两居', '低密小区，生活节奏舒适，适合机场和迪士尼方向通勤。', '上海', '浦东新区', '川沙路附近', 4600, 2, 1, 62.00, JSON_ARRAY('可做饭', '电梯', '宠物友好', '小区安静'), JSON_ARRAY('https://images.unsplash.com/photo-1615874959474-d609969a20ed?auto=format&fit=crop&w=1200&q=80'), 'active'),
    (18, 2, '临港新城海风一居', '视野开阔，社区新，适合临港办公和远程办公。', '上海', '浦东新区', '环湖西一路附近', 3500, 1, 1, 42.00, JSON_ARRAY('新小区', '可做饭', '独立阳台', '停车方便'), JSON_ARRAY('https://images.unsplash.com/photo-1618220179428-22790b461013?auto=format&fit=crop&w=1200&q=80'), 'active'),
    (19, 2, '陆家嘴合租次卧', '房间独立带书桌，公共区域整洁，租金包含宽带。', '上海', '浦东新区', '福山路附近', 4300, 1, 1, 20.00, JSON_ARRAY('近地铁', '合租', '宽带', '定期保洁'), JSON_ARRAY('https://images.unsplash.com/photo-1554995207-c18c203602cb?auto=format&fit=crop&w=1200&q=80'), 'active'),
    (20, 2, '张江软件园两居', '主次卧分离，厨房设备完善，团队合租友好。', '上海', '浦东新区', '祖冲之路附近', 7500, 2, 1, 72.00, JSON_ARRAY('近地铁', '可做饭', '双卧独立', '电梯'), JSON_ARRAY('https://images.unsplash.com/photo-1618219740975-d40978bb7378?auto=format&fit=crop&w=1200&q=80'), 'active'),
    (21, 2, '金桥碧云三居', '三房两卫，客厅宽敞，适合家庭或稳定合租。', '上海', '浦东新区', '明月路附近', 12500, 3, 2, 108.00, JSON_ARRAY('三房', '两卫', '电梯', '独立阳台'), JSON_ARRAY('https://images.unsplash.com/photo-1600566753051-f0b89df2dd90?auto=format&fit=crop&w=1200&q=80'), 'active'),
    (22, 2, '北蔡新里一居', '房屋整洁，社区管理规范，步行可到生活广场。', '上海', '浦东新区', '莲溪路附近', 4200, 1, 1, 38.00, JSON_ARRAY('可做饭', '门禁', '电梯', '近商场'), JSON_ARRAY('https://images.unsplash.com/photo-1616594039964-ae9021a400a0?auto=format&fit=crop&w=1200&q=80'), 'active'),
    (23, 2, '三林前滩通勤两居', '通勤前滩约二十分钟，带储物间和双阳台。', '上海', '浦东新区', '上南路附近', 6900, 2, 1, 70.00, JSON_ARRAY('近地铁', '双阳台', '可做饭', '储物间'), JSON_ARRAY('https://images.unsplash.com/photo-1600047509807-ba8f99d2cdde?auto=format&fit=crop&w=1200&q=80'), 'active'),
    (24, 2, '唐镇安静一居', '社区绿化好，适合张江通勤，房东直租。', '上海', '浦东新区', '唐安路附近', 4500, 1, 1, 44.00, JSON_ARRAY('小区安静', '可做饭', '电梯', '绿化好'), JSON_ARRAY('https://images.unsplash.com/photo-1600121848594-d8644e57abab?auto=format&fit=crop&w=1200&q=80'), 'active'),
    (25, 2, '曹路实用两居', '户型方正，卧室均有采光，租金预算友好。', '上海', '浦东新区', '金海路附近', 5000, 2, 1, 65.00, JSON_ARRAY('可做饭', '电梯', '宠物友好', '采光好'), JSON_ARRAY('https://images.unsplash.com/photo-1600607687939-ce8a6c25118c?auto=format&fit=crop&w=1200&q=80'), 'active'),
    (26, 2, '惠南新城一居', '近地铁站与商业街，新房精装，支持长租。', '上海', '浦东新区', '拱极路附近', 3300, 1, 1, 40.00, JSON_ARRAY('近地铁', '精装修', '可做饭', '新小区'), JSON_ARRAY('https://images.unsplash.com/photo-1600607688960-e095ff83135c?auto=format&fit=crop&w=1200&q=80'), 'active'),
    (27, 2, '世博滨江一居', '靠近滨江绿地，客厅开阔，通勤陆家嘴便捷。', '上海', '浦东新区', '耀华路附近', 7300, 1, 1, 54.00, JSON_ARRAY('近地铁', '近公园', '可做饭', '独立阳台'), JSON_ARRAY('https://images.unsplash.com/photo-1600047509358-9dc75507daeb?auto=format&fit=crop&w=1200&q=80'), 'active'),
    (28, 2, '源深体育中心两居', '高楼层视野好，双卧带飘窗，附近运动设施丰富。', '上海', '浦东新区', '源深路附近', 8600, 2, 1, 74.00, JSON_ARRAY('近地铁', '飘窗', '电梯', '近运动场'), JSON_ARRAY('https://images.unsplash.com/photo-1600607687644-c7171b42498f?auto=format&fit=crop&w=1200&q=80'), 'active'),
    (29, 2, '高行新江湾一居', '小区安静，厨房可明火，适合金桥方向上班。', '上海', '浦东新区', '东靖路附近', 4100, 1, 1, 43.00, JSON_ARRAY('可做饭', '电梯', '小区安静', '门禁'), JSON_ARRAY('https://images.unsplash.com/photo-1615529162924-f8605388461d?auto=format&fit=crop&w=1200&q=80'), 'active'),
    (30, 2, '周浦品质三居', '三卧室均带收纳，两个卫生间，适合家庭长期住。', '上海', '浦东新区', '周康路附近', 8200, 3, 2, 98.00, JSON_ARRAY('三房', '两卫', '可做饭', '电梯'), JSON_ARRAY('https://images.unsplash.com/photo-1600607687920-4e2a09cf159d?auto=format&fit=crop&w=1200&q=80'), 'active'),
    (31, 2, '张江科创社区一居', '公区有共享会客区，步行至园区班车点。', '上海', '浦东新区', '盛夏路附近', 5900, 1, 1, 47.00, JSON_ARRAY('近班车', '共享公区', '可做饭', '精装修'), JSON_ARRAY('https://images.unsplash.com/photo-1600566753190-17f0baa2a6c3?auto=format&fit=crop&w=1200&q=80'), 'active'),
    (32, 2, '陆家嘴金融城两居', '高层电梯房，双卧私密性好，适合两人稳定合租。', '上海', '浦东新区', '商城路附近', 10800, 2, 1, 78.00, JSON_ARRAY('近地铁', '高层', '电梯', '可做饭'), JSON_ARRAY('https://images.unsplash.com/photo-1600585152915-d208bec867a1?auto=format&fit=crop&w=1200&q=80'), 'active'),
    (33, 2, '浦东机场通勤一居', '靠近机场工作区，带独立厨房和充足收纳。', '上海', '浦东新区', '启航路附近', 3800, 1, 1, 41.00, JSON_ARRAY('机场通勤', '可做饭', '独立卫浴', '停车方便'), JSON_ARRAY('https://images.unsplash.com/photo-1600566753086-00f18fb6b3ea?auto=format&fit=crop&w=1200&q=80'), 'active'),
    (34, 2, '花木世纪公园两居', '安静小区，卧室朝南，步行可达世纪公园地铁站。', '上海', '浦东新区', '梅花路附近', 9300, 2, 1, 75.00, JSON_ARRAY('近地铁', '近公园', '电梯', '朝南'), JSON_ARRAY('https://images.unsplash.com/photo-1600607687939-ce8a6c25118c?auto=format&fit=crop&w=1200&q=80'), 'active'),
    (35, 2, '洋泾轻奢一居', '干湿分离卫浴，客厅明亮，通勤八佰伴和陆家嘴方便。', '上海', '浦东新区', '羽山路附近', 6500, 1, 1, 49.00, JSON_ARRAY('近地铁', '干湿分离', '精装修', '可做饭'), JSON_ARRAY('https://images.unsplash.com/photo-1600585154526-990dced4db0d?auto=format&fit=crop&w=1200&q=80'), 'active')
ON DUPLICATE KEY UPDATE
    landlord_id = VALUES(landlord_id),
    title = VALUES(title);
