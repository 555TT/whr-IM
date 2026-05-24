CREATE DATABASE IF NOT EXISTS whr_im DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
USE whr_im;

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
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    PRIMARY KEY (id),
    UNIQUE KEY uk_users_username (username)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户表';

CREATE TABLE friend_requests (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键',
    from_user_id BIGINT UNSIGNED NOT NULL COMMENT '申请人 ID',
    to_user_id BIGINT UNSIGNED NOT NULL COMMENT '接收人 ID',
    message VARCHAR(255) NOT NULL DEFAULT '' COMMENT '申请附言',
    status ENUM('pending', 'accepted', 'rejected') NOT NULL DEFAULT 'pending' COMMENT '申请状态',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    handled_at DATETIME NULL DEFAULT NULL COMMENT '处理时间',
    PRIMARY KEY (id),
    KEY idx_friend_requests_from_user_id (from_user_id),
    KEY idx_friend_requests_to_user_id (to_user_id),
    KEY idx_friend_requests_status (status),
    UNIQUE KEY uk_friend_requests_pending_pair (from_user_id, to_user_id, status),
    CONSTRAINT fk_friend_requests_from_user FOREIGN KEY (from_user_id) REFERENCES users (id) ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT fk_friend_requests_to_user FOREIGN KEY (to_user_id) REFERENCES users (id) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='好友申请表';

CREATE TABLE friends (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键',
    user_id BIGINT UNSIGNED NOT NULL COMMENT '用户 ID',
    friend_id BIGINT UNSIGNED NOT NULL COMMENT '好友 ID',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    PRIMARY KEY (id),
    UNIQUE KEY uk_friends_user_friend (user_id, friend_id),
    KEY idx_friends_friend_id (friend_id),
    CONSTRAINT fk_friends_user FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT fk_friends_friend FOREIGN KEY (friend_id) REFERENCES users (id) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='好友关系表';

CREATE TABLE messages (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT COMMENT '主键',
    sender_id BIGINT UNSIGNED NOT NULL COMMENT '发送者 ID',
    receiver_id BIGINT UNSIGNED NOT NULL COMMENT '接收者 ID',
    ciphertext TEXT NOT NULL COMMENT '密文消息内容',
    algorithm VARCHAR(50) NOT NULL COMMENT '消息加密算法',
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '发送时间',
    PRIMARY KEY (id),
    KEY idx_messages_sender_receiver_created_at (sender_id, receiver_id, created_at),
    KEY idx_messages_receiver_sender_created_at (receiver_id, sender_id, created_at),
    CONSTRAINT fk_messages_sender FOREIGN KEY (sender_id) REFERENCES users (id) ON DELETE CASCADE ON UPDATE CASCADE,
    CONSTRAINT fk_messages_receiver FOREIGN KEY (receiver_id) REFERENCES users (id) ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='聊天消息表';
