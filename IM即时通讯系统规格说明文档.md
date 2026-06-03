# IM 即时通讯系统规格说明文档

## 1. 项目概述

### 1.1 项目名称
基于 Golang + Vue + WebSocket 的 IM 即时通讯系统

### 1.2 项目背景
随着互联网应用的发展，即时通讯系统已经成为常见的基础功能。为了学习和实践前后端分离开发、实时通信、数据库设计与用户关系管理，本项目拟开发一个面向 Web 端的轻量级 IM 系统。

### 1.3 项目目标
实现一个具备社交属性的即时通讯系统，支持以下核心能力：

- 用户注册与登录
- 个人信息查看与修改（头像除外）
- 个人主页风格皮肤选择（个性装扮）
- 查看好友个人主页
- 添加好友申请
- 同意/拒绝好友申请
- 好友列表管理
- 与好友进行实时单聊
- 创建群聊、邀请好友入群与退出群聊
- 群聊历史消息查询与实时消息推送
- 基于 RSA-OAEP 的单聊端到端加密文本消息
- 基于 AES-GCM + RSA-OAEP 混合加密的群聊端到端加密文本消息
- 朋友圈动态发布（含图片上传)、查看、点赞、评论、删除
- 基于 DeepSeek 大模型的朋友圈 AI 文案生成与润色助手
- 基于 MinIO 对象存储的图片上传

### 1.4 项目范围
本项目当前范围覆盖 **IM 聊天 + 轻量社交动态**，支持：

- Web 端使用
- 单人对单人聊天
- 基础群聊能力
- 文本聊天消息（端到端加密）
- 基础好友关系管理
- 单聊与群聊的端到端加密消息
- 朋友圈图文动态（明文存储,仅好友可见)
- AI 文案辅助
- 个性化主页皮肤装扮

本阶段**不包含**：

- 文件消息（图片仅支持朋友圈,不支持聊天图片消息）
- 聊天消息已读回执
- 聊天消息撤回
- 多端同步
- 离线推送
- 群主转让、禁言、解散群、邀请审批等复杂群管理功能
- 朋友圈消息加密（朋友圈正文与图片为明文）

---

## 2. 技术栈

### 2.1 后端
- Golang
- Gin
- GORM
- WebSocket
- MySQL
- MinIO（对象存储,用于朋友圈图片）
- DeepSeek（大模型 Chat Completions API,用于朋友圈 AI 文案助手）

### 2.2 前端
- Vue 3
- Axios
- Vue Router
- Pinia
- WebSocket API
- Web Crypto API（端到端加密）

### 2.3 开发模式
采用 **单仓库 Monorepo** 结构：

```text
whr-im/
├── server/   # Go 后端
├── web/      # Vue 前端
└── docs/     # 文档
```

---

## 3. 系统总体架构

系统采用 **REST API + WebSocket** 的架构模式。

### 3.1 架构说明
- **REST API**
  - 负责注册、登录、资料修改、好友申请、好友列表、历史消息查询
- **WebSocket**
  - 负责聊天消息的实时收发

### 3.2 模块划分

#### 后端模块
1. 用户认证模块
2. 用户资料模块（含主页皮肤、好友可见公开主页）
3. 好友申请模块
4. 好友关系模块
5. 单聊消息模块
6. 群聊模块
7. WebSocket 连接管理模块
8. 朋友圈模块（动态、点赞、评论）
9. 朋友圈 AI 文案助手模块（DeepSeek 调用）
10. 图片上传模块（MinIO 对象存储）

#### 前端模块
1. 登录/注册页
2. 主聊天页
3. 好友申请页
4. 个人资料页（含主页皮肤选择）
5. 好友个人主页页（好友主页装扮 + 该好友朋友圈）
6. 好友列表与聊天窗口组件
7. 群聊创建与群信息组件
8. 朋友圈动态页（发布、AI 文案、点赞、评论、删除）
9. 全局导航栏组件

---

## 4. 功能需求

## 4.1 用户注册
### 功能描述
用户通过用户名和密码完成注册。

### 输入
- 用户名
- 密码
- 确认密码

### 处理逻辑
- 校验用户名长度需在 4 到 20 位之间
- 校验密码长度需在 6 到 20 位之间
- 校验用户名是否已存在
- 校验密码是否符合基本要求
- 对密码进行加密存储
- 创建用户记录

### 输出
- 注册成功提示
- 或失败原因提示

---

## 4.2 用户登录
### 功能描述
用户通过用户名和密码登录系统。

### 输入
- 用户名
- 密码

### 处理逻辑
- 校验用户名长度需在 4 到 20 位之间
- 校验密码长度需在 6 到 20 位之间
- 校验用户名是否存在
- 校验密码是否正确
- 登录成功后返回 token

### 输出
- token
- 用户基础信息

---

## 4.3 个人信息查看与修改
### 功能描述
用户可查看个人资料，并修改除头像外的资料信息，可在多套预设主页皮肤中选择一种作为个性装扮。

### 可修改字段
- 昵称
- 性别
- 个性签名
- 主页皮肤标识 `homepageSkin`

### 头像规则
- 头像不允许用户自行上传或修改
- 系统为新注册用户分配默认头像 URL
- 默认头像可直接使用公开图片地址，例如：`https://api.dicebear.com/7.x/initials/svg?seed=default-user`

### 主页皮肤规则
- 注册时默认皮肤为 `aurora`
- 允许的皮肤标识固定为：`aurora`、`sunset`、`galaxy`、`mint`、`peach`
- 提交其他皮肤标识时,服务端返回 400 错误
- 皮肤标识仅作为前端渲染主页背景与装饰风格的依据,不携带其他业务含义

### 处理逻辑
- 登录后通过 token 获取当前用户信息
- 注册时为用户写入默认头像 URL 与默认皮肤标识
- 保存修改后的资料到数据库，但不允许更新头像字段
- 修改主页皮肤时校验值必须在允许集合内

---

## 4.3.1 好友主页查看
### 功能描述
用户可查看自己或好友的公开主页信息,用于浏览对方装扮风格、签名以及朋友圈。

### 输入
- 目标用户 ID

### 处理逻辑
- 校验当前登录用户是目标用户本人,或与目标用户为好友
- 非好友访问时返回错误,不暴露任何用户资料
- 返回内容仅包含公开字段,不返回密码、公钥、性别等敏感字段

### 输出（公开主页字段）
- `id`
- `nickname`
- `avatar`
- `signature`
- `homepageSkin`

---

## 4.4 添加好友申请
### 功能描述
用户可向其他用户发起好友申请。

### 输入
- 目标用户名或目标用户 ID
- 申请附言

### 处理逻辑
- 校验目标用户是否存在
- 校验不能添加自己
- 校验双方是否已经是好友
- 校验是否已有未处理申请
- 创建好友申请记录

### 输出
- 申请发送成功
- 或失败提示

---

## 4.5 查看好友申请
### 功能描述
用户可查看收到的好友申请。

### 展示内容
- 申请人信息
- 申请附言
- 申请状态
- 申请时间

---

## 4.6 同意/拒绝好友申请
### 功能描述
用户对收到的申请进行处理。

### 处理逻辑
- 同意时：
  - 更新申请状态为 accepted
  - 建立双方好友关系
- 拒绝时：
  - 更新申请状态为 rejected

### 输出
- 处理成功提示

---

## 4.7 好友列表查询
### 功能描述
用户可查看自己的好友列表。

### 展示内容
- 好友昵称
- 头像
- 个性签名
- 在线状态（MVP 可选做简单在线/离线）

---

## 4.8 单聊功能
### 功能描述
用户可与好友进行实时文本聊天。

### 输入
- 接收方用户 ID
- 消息内容

### 处理逻辑
- 校验发送方已登录
- 校验接收方是好友
- 发送方客户端在本地生成或加载自己的 RSA-OAEP 密钥对，并确保公钥已上传到服务端
- 发送方客户端从好友列表中获取接收方公钥
- 对同一条明文分别使用发送方公钥与接收方公钥加密，生成两份密文：
  - `senderCiphertext`：仅发送方自己的私钥可解密
  - `receiverCiphertext`：仅接收方自己的私钥可解密
- 客户端向后端提交 `receiverId`、两份密文及对应算法标识 `rsa-oaep-sha256`
- 后端校验密文与算法字段完整性后保存消息到数据库
- 若接收方在线，则通过 WebSocket 实时推送该条消息
- 发送方通过创建消息接口返回体拿到已保存消息，用本地私钥解密 `senderCiphertext` 后展示明文

### 输出
- 发送成功回执（响应中仅包含双份密文，不返回明文）
- 接收方收到实时消息（消息体仅包含双份密文，不返回明文）

---

## 4.9 历史消息查询
### 功能描述
用户打开聊天窗口时可查看与某好友的历史聊天记录。

### 展示内容
- 发送方
- 接收方
- 消息内容（由当前登录用户使用本地私钥解密自己可读的那份密文后展示）
- 发送时间

### 处理逻辑
- 前端请求 `/api/messages?friendId=...` 获取历史消息
- 若当前登录用户是该条消息发送方，则优先取 `senderCiphertext`
- 若当前登录用户是该条消息接收方，则优先取 `receiverCiphertext`
- 前端使用本地保存的私钥解密对应密文并展示明文
- 若本地私钥不存在或解密失败，则以前端占位文案 `***（已加密）` 展示，不向服务端请求明文兜底

---

## 4.10 群聊功能
### 功能描述
用户可创建群聊、邀请自己的好友加入群聊、退出群聊，并在群内进行实时文本聊天。

### 输入
- 群名称
- 被邀请成员 ID 列表
- 群聊消息内容

### 处理逻辑
- 创建群时，创建者必须已登录
- 创建群时，仅允许从自己的好友列表中选择群成员
- 创建成功后，系统自动将创建者本人加入群成员列表
- 群成员可邀请自己的好友加入当前群聊
- 用户可主动退出群聊
- 群聊消息采用“消息内容共享密文 + 每个成员一份会话密钥密文”的混合加密模型
- 服务端只保存群聊密文、会话密钥密文与算法标识，不保存消息明文
- 若群成员在线，则通过 WebSocket 实时推送群消息

### 输出
- 创建群成功回执
- 邀请成员成功回执
- 退出群聊成功回执
- 群聊消息发送成功回执

---

## 4.11 群聊历史消息查询
### 功能描述
用户打开群聊窗口时可查看自己加入该群之后可访问的历史消息。

### 展示内容
- 群名称
- 发送方
- 消息内容（由当前登录用户使用本地私钥解密自己那一份群消息会话密钥后再解密正文）
- 发送时间

### 处理逻辑
- 前端请求 `/api/groups/{id}/messages` 获取当前用户可访问的群历史消息
- 每条群消息仅返回一份共享正文密文 `contentCiphertext` 和当前登录用户对应的 `keyCiphertext`
- 前端先用本地 RSA 私钥解密 `keyCiphertext`，得到该条消息的 AES 会话密钥
- 再使用该 AES 会话密钥解密 `contentCiphertext`，得到消息明文
- 新成员默认看不到入群前的历史消息，因为服务端不会为其补发旧消息对应的会话密钥密文
- 若本地私钥不存在或解密失败，则以前端占位文案 `***（已加密）` 展示，不向服务端请求明文兜底

---

## 4.12 图片上传
### 功能描述
用户在发布朋友圈动态等场景下可将本地图片上传到服务端的对象存储,服务端返回可访问的 URL 与对象存储键。

### 输入
- 表单字段 `file`(图片二进制文件)

### 处理逻辑
- 校验当前用户已登录(JWT 鉴权)
- 校验文件存在且 `Content-Type` 以 `image/` 开头
- 按 `uploads/users/{userId}/{nanoTimestamp}{ext}` 规则生成对象存储 Key
- 调用对象存储(MinIO 或静态存储)上传文件
- 返回 `objectKey` 与可访问的公开 URL

### 输出
- `objectKey`:对象存储中该文件的键,用于后续发布朋友圈时引用
- `url`:可直接在前端 `<img>` 标签使用的访问地址

### 限制
- 当前仅支持图片类型
- 不进行内容审核
- 单文件大小受 Gin 默认 multipart 解析上限限制

---

## 4.13 朋友圈动态发布
### 功能描述
用户可以发布带正文与若干张图片的朋友圈动态,动态默认对自己与好友可见。

### 输入
- 正文 `content`
- 图片 `imageKeys`(数组,元素为上传接口返回的 `objectKey`)

### 处理逻辑
- 校验正文经 trim 之后不为空
- 将 `imageKeys` 序列化为 JSON 存入 `moments.images_json`
- 关联当前用户 ID
- 创建成功后返回该动态的视图(含已解析为可访问 URL 的图片列表)

### 输出
- 包含 `id`、`userId`、`nickname`、`avatar`、`content`、`images`、`likeCount`、`likedByMe`、`comments`、`createdAt` 的动态视图

### 安全/可见性
- 朋友圈正文与图片为明文存储,不进行端到端加密
- 仅本人与好友可见

---

## 4.14 朋友圈动态列表查询
### 功能描述
用户可拉取自己可见的朋友圈时间线,或拉取某位好友(或自己)所有的朋友圈动态。

### 处理逻辑
- 时间线列表:返回当前用户自己 + 所有好友的朋友圈动态,按时间倒序
- 指定用户列表:仅当目标为本人或当前用户的好友时返回,否则返回错误
- 每条动态附带:点赞数、当前用户是否已点赞、评论列表(含评论人昵称与时间)

### 输出
- 朋友圈动态视图数组,每条结构同 4.13

---

## 4.15 朋友圈点赞与取消点赞
### 功能描述
用户可对自己可见的朋友圈动态进行点赞或取消点赞。

### 处理逻辑
- 校验动态存在且当前用户可见(自己或好友)
- 同一用户对同一条动态只能保留一条点赞记录(数据库唯一索引保证)
- 取消点赞会删除对应的点赞记录

### 输出
- 点赞:201 Created
- 取消点赞:204 No Content

---

## 4.16 朋友圈评论
### 功能描述
用户可对自己可见的朋友圈动态发表评论,评论按时间正序展示。

### 输入
- 动态 ID(URL 参数)
- 评论正文 `content`

### 处理逻辑
- 校验动态存在且当前用户可见
- 评论内容 trim 之后不能为空
- 评论与发布人、动态、创建时间一并落库

### 输出
- 201 Created(评论内容通过下次拉取动态列表/详情返回)

---

## 4.17 朋友圈删除
### 功能描述
朋友圈作者可删除自己发布的动态;删除后该动态对应的点赞与评论由数据库级联删除。

### 处理逻辑
- 校验动态存在
- 校验当前用户是动态作者,否则禁止删除

### 输出
- 204 No Content

---

## 4.18 朋友圈 AI 文案助手
### 功能描述
朋友圈发布页提供 AI 文案助手,可根据用户的想法直接生成一段朋友圈文案,或对用户已写的草稿进行润色。底层通过调用 DeepSeek Chat Completions API 完成。

### 输入
- `mode`:`generate` 或 `polish`
- `prompt`:生成模式下用户输入的想法
- `content`:润色模式下的原始文案
- `tone`:可选,期望的语气(如 治愈、文艺、搞笑等);未指定时按"自然"语气处理
- `hasImage`:布尔,告诉模型当前是否带图,以便文案与配图风格一致

### 处理逻辑
- 校验 `mode` 必须为 `generate` 或 `polish`
- `generate` 模式下 `prompt` 必填
- `polish` 模式下 `content` 必填
- 当后端未配置 `DEEPSEEK_API_KEY` 时,接口返回 503,提示"AI 助手暂未配置"
- 调用 DeepSeek 模型时使用系统提示:中文社交平台朋友圈文案助手,要求输出 1~3 句口语化文案,不带 markdown
- DeepSeek 返回非 2xx 或空内容时,返回 502,提示"AI 生成暂时失败"
- 该接口仅辅助生成文案,不持久化任何 AI 内容,最终是否使用、是否修改由用户决定

### 输出
- `{ "text": "..." }`:AI 给出的最终文案文本

---

## 5. 非功能需求

### 5.1 性能要求
- 支持基础并发访问
- 支持多个用户同时在线聊天
- 页面响应时间正常情况下不超过 3 秒
- 朋友圈列表、单聊/群聊历史消息建议按时间倒序索引以保证查询性能

### 5.2 安全要求
- 用户密码加密存储
- 登录鉴权使用 JWT
- 接口需校验身份
- WebSocket 建连时需校验 token
- 聊天明文不落库，服务端只保存密文与算法标识
- 每个用户需维护自己的公私钥对：公钥上传服务端用于被他人加密，私钥仅保存在本地浏览器 `localStorage`
- 单聊消息采用 `RSA-OAEP + SHA-256`（标识为 `rsa-oaep-sha256`）进行端到端加密
- 群聊消息采用 `AES-GCM-256 + RSA-OAEP` 混合加密：消息正文由 AES-GCM 加密，每个成员各自持有一份用 RSA 公钥包裹的 AES 会话密钥
- 新成员默认不具备解密入群前历史群消息的能力，服务端不为其补发旧消息对应的会话密钥密文
- 好友主页与朋友圈仅对本人或好友可见,服务端在所有相关接口都需做可见性校验
- 朋友圈正文与图片为明文存储,不属于端到端加密范围
- 图片上传接口仅允许 `image/*` 类型,对象 Key 按用户分目录隔离
- AI 文案助手所用 DeepSeek API Key 仅保存在服务端配置/环境变量,不下发到前端

### 5.3 可维护性要求
- 后端分层清晰
- 前端组件拆分合理
- 接口命名规范
- 代码结构清晰，便于维护和答辩展示

### 5.4 可用性要求
- 页面操作简单直观
- 功能流程完整
- 错误提示明确

---

## 6. 数据库设计

## 6.1 users 用户表

| 字段名 | 类型 | 说明 |
|---|---|---|
| id | bigint | 主键 |
| username | varchar | 用户名，唯一 |
| password_hash | varchar | 加密后的密码 |
| nickname | varchar | 昵称 |
| avatar | varchar | 系统默认头像地址，不允许用户修改 |
| gender | tinyint | 性别:0-未知,1-男,2-女 |
| signature | varchar | 个性签名 |
| homepage_skin | varchar | 主页预设皮肤标识,默认 `aurora`,合法值 aurora/sunset/galaxy/mint/peach |
| public_key | text | 用户公钥 |
| public_key_algorithm | varchar | 公钥算法标识 |

---

## 6.2 friend_requests 好友申请表

| 字段名 | 类型 | 说明 |
|---|---|---|
| id | bigint | 主键 |
| from_user_id | bigint | 申请人 ID |
| to_user_id | bigint | 接收人 ID |
| message | varchar | 申请附言 |
| status | varchar | pending/accepted/rejected |

---

## 6.3 friends 好友关系表

| 字段名 | 类型 | 说明 |
|---|---|---|
| id | bigint | 主键 |
| user_id | bigint | 用户 ID |
| friend_id | bigint | 好友 ID |

说明：
- 采用"双向两条记录"的方式保存好友关系,(user_id, friend_id) 上有唯一索引,查询更直接。

---

## 6.4 messages 消息表

| 字段名 | 类型 | 说明 |
|---|---|---|
| id | bigint | 主键 |
| sender_id | bigint | 发送者 ID |
| receiver_id | bigint | 接收者 ID |
| sender_ciphertext | text | 发送方可解密密文 |
| sender_algorithm | varchar | 发送方密文算法标识 |
| receiver_ciphertext | text | 接收方可解密密文 |
| receiver_algorithm | varchar | 接收方密文算法标识 |
| created_at | datetime | 发送时间 |

---

## 6.5 chat_groups 群组表

| 字段名 | 类型 | 说明 |
|---|---|---|
| id | bigint | 主键 |
| name | varchar | 群名称 |
| owner_id | bigint | 创建者 ID |
| created_at | datetime | 创建时间 |

---

## 6.6 group_members 群成员表

| 字段名 | 类型 | 说明 |
|---|---|---|
| id | bigint | 主键 |
| group_id | bigint | 群 ID |
| user_id | bigint | 成员 ID |
| joined_at | datetime | 入群时间 |

说明：
- 同一用户在同一群内只能存在一条成员记录
- 创建群时会自动写入创建者本人记录

---

## 6.7 group_messages 群消息表

| 字段名 | 类型 | 说明 |
|---|---|---|
| id | bigint | 主键 |
| group_id | bigint | 群 ID |
| sender_id | bigint | 发送者 ID |
| content_ciphertext | text | AES-GCM 加密后的群消息正文密文 |
| content_iv | varchar | AES-GCM 使用的随机 IV |
| content_algorithm | varchar | 正文加密算法标识，当前为 `aes-gcm-256` |
| created_at | datetime | 发送时间 |

---

## 6.8 group_message_keys 群消息密钥表

| 字段名 | 类型 | 说明 |
|---|---|---|
| id | bigint | 主键 |
| message_id | bigint | 对应的群消息 ID |
| user_id | bigint | 该条会话密钥密文对应的成员 ID |
| key_ciphertext | text | 使用该成员 RSA 公钥加密后的 AES 会话密钥 |
| key_algorithm | varchar | 密钥包裹算法标识，当前为 `rsa-oaep-sha256` |

说明：
- 一条群消息在 `group_messages` 中只保存一份正文密文
- 同一条群消息会在 `group_message_keys` 中按群成员数量保存多份会话密钥密文
- 每个成员只能解密属于自己的那一份 `key_ciphertext`

---

## 6.9 moments 朋友圈动态表

| 字段名 | 类型 | 说明 |
|---|---|---|
| id | bigint | 主键 |
| user_id | bigint | 发布用户 ID |
| content | text | 朋友圈正文 |
| images_json | text | 图片 objectKey 列表 JSON |
| created_at | datetime | 创建时间 |

说明:
- `images_json` 中保存上传接口返回的 `objectKey` 数组,展示时由后端根据对象存储配置解析为完整 URL
- 朋友圈正文不加密,以明文存储

---

## 6.10 moment_likes 朋友圈点赞表

| 字段名 | 类型 | 说明 |
|---|---|---|
| id | bigint | 主键 |
| moment_id | bigint | 朋友圈动态 ID |
| user_id | bigint | 点赞用户 ID |
| created_at | datetime | 点赞时间 |

说明:
- (moment_id, user_id) 上有唯一索引,保证同一用户对同一动态最多一条点赞记录

---

## 6.11 moment_comments 朋友圈评论表

| 字段名 | 类型 | 说明 |
|---|---|---|
| id | bigint | 主键 |
| moment_id | bigint | 朋友圈动态 ID |
| user_id | bigint | 评论用户 ID |
| content | text | 评论内容 |
| created_at | datetime | 评论时间 |

---

## 7. 接口设计

## 7.1 认证接口

### 注册
**POST** `/api/auth/register`

请求参数：
```json
{
  "username": "zhangsan",
  "password": "123456",
  "confirmPassword": "123456"
}
```

### 登录
**POST** `/api/auth/login`

请求参数：
```json
{
  "username": "zhangsan",
  "password": "123456"
}
```

响应示例：
```json
{
  "token": "jwt-token",
  "user": {
    "id": 1,
    "username": "zhangsan",
    "nickname": "张三"
  }
}
```

---

## 7.2 用户接口

### 获取当前用户信息
**GET** `/api/users/me`

### 修改当前用户信息
**PUT** `/api/users/me`

请求参数：
```json
{
  "nickname": "张三",
  "gender": 1,
  "signature": "hello",
  "homepageSkin": "aurora"
}
```

说明:
- `homepageSkin` 必须为 `aurora`/`sunset`/`galaxy`/`mint`/`peach` 之一
- 不允许通过该接口修改头像

### 获取指定用户的公开主页
**GET** `/api/users/{id}/profile`

说明:
- 仅当目标为本人或当前用户的好友时返回
- 响应仅包含 `id`、`nickname`、`avatar`、`signature`、`homepageSkin`

响应示例:
```json
{
  "id": 2,
  "nickname": "李四",
  "avatar": "https://api.dicebear.com/7.x/initials/svg?seed=default-user",
  "signature": "保持热爱",
  "homepageSkin": "sunset"
}
```

### 获取指定用户的朋友圈
**GET** `/api/users/{id}/moments`

说明:
- 仅当目标为本人或当前用户的好友时返回
- 响应同朋友圈动态列表(见 7.7)

---

## 7.2.1 图片上传接口

### 上传图片
**POST** `/api/uploads/images`

请求方式: `multipart/form-data`,表单字段:
- `file`: 图片文件

响应示例:
```json
{
  "objectKey": "uploads/users/1/1717410000000000000.png",
  "url": "https://minio.example.com/whr-im/uploads/users/1/1717410000000000000.png"
}
```

说明:
- 仅支持 `image/*` 类型
- 返回的 `objectKey` 可用于发布朋友圈时的 `imageKeys` 字段
- 返回的 `url` 可直接用于前端 `<img>` 展示

---

## 7.3 好友申请接口

### 发起好友申请
**POST** `/api/friend-requests`

```json
{
  "toUserId": 2,
  "message": "你好，想加你为好友"
}
```

### 查看收到的申请
**GET** `/api/friend-requests/incoming`

### 同意申请
**PUT** `/api/friend-requests/{id}/accept`

### 拒绝申请
**PUT** `/api/friend-requests/{id}/reject`

---

## 7.4 好友接口

### 获取好友列表
**GET** `/api/friends`

---

## 7.5 消息接口

### 上传当前用户公钥
**PUT** `/api/users/me/public-key`

请求参数：
```json
{
  "publicKey": "base64-spki-public-key",
  "algorithm": "rsa-oaep-sha256"
}
```

说明：
- 用户登录后，客户端需确保当前账号公钥已上传
- 服务端保存公钥，供好友发送加密消息时使用

### 发送单聊消息
**POST** `/api/messages`

请求参数：
```json
{
  "receiverId": 2,
  "senderCiphertext": "base64-ciphertext-for-sender",
  "senderAlgorithm": "rsa-oaep-sha256",
  "receiverCiphertext": "base64-ciphertext-for-receiver",
  "receiverAlgorithm": "rsa-oaep-sha256"
}
```

响应示例：
```json
{
  "id": 1001,
  "senderId": 1,
  "receiverId": 2,
  "senderCiphertext": "base64-ciphertext-for-sender",
  "senderAlgorithm": "rsa-oaep-sha256",
  "receiverCiphertext": "base64-ciphertext-for-receiver",
  "receiverAlgorithm": "rsa-oaep-sha256",
  "createdAt": "2026-05-25T10:00:00Z"
}
```

### 获取单聊历史消息
**GET** `/api/messages?friendId=2`

响应说明：
- 返回聊天记录数组
- 每条记录仅返回双份密文与算法标识，不返回明文内容

---

## 7.6 群聊接口

### 创建群聊
**POST** `/api/groups`

请求参数：
```json
{
  "name": "项目讨论群",
  "memberIds": [2, 3]
}
```

### 获取我的群聊列表
**GET** `/api/groups`

### 获取群详情
**GET** `/api/groups/{id}`

响应说明：
- 返回群基础信息
- 返回群成员列表
- 返回每个成员的公钥与公钥算法，供前端发送群聊加密消息时使用

### 邀请好友入群
**POST** `/api/groups/{id}/members`

请求参数：
```json
{
  "memberIds": [4, 5]
}
```

### 退出群聊
**DELETE** `/api/groups/{id}/members/me`

### 发送群聊消息
**POST** `/api/groups/{id}/messages`

请求参数：
```json
{
  "contentCiphertext": "base64-aes-ciphertext",
  "contentIv": "base64-iv",
  "contentAlgorithm": "aes-gcm-256",
  "memberKeys": [
    {
      "userId": 1,
      "keyCiphertext": "base64-wrapped-aes-key-for-user-1",
      "keyAlgorithm": "rsa-oaep-sha256"
    },
    {
      "userId": 2,
      "keyCiphertext": "base64-wrapped-aes-key-for-user-2",
      "keyAlgorithm": "rsa-oaep-sha256"
    }
  ]
}
```

响应说明：
- 响应中仅返回共享正文密文、当前发送者自己的 `keyCiphertext` 以及算法标识
- 不返回任何明文消息内容

### 获取群聊历史消息
**GET** `/api/groups/{id}/messages`

响应说明：
- 返回当前登录用户可访问的群消息数组
- 每条记录包含：共享正文密文 `contentCiphertext`、`contentIv`、当前用户自己的 `keyCiphertext` 与算法标识
- 不返回其他成员的会话密钥密文，也不返回明文内容

---

## 7.7 朋友圈接口

### 发布朋友圈
**POST** `/api/moments`

请求参数:
```json
{
  "content": "今天的咖啡很好喝",
  "imageKeys": [
    "uploads/users/1/1717410000000000000.png"
  ]
}
```

响应示例:
```json
{
  "id": 101,
  "userId": 1,
  "nickname": "张三",
  "avatar": "https://api.dicebear.com/7.x/initials/svg?seed=default-user",
  "content": "今天的咖啡很好喝",
  "images": ["https://minio.example.com/whr-im/uploads/users/1/1717410000000000000.png"],
  "likeCount": 0,
  "likedByMe": false,
  "comments": [],
  "createdAt": "2026-06-03T12:00:00Z"
}
```

### 获取朋友圈时间线
**GET** `/api/moments`

说明:
- 返回当前用户自己 + 所有好友的动态,按时间倒序
- 每条动态结构同上,含 `likeCount`、`likedByMe`、`comments`

### 删除朋友圈
**DELETE** `/api/moments/{id}`

说明:
- 仅作者可删除,删除后相关点赞、评论由数据库级联清理
- 成功返回 204

### 点赞朋友圈
**POST** `/api/moments/{id}/likes`

说明:
- 仅本人或好友可见的动态可点赞
- 同一用户对同一动态重复点赞返回失败
- 成功返回 201

### 取消点赞
**DELETE** `/api/moments/{id}/likes/me`

说明:
- 取消当前用户对该动态的点赞
- 成功返回 204

### 发表朋友圈评论
**POST** `/api/moments/{id}/comments`

请求参数:
```json
{
  "content": "看起来很不错!"
}
```

说明:
- 评论内容 trim 后不能为空
- 成功返回 201,评论列表通过下次拉取动态接口返回

### 朋友圈 AI 文案助手
**POST** `/api/moments/ai-assist`

请求参数(根据 `mode` 二选一字段):
```json
{
  "mode": "generate",
  "prompt": "周末爬山的小感受",
  "tone": "治愈",
  "hasImage": true
}
```
或:
```json
{
  "mode": "polish",
  "content": "今天和朋友爬山看到了云海",
  "tone": "文艺",
  "hasImage": false
}
```

响应示例:
```json
{
  "text": "云在山间散步,我也学着慢下来。"
}
```

错误码:
- 400: `mode` 非法、`prompt` 或 `content` 缺失
- 503: AI 助手未配置 (`DEEPSEEK_API_KEY` 为空)
- 502: 上游 DeepSeek 调用失败或返回为空

---

## 8. WebSocket 通信协议

## 8.1 连接地址
```text
/ws
```

连接时携带 token 用于身份认证。

---

## 8.2 客户端发送事件

### 发送聊天消息
```json
{
  "type": "chat_send",
  "data": {
    "receiverId": 2,
    "content": "你好"
  }
}
```

---

## 8.3 服务端推送事件

### 收到单聊消息
```json
{
  "type": "chat_message",
  "data": {
    "id": 1001,
    "senderId": 1,
    "receiverId": 2,
    "senderCiphertext": "base64-ciphertext-for-sender",
    "senderAlgorithm": "rsa-oaep-sha256",
    "receiverCiphertext": "base64-ciphertext-for-receiver",
    "receiverAlgorithm": "rsa-oaep-sha256",
    "createdAt": "2026-05-24T10:00:00Z"
  }
}
```

说明：
- 服务端通过 WebSocket 仅推送密文，不推送明文
- 接收方前端使用本地私钥解密 `receiverCiphertext`
- 发送方本地展示已发送消息时使用本地私钥解密 `senderCiphertext`

### 收到群聊消息
```json
{
  "type": "group_message",
  "data": {
    "id": 2001,
    "groupId": 10,
    "senderId": 1,
    "contentCiphertext": "base64-aes-ciphertext",
    "contentIv": "base64-iv",
    "contentAlgorithm": "aes-gcm-256",
    "keyCiphertext": "base64-wrapped-aes-key-for-current-user",
    "keyAlgorithm": "rsa-oaep-sha256",
    "createdAt": "2026-06-02T18:00:00Z"
  }
}
```

说明：
- 服务端对每个在线群成员分别推送同一条群消息
- 每个成员收到的 `contentCiphertext` 相同，但 `keyCiphertext` 是属于该成员自己的那一份 AES 会话密钥密文
- 前端先使用本地 RSA 私钥解密 `keyCiphertext`，再用得到的 AES 密钥解密 `contentCiphertext`

### 发送确认
```json
{
  "type": "chat_ack",
  "data": {
    "tempId": "local-msg-1",
    "status": "success"
  }
}
```

### 错误事件
```json
{
  "type": "error",
  "message": "非好友关系，不能发送消息"
}
```

---

## 9. 页面设计

## 9.1 登录/注册页
功能：
- 用户注册
- 用户登录
- 登录成功后跳转主聊天页

## 9.2 主聊天页
布局建议：
- 左侧：好友列表 / 群聊列表切换区
- 右侧：聊天窗口

群聊相关补充：
- 支持“新建群聊”弹窗
- 支持“群信息”面板，用于查看成员、邀请好友、退出群聊

## 9.3 好友申请页
功能：
- 查看收到的好友申请
- 同意或拒绝申请

## 9.4 个人资料页
功能：
- 查看当前资料
- 修改昵称、性别、签名等信息
- 切换主页皮肤(`aurora`/`sunset`/`galaxy`/`mint`/`peach`)
- 展示系统默认头像，头像不可编辑

## 9.5 好友主页页
功能:
- 展示好友昵称、头像、签名,以及按其主页皮肤渲染的个性背景/装饰
- 展示该好友发布的朋友圈动态
- 仅好友之间可访问,非好友访问被拦截

## 9.6 朋友圈页
功能:
- 发布朋友圈(正文 + 图片上传)
- 调用 AI 文案助手生成或润色文案(可选语气)
- 浏览自己与好友的朋友圈时间线
- 对动态进行点赞/取消点赞
- 对动态发表评论
- 删除自己发布的动态

## 9.7 全局导航栏
- 提供"消息"、"朋友圈"、"个人资料"等入口
- 显示当前登录用户头像与昵称

---

## 10. 业务流程

## 10.1 注册登录流程
1. 用户注册账号
2. 用户登录系统
3. 后端返回 token
4. 前端保存 token 并进入主页面

## 10.2 添加好友流程
1. 用户 A 发起好友申请
2. 用户 B 查看申请
3. 用户 B 同意申请
4. 系统建立好友关系
5. A 和 B 可开始聊天

## 10.3 单聊流程
1. 用户登录后，前端生成或加载本地 RSA 私钥，并将对应公钥上传到服务端
2. 用户打开与好友的聊天窗口
3. 前端加载历史消息，并按“我是发送方/我是接收方”选择对应密文进行本地解密展示
4. 用户输入明文消息后，前端分别使用自己的公钥和好友公钥加密同一条消息
5. 前端调用消息发送接口提交双份密文
6. 后端校验、保存消息，并通过 WebSocket 将同一条双份密文消息推送给接收方
7. 发送方与接收方各自在本地使用自己的私钥解密可读密文并展示明文

## 10.4 群聊流程
1. 用户登录后，前端生成或加载本地 RSA 私钥，并将对应公钥上传到服务端
2. 用户创建群聊或加入已有群聊
3. 用户打开群聊窗口，前端拉取群详情，获得群成员列表与成员公钥
4. 前端加载群历史消息，针对每条消息取当前用户自己的 `keyCiphertext` 进行本地解密
5. 用户输入群消息后，前端为该条消息生成一次性的 AES 会话密钥和 IV
6. 前端使用该 AES 会话密钥加密消息正文，生成一份共享正文密文 `contentCiphertext`
7. 前端再使用每个群成员的 RSA 公钥分别加密这份 AES 会话密钥，生成多份 `keyCiphertext`
8. 前端调用群消息发送接口提交：共享正文密文、IV、正文算法标识，以及按成员拆分的会话密钥密文数组
9. 后端校验成员覆盖是否完整、算法是否合法后保存消息，并通过 WebSocket 分别向每个在线成员推送属于他的那一份 `keyCiphertext`
10. 各成员在本地先解密 AES 会话密钥，再解密群消息正文并展示明文

## 10.5 消息加解密流程说明
### 流程概述
系统当前包含两套端到端加密消息方案：
- **单聊**：基于 `RSA-OAEP + SHA-256` 的双密文方案
- **群聊**：基于 `AES-GCM-256 + RSA-OAEP` 的混合加密方案

两种方案下，消息明文都只在发送端与接收端浏览器本地参与加密和解密，服务端不处理消息明文，只负责保存和转发密文。

### 单聊加解密详细流程
1. **登录后的密钥准备**
   - 用户登录后，前端为当前用户生成或加载本地 RSA 密钥对。
   - 私钥仅保存在浏览器本地 `localStorage`，不上传到服务端。
   - 公钥通过 `PUT /api/users/me/public-key` 接口上传到服务端，用于被好友发送消息时加密。

2. **聊天初始化**
   - 用户进入聊天页后，前端先获取好友列表和历史消息。
   - 好友列表中需包含好友的公钥及公钥算法标识，供发送消息时使用。
   - 历史消息接口返回的是双份密文数据，不返回消息明文。

3. **消息发送加密**
   - 用户输入消息明文后，前端对同一条消息执行两次加密：
     - 使用当前用户自己的公钥加密，生成 `senderCiphertext`；
     - 使用好友公钥加密，生成 `receiverCiphertext`。
   - 两份密文分别对应发送方可解密版本与接收方可解密版本。
   - 前端将 `receiverId`、`senderCiphertext`、`senderAlgorithm`、`receiverCiphertext`、`receiverAlgorithm` 一并提交给后端。

4. **服务端保存与转发**
   - 后端仅校验请求参数完整性与算法合法性，当前支持的算法标识为 `rsa-oaep-sha256`。
   - 后端不会解密消息，也不会保存明文内容。
   - 后端将双份密文保存到消息表中，并在接收方在线时通过 WebSocket 推送该消息。

5. **消息展示与本地解密**
   - 前端展示消息时，根据当前登录用户身份选择对应密文：
     - 当前用户是发送方时，使用 `senderCiphertext` 解密；
     - 当前用户是接收方时，使用 `receiverCiphertext` 解密。
   - 解密操作全部在前端本地完成，解密成功后显示消息明文。

### 群聊加密消息发送流程（重点）
1. **群消息发送前的准备**
   - 发送者进入群聊窗口后，前端先拉取群详情。
   - 群详情中需包含所有群成员的公钥与公钥算法标识。
   - 若某位成员尚未上传公钥，则当前这条群消息不能发送。

2. **生成一次性 AES 会话密钥**
   - 发送者前端在本地为当前这条群消息临时生成一个 `AES-GCM-256` 会话密钥。
   - 同时生成一段随机 IV（初始化向量）。
   - 该 AES 会话密钥只服务于这一条群消息，不直接上传明文到服务端。

3. **使用 AES-GCM 加密群消息正文**
   - 前端使用刚生成的 AES 会话密钥对消息明文进行加密。
   - 得到一份共享的正文密文 `contentCiphertext`。
   - 因为群内所有成员看到的消息正文相同，所以正文密文只需要生成一份。

4. **为每个群成员分别包裹 AES 会话密钥**
   - 前端遍历当前群成员列表。
   - 对每个成员，使用该成员的 RSA 公钥加密同一个 AES 会话密钥原始字节。
   - 最终得到多份 `keyCiphertext`：
     - 每份 `keyCiphertext` 都只对应一个成员；
     - 每个成员只能用自己的私钥解开属于自己的那一份会话密钥密文。

5. **向服务端提交群消息密文结构**
   - 前端调用 `POST /api/groups/{id}/messages`。
   - 请求体包含：
     - `contentCiphertext`：共享正文密文
     - `contentIv`：AES-GCM 使用的随机 IV
     - `contentAlgorithm`：正文算法标识，当前为 `aes-gcm-256`
     - `memberKeys`：按群成员拆分的会话密钥密文数组，每项包含 `userId`、`keyCiphertext`、`keyAlgorithm`

6. **服务端校验与保存**
   - 后端首先校验发送方是否为当前群成员。
   - 然后校验 `memberKeys` 中的用户集合是否完整覆盖当前群成员。
   - 再校验算法标识是否符合当前支持的格式。
   - 校验通过后：
     - 在 `group_messages` 表中保存一份共享正文密文；
     - 在 `group_message_keys` 表中按成员逐条保存每一份会话密钥密文。
   - 服务端不会解密 AES 会话密钥，也不会解密群消息明文。

7. **服务端实时推送**
   - 若群成员在线，服务端通过 WebSocket 分别向每个成员推送同一条群消息。
   - 推送时：
     - `contentCiphertext`、`contentIv`、`contentAlgorithm` 对所有成员相同；
     - `keyCiphertext` 则替换为当前接收成员自己的那一份。

8. **接收方本地解密**
   - 接收方先使用本地 RSA 私钥解密自己的 `keyCiphertext`，得到这条群消息对应的 AES 会话密钥。
   - 再使用该 AES 会话密钥和 `contentIv` 解密 `contentCiphertext`，得到最终消息明文。
   - 整个解密过程仍然完全在浏览器本地完成。

9. **为什么群聊采用混合加密**
   - 若直接对群聊正文用 RSA 分别加密，会产生 N 份完整正文密文，群成员越多，数据库和网络开销越大。
   - 采用 `AES-GCM + RSA-OAEP` 混合加密后：
     - 正文密文只保存一份；
     - 仅为每个成员额外保存一份较小的 AES 会话密钥密文；
     - 更适合群聊场景，也更符合工程实践。

10. **新成员看不到入群前历史消息的原因**
   - 新成员加入群聊时，服务端不会为其补发旧消息的 `keyCiphertext`。
   - 因此，新成员即使能拿到旧的 `contentCiphertext`，也无法解出对应 AES 会话密钥。
   - 这意味着其默认不具备解密入群前历史群消息的能力。

### 异常处理
- 若本地私钥丢失、损坏或解密失败，前端不得向服务端请求明文兜底。
- 此类情况下，消息界面统一显示占位文案 `***（已加密）`，以表明该消息已加密但当前设备无法解密查看。

tips
1. 场景说明：当用户用一台新设备（新浏览器）登录后，历史的消息会被加密，严格防止了账号被盗用后的消息泄密；同样，换个新设备发送的消息，用原来的老设备也同样看不到历史的消息。
2. 群聊中新增成员默认看不到入群前消息，可作为“历史消息最小可见范围控制”的安全特性进行说明。

---

## 11. 开发阶段划分

### 第一阶段：项目初始化
- 创建 `server/` 与 `web/`
- 初始化后端和前端工程
- 配置数据库连接

### 第二阶段：用户与认证
- 注册
- 登录
- JWT 鉴权
- 个人信息接口

### 第三阶段：好友关系
- 发起申请
- 查看申请
- 同意/拒绝
- 好友列表

### 第四阶段：聊天功能
- 消息表
- 历史消息查询
- WebSocket 实时消息通信

### 第五阶段：前端联调
- 页面开发
- 接口联调
- 聊天功能联调

### 第六阶段：朋友圈与个性主页
- 朋友圈动态/点赞/评论
- 图片上传(MinIO)
- 主页皮肤选择
- 好友主页查看
- 朋友圈 AI 文案助手(DeepSeek)

### 第七阶段：测试与优化
- 接口测试
- 聊天流程测试
- 朋友圈流程测试
- 页面交互优化

---

## 12. 验收标准

系统达到以下条件即可认为完成：

- 用户可以成功注册和登录
- 用户可以查看并修改个人资料,包括切换主页皮肤
- 用户可以查看好友的公开主页
- 用户可以发起好友申请
- 用户可以同意/拒绝好友申请
- 用户可以查看好友列表
- 用户可以与好友进行实时单聊
- 用户可以创建群聊、邀请好友入群、退出群聊
- 单聊与群聊聊天记录都可以保存并查询
- 单聊与群聊消息都能以加密形式完成发送、存储与展示
- 用户可以发布、查看、点赞、评论、删除朋友圈动态
- 朋友圈仅对本人和好友可见,非好友被拦截
- 用户可在朋友圈发布页调用 AI 文案助手生成或润色文案
- 用户可上传图片并在朋友圈中正常展示
- 前后端主流程可完整演示

---

## 13. 项目功能边界说明

本项目功能范围固定为以下内容：

- 用户注册与登录
- 个人信息查看与修改（头像不可修改）
- 主页皮肤切换(`aurora`/`sunset`/`galaxy`/`mint`/`peach`)
- 好友公开主页查看
- 发起好友申请
- 同意/拒绝好友申请
- 好友列表查看
- 与好友进行实时单聊
- 创建群聊、邀请好友入群、退出群聊
- 单聊与群聊历史消息查询
- 单聊基于 RSA-OAEP 的端到端加密
- 群聊基于 AES-GCM + RSA-OAEP 的混合加密端到端消息传输
- 朋友圈图文动态发布、删除
- 朋友圈时间线、好友主页朋友圈查看
- 朋友圈点赞与评论
- 朋友圈 AI 文案助手(基于 DeepSeek)
- 基于 MinIO 的图片上传

除上述功能外,本项目暂不扩展聊天图片/文件消息、消息撤回、已读回执、多端同步、群管理、朋友圈端到端加密等其他功能。
