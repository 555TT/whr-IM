CREATE DATABASE IF NOT EXISTS whr_im DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE whr_im;

DROP TABLE IF EXISTS moment_comments;
DROP TABLE IF EXISTS moment_likes;
DROP TABLE IF EXISTS moments;
DROP TABLE IF EXISTS group_message_keys;
DROP TABLE IF EXISTS group_messages;
DROP TABLE IF EXISTS group_members;
DROP TABLE IF EXISTS chat_groups;
DROP TABLE IF EXISTS messages;
DROP TABLE IF EXISTS friends;
DROP TABLE IF EXISTS friend_requests;
DROP TABLE IF EXISTS users;

CREATE TABLE users (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键',
    username VARCHAR(50) NOT NULL COMMENT '用户名，唯一',
    password_hash VARCHAR(255) NOT NULL COMMENT '加密后的密码',
    nickname VARCHAR(50) NOT NULL COMMENT '昵称',
    avatar VARCHAR(255) NOT NULL DEFAULT 'https://api.dicebear.com/7.x/initials/svg?seed=default-user' COMMENT '系统默认头像地址，不允许用户修改',
    gender TINYINT NOT NULL DEFAULT 0 COMMENT '性别：0-未知，1-男，2-女',
    signature VARCHAR(255) NOT NULL DEFAULT '' COMMENT '个性签名',
    public_key TEXT NOT NULL COMMENT '用户公钥',
    public_key_algorithm VARCHAR(50) NOT NULL DEFAULT '' COMMENT '公钥算法',
    PRIMARY KEY (id),
    UNIQUE KEY uk_users_username (username)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户表';

CREATE TABLE friend_requests (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键',
    from_user_id BIGINT UNSIGNED NOT NULL COMMENT '申请人 ID',
    to_user_id BIGINT UNSIGNED NOT NULL COMMENT '接收人 ID',
    message VARCHAR(255) NOT NULL DEFAULT '' COMMENT '申请附言',
    status VARCHAR(20) NOT NULL DEFAULT 'pending' COMMENT '申请状态',
    PRIMARY KEY (id),
    KEY idx_friend_requests_from_user_id (from_user_id),
    KEY idx_friend_requests_to_user_id (to_user_id),
    KEY idx_friend_requests_status (status),
    CONSTRAINT fk_friend_requests_from_user FOREIGN KEY (from_user_id) REFERENCES users (id) ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT fk_friend_requests_to_user FOREIGN KEY (to_user_id) REFERENCES users (id) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='好友申请表';

CREATE TABLE friends (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键',
    user_id BIGINT UNSIGNED NOT NULL COMMENT '用户 ID',
    friend_id BIGINT UNSIGNED NOT NULL COMMENT '好友 ID',
    PRIMARY KEY (id),
    UNIQUE KEY uk_friends_user_friend (user_id, friend_id),
    KEY idx_friends_friend_id (friend_id),
    CONSTRAINT fk_friends_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT fk_friends_friend FOREIGN KEY (friend_id) REFERENCES users (id) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='好友关系表';

CREATE TABLE chat_groups (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键',
    name VARCHAR(50) NOT NULL COMMENT '群名',
    owner_id BIGINT UNSIGNED NOT NULL COMMENT '创建者 ID',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    PRIMARY KEY (id),
    KEY idx_chat_groups_owner (owner_id),
    CONSTRAINT fk_chat_groups_owner FOREIGN KEY (owner_id) REFERENCES users (id) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='聊天群组表';

CREATE TABLE group_members (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键',
    group_id BIGINT UNSIGNED NOT NULL COMMENT '群 ID',
    user_id BIGINT UNSIGNED NOT NULL COMMENT '成员 ID',
    joined_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '入群时间',
    PRIMARY KEY (id),
    UNIQUE KEY uk_group_members_group_user (group_id, user_id),
    KEY idx_group_members_user (user_id),
    CONSTRAINT fk_group_members_group FOREIGN KEY (group_id) REFERENCES chat_groups (id) ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT fk_group_members_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='群成员关系表';

CREATE TABLE group_messages (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键',
    group_id BIGINT UNSIGNED NOT NULL COMMENT '群 ID',
    sender_id BIGINT UNSIGNED NOT NULL COMMENT '发送者 ID',
    content_ciphertext TEXT NOT NULL COMMENT 'AES-GCM 加密后的消息密文 (base64)',
    content_iv VARCHAR(64) NOT NULL COMMENT 'AES-GCM IV (base64)',
    content_algorithm VARCHAR(50) NOT NULL COMMENT '内容加密算法,如 aes-gcm-256',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '发送时间',
    PRIMARY KEY (id),
    KEY idx_group_messages_group_created_at (group_id, created_at),
    KEY idx_group_messages_sender (sender_id),
    CONSTRAINT fk_group_messages_group FOREIGN KEY (group_id) REFERENCES chat_groups (id) ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT fk_group_messages_sender FOREIGN KEY (sender_id) REFERENCES users (id) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='群聊消息表';

CREATE TABLE group_message_keys (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键',
    message_id BIGINT UNSIGNED NOT NULL COMMENT '群消息 ID',
    user_id BIGINT UNSIGNED NOT NULL COMMENT '该 AES 密钥密文的接收者',
    key_ciphertext TEXT NOT NULL COMMENT 'RSA-OAEP 加密的 AES 密钥 (base64)',
    key_algorithm VARCHAR(50) NOT NULL COMMENT 'AES 密钥包裹算法,如 rsa-oaep-sha256',
    PRIMARY KEY (id),
    UNIQUE KEY uk_group_message_keys_message_user (message_id, user_id),
    KEY idx_group_message_keys_user (user_id),
    CONSTRAINT fk_group_message_keys_message FOREIGN KEY (message_id) REFERENCES group_messages (id) ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT fk_group_message_keys_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='群消息会话密钥包裹表';

CREATE TABLE moments (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键',
    user_id BIGINT UNSIGNED NOT NULL COMMENT '发布用户 ID',
    content TEXT NOT NULL COMMENT '朋友圈正文',
    images_json TEXT NOT NULL COMMENT '图片 objectKey 列表 JSON',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    PRIMARY KEY (id),
    KEY idx_moments_user (user_id),
    CONSTRAINT fk_moments_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='朋友圈动态表';

CREATE TABLE moment_likes (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键',
    moment_id BIGINT UNSIGNED NOT NULL COMMENT '朋友圈动态 ID',
    user_id BIGINT UNSIGNED NOT NULL COMMENT '点赞用户 ID',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '点赞时间',
    PRIMARY KEY (id),
    UNIQUE KEY uk_moment_likes_moment_user (moment_id, user_id),
    KEY idx_moment_likes_moment_id (moment_id),
    KEY idx_moment_likes_user_id (user_id),
    CONSTRAINT fk_moment_likes_moment FOREIGN KEY (moment_id) REFERENCES moments (id) ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT fk_moment_likes_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='朋友圈点赞表';

CREATE TABLE moment_comments (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键',
    moment_id BIGINT UNSIGNED NOT NULL COMMENT '朋友圈动态 ID',
    user_id BIGINT UNSIGNED NOT NULL COMMENT '评论用户 ID',
    content TEXT NOT NULL COMMENT '评论内容',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '评论时间',
    PRIMARY KEY (id),
    KEY idx_moment_comments_moment_id (moment_id),
    KEY idx_moment_comments_user_id (user_id),
    CONSTRAINT fk_moment_comments_moment FOREIGN KEY (moment_id) REFERENCES moments (id) ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT fk_moment_comments_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='朋友圈评论表';

CREATE TABLE messages (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键',
    sender_id BIGINT UNSIGNED NOT NULL COMMENT '发送者 ID',
    receiver_id BIGINT UNSIGNED NOT NULL COMMENT '接收者 ID',
    sender_ciphertext TEXT NOT NULL COMMENT '发送者可解密密文',
    sender_algorithm VARCHAR(50) NOT NULL COMMENT '发送者密文算法',
    receiver_ciphertext TEXT NOT NULL COMMENT '接收者可解密密文',
    receiver_algorithm VARCHAR(50) NOT NULL COMMENT '接收者密文算法',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '发送时间',
    PRIMARY KEY (id),
    KEY idx_messages_sender_receiver_created_at (sender_id, receiver_id, created_at),
    KEY idx_messages_receiver_sender_created_at (receiver_id, sender_id, created_at),
    CONSTRAINT fk_messages_sender FOREIGN KEY (sender_id) REFERENCES users (id) ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT fk_messages_receiver FOREIGN KEY (receiver_id) REFERENCES users (id) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='聊天消息表';
