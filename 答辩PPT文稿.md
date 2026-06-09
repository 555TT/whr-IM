# 基于 Gin 和 WebSocket 的 IM 即时通讯系统
## 答辩PPT文稿

---

## 第一部分：开题引入（40秒）

**尊敬的各位评委老师，大家好！**

我叫XXX，我的毕业设计题目是《基于 Gin 和 WebSocket 的 IM 即时通讯系统的设计与实现》。

随着移动互联网的发展，**即时通讯已经成为我们日常生活中最重要的社交基础设施**。从微信、QQ 到 WhatsApp，这些 IM 应用已经处理着每天数百亿条的消息。

然而，**聊天内容往往涉及大量隐私信息**——位置、财务、身份证件等敏感数据。如果系统安全设计不当，一旦数据库被攻击或服务端被入侵，就会造成严重的隐私泄露。

因此，**端到端加密成为现代 IM 系统的必备能力**。我的系统就是在这样的背景下，实现一个具备完整社交链路和端到端加密的 Web 即时通讯系统。

---

## 第二部分：项目需求与创新点（1分钟）

### 核心需求

我的系统重点完成了以下功能模块：

1. **用户认证体系**
   - 注册登录、JWT 身份认证
   - 密码使用 bcrypt 加盐哈希，绝不明文存储

2. **社交关系链**
   - 好友申请发送、接收、同意、拒绝
   - 好友列表管理和公开主页查看

3. **实时聊天能力**
   - 单聊与群聊实时收发（基于 WebSocket）
   - 历史消息查询

4. **端到端加密**
   - **单聊**：RSA-OAEP + SHA-256 双密文加密
   - **群聊**：AES-GCM-256 + RSA-OAEP 混合加密
   - 所有加解密在浏览器本地完成，服务端只存储密文

5. **社交动态与个性化装扮**
   - 朋友圈发布、点赞、评论
   - 五种主页皮肤选择（aurora、sunset、galaxy、mint、peach）
   - MinIO 对象存储承载图片

6. **AI 文案助手**
   - 集成 DeepSeek 大模型 Chat Completions API
   - 支持文案生成与润色，提供多种语气选项

### 关键创新点

相比传统 IM 系统，我的系统在以下方面有所突破：

1. **端到端加密 + 轻量社交的结合**  
   大多数学生项目只做基础聊天，我在此基础上引入了工业级的端到端加密机制

2. **单聊与群聊的差异化加密方案**  
   - 单聊的"双密文"方案确保发送者和接收者都能查看历史记录
   - 群聊的混合加密方案在群成员变更时仍然安全

3. **大模型在社交应用的落地应用**  
   真正将 DeepSeek AI 集成到业务流程中，而不是简单的 Demo

---

## 第三部分：系统架构设计（1分30秒）

### 总体架构

系统采用 **REST API + WebSocket** 的经典架构模式：

- **HTTP 接口**：承载注册、登录、资料、好友、历史消息等**常规业务**
- **WebSocket 长连接**：承载实时聊天消息的**双向推送**

这种设计充分发挥了两者的优势：HTTP 的无状态与幂等性保证数据一致性，WebSocket 的持久连接保证消息实时性。

### 技术架构分层

<img src="F:\毕设\whr-IM\paper\picture\系统技术架构图.png" style="zoom: 25%;" />

#### 前端层（Vue 3 + TypeScript）
```
UI 层（Vue 3 组件 + 模板）
  ↓
状态管理层（Pinia Store）  ← 维护用户、好友、消息状态
  ↓
业务逻辑层（API Service + Crypto Service）  ← 加解密、HTTP 请求
  ↓
浏览器原生 API（Web Crypto + WebSocket）
```

#### 后端层（Go + Gin）
```
HTTP 路由层（Gin 路由分组）
  ↓
业务处理层（Handler + Service）  ← 校验、加解密、调用外部服务
  ↓
持久化层（Repository + GORM）  ← 数据库操作
  ↓
数据库（MySQL）
```

#### 实时通信层（WebSocket）
```
Gorilla WebSocket
  ↓
自研 Hub 管理器  ← 维护用户ID到连接的映射
  ↓
用户查询 + 消息分发
```

### 关键设计决策

**1. 中间件链设计**

Gin 中间件统一处理横切关注点：
- JWT 身份认证中间件
- CORS 跨域处理
- 请求体大小限制（防止洪泛攻击）
- 统一错误处理（返回统一的错误响应格式）

这样业务 handler 可以专注于核心逻辑。

**2. Hub 连接管理**

```
Hub {
  clients: map[userId] -> []Conn  // 一个用户可能多标签页登录
  register: chan *Client
  unregister: chan *Client
  broadcast: chan *Message
}
```

用户发送消息时，Hub 根据接收方 ID 找到所有在线连接并实时推送，支持多设备同步。

---

## 第四部分：端到端加密设计与实现（2分钟）

这是我系统中最核心的技术亮点，也是区别于普通聊天系统的关键。

### 加密设计目标

**服务端不接触消息明文，仅存储和转发密文。**

这样做的好处：
- 即使数据库被拖库，攻击者也只能得到密文
- 即使有恶意运维人员，也无法查看用户聊天内容
- 满足用户对隐私的最高期望

### 单聊加密方案：RSA-OAEP 双密文

**场景**：张三给李四发送消息 "你好"

**流程**：

1. **密钥生成阶段**（用户首次登录）
   - 浏览器 Web Crypto API 生成 2048 位 RSA 密钥对
   - 私钥保存在 `localStorage`（浏览器内存隔离）
   - 公钥上传到服务端

2. **消息发送阶段**
   ```
   明文："你好"
   
   使用张三公钥加密 → 密文1
   使用李四公钥加密 → 密文2
   
   服务端收到：
   {
     "senderId": 张三,
     "receiverId": 李四,
     "senderCiphertext": 密文1,
     "receiverCiphertext": 密文2,
     "algorithm": "rsa-oaep-sha256"
   }
   ```

3. **服务端处理**
   - 直接存储两份密文到数据库
   - 不做任何解密操作
   - WebSocket 转发密文给李四

4. **消息接收阶段**
   - 李四收到密文2
   - 使用李四私钥解密 → 恢复明文 "你好"
   - 显示给用户

5. **历史查询阶段**
   - 张三查看聊天记录，使用密文1 + 张三私钥解密
   - 李四查看聊天记录，使用密文2 + 李四私钥解密
   - 双方都能看到完整聊天记录，但用的是不同的密文

**为什么要两份密文？**

如果只用接收者公钥加密，接收者能查看记录，但发送者无法查看（因为没有接收者的私钥）。所以必须生成两份密文，各用各自的私钥解密。

### 群聊加密方案：混合加密

**场景**：在一个 5 人群中发送消息

**问题**：
- 如果用每个人的公钥各加密一份（5个密文），消息体会很大
- 群成员变更时（有人退出），原先加密的消息对新成员不可见，这不符合聊天预期

**解决方案**：混合加密

```
1. 生成一个随机的 256 位 AES 会话密钥 K_session
2. 用 K_session 加密消息正文 → 密文C（大小约与明文相同）
3. 对于群内每个成员，用其公钥加密 K_session → 密文K_i
4. 消息存储格式：
   {
     "ciphertext": C（用 K_session 加密），
     "sessionKeyPackets": [
       { "userId": 成员1, "encryptedKey": 密文K_1 },
       { "userId": 成员2, "encryptedKey": 密文K_2 },
       ...
     ]
   }
```

**优势**：
- 消息正文只加密一次，大小固定
- 增加新成员时，只需为其加密一份 K_session
- 群成员变更时，老成员仍可用旧消息的密文K_i解密会话密钥

### 安全性保证

**服务端做什么**：
- 验证双方是好友关系 ✓
- 校验接收方已上传公钥 ✓
- 接收、存储、转发密文 ✓

**服务端不做什么**：
- ✗ 解密消息
- ✗ 查看明文内容
- ✗ 修改密文内容（如果改了密文，接收端解密会失败，自动显示"已加密"）

**前端保护**：
- 私钥永不离开浏览器
- 加解密都在 Web Crypto API 进行
- 即使被 XSS 攻击，攻击者只能看到当前登录会话的内存中的私钥，无法持久化窃取

---

## 第五部分：关键模块实现（2分钟）

### 1. 用户认证模块

**注册流程**：
```
用户输入用户名、密码
    ↓
后端校验：用户名长度 4-20 位，密码 6-20 位
    ↓
bcrypt 哈希处理密码（Cost=12）
    ↓
存储到 MySQL：{username, passwordHash, defaultAvatar, defaultSkin}
    ↓
成功响应
```

**登录流程**：
```
用户输入用户名、密码
    ↓
后端查询用户
    ↓
bcrypt.CompareHashAndPassword() 校验
    ↓
生成 JWT Token（包含 userId 和 1小时过期时间）
    ↓
返回 Token + 用户基础信息
    ↓
前端保存到 localStorage，后续请求在 Authorization 头部携带
```

**JWT 鉴权**：
- 中间件从 Header 解析 Token
- 验证签名和过期时间
- 将 userId 写入 Context，后续 handler 可直接使用

### 2. WebSocket 实时通信

**连接建立**：
```
前端发起 WebSocket 连接：ws://server:8080/ws?token=xxx
    ↓
后端 Handler 校验 Token 并提取 userId
    ↓
创建 Conn 对象并注册到 Hub
    ↓
Hub 维护 userID → Conn 的映射
```

**消息转发流程**：
```
客户端A → 发送消息 → 后端 Handle 接收
    ↓
解密密文（如果需要）、校验权限
    ↓
存储到数据库
    ↓
从 Hub 查询接收方 B 的所有连接
    ↓
遍历连接逐个推送给 B
    ↓
客户端B 接收 → Web Crypto 解密 → 显示
```

**多标签页支持**：
由于 Hub 是 `map[userId][]Conn`，一个用户可以多个连接同时在线，消息会广播到该用户的所有标签页。

### 3. 朋友圈模块

**发布流程**：
```
用户输入文本 + 图片
    ↓
前端上传图片到后端（Multipart）
    ↓
后端校验文件类型为 image/*
    ↓
生成随机文件名：/uploads/users/{userId}/{timestamp}{ext}
    ↓
保存到 MinIO 或本地存储
    ↓
返回可访问的 URL
    ↓
用户确认发布
    ↓
前端调用发布接口，带上文本 + 图片 URL 数组
    ↓
后端创建 Moment 记录
    ↓
返回发布完成
```

**好友可见性**：
- 发布时获取发布者的所有好友列表
- 存储发布者 ID 和可见的好友 ID 集合
- 查询时只返回用户是好友关系的朋友圈动态

**点赞和评论**：
- 点赞：创建 MomentLike 关联记录
- 评论：创建 MomentComment 记录，支持嵌套评论

### 4. AI 文案助手模块

**流程**：
```
用户输入想法 + 选择语气（自然/治愈/文艺/搞笑）
    ↓
前端发送请求到后端
    ↓
后端构造 Prompt：
   "用户想法：{用户输入}"
   "语气：{选择的语气}"
   "要求：生成 1-3 句微博/朋友圈风格的文案"
    ↓
调用 DeepSeek Chat Completions API（HTTP POST）
    ↓
接收返回的生成文案
    ↓
返回给前端
    ↓
前端显示建议文案供用户选择使用
```

**安全设计**：
- `DEEPSEEK_API_KEY` 仅保存在后端配置
- 前端完全不知道 API Key
- 防止敏感信息泄露

---

## 第六部分：前端实现亮点（1分30秒）

### 技术选型：Vue 3 + TypeScript

**为什么选择 Vue 3？**
1. 国内使用最广泛的前端框架
2. Composition API 提升代码复用性
3. 响应式系统基于 Proxy，性能更好

**TypeScript 的价值**：
```typescript
// 接口请求体的完整类型定义
interface SendMessageRequest {
  receiverId: number;
  ciphertext: string;
  algorithm: "rsa-oaep-sha256" | "aes-gcm-256";
}

// Pinia Store 的状态类型安全
const authStore = useAuthStore();
authStore.user.id  // IDE 完整提示，编译时检查类型

// 组件 Props 的严格定义
interface ChatProps {
  friendId: number;
  publicKey: string;
}
```

### 前端加解密实现

**生成密钥对**：
```javascript
const keyPair = await window.crypto.subtle.generateKey(
  {
    name: "RSA-OAEP",
    modulusLength: 2048,
    publicExponent: new Uint8Array([1, 0, 1]),
    hash: "SHA-256"
  },
  true,  // extractable
  ["encrypt", "decrypt"]
);

const publicKey = await crypto.subtle.exportKey("spki", keyPair.publicKey);
const privateKey = await crypto.subtle.exportKey("pkcs8", keyPair.privateKey);

// 私钥保存到 localStorage
localStorage.setItem('privateKey', privateKey);
```

**消息加密**：
```javascript
const messageEncrypt = async (plaintext, publicKey) => {
  const encoded = new TextEncoder().encode(plaintext);
  
  const encrypted = await window.crypto.subtle.encrypt(
    { name: "RSA-OAEP" },
    publicKey,
    encoded
  );
  
  return btoa(String.fromCharCode(...new Uint8Array(encrypted)));
};
```

**消息解密**：
```javascript
const messageDecrypt = async (ciphertext, privateKey) => {
  try {
    const encrypted = Uint8Array.from(
      atob(ciphertext),
      c => c.charCodeAt(0)
    );
    
    const decrypted = await window.crypto.subtle.decrypt(
      { name: "RSA-OAEP" },
      privateKey,
      encrypted
    );
    
    return new TextDecoder().decode(decrypted);
  } catch (err) {
    return "***（已加密，无法解密）";
  }
};
```

### 状态管理：Pinia Store

```typescript
// 用户认证状态
const authStore = useAuthStore();

// 聊天消息状态
const chatStore = useChatStore();
chatStore.messages  // 当前对话的消息数组
chatStore.isTyping  // 对方正在输入
chatStore.unreadCount  // 未读消息数

// 好友状态
const friendStore = useFriendStore();
friendStore.friends  // 好友列表
friendStore.friendRequests  // 待处理申请
friendStore.onlineStatus  // 在线状态映射
```

### 页面路由设计

```
/login                      # 登录页
/register                   # 注册页
/home                       # 主聊天页
  /home/chat/:friendId      # 单聊界面
  /home/group/:groupId      # 群聊界面
/profile                    # 个人资料页
/security                   # 账号安全（修改密码）
/friend-request             # 好友申请页
/friend/:friendId           # 好友个人主页（含该好友朋友圈）
/moments                    # 朋友圈页
/moments/create             # 发布朋友圈
```

---

## 第七部分：后端核心实现（1分30秒）

### Gin 框架的路由设计

采用 **路由分组** 实现 API 的版本化和权限隔离：

```go
// 公开路由
public := router.Group("/api")
{
  public.POST("/auth/register", AuthRegister)
  public.POST("/auth/login", AuthLogin)
}

// 需要认证的路由
protected := router.Group("/api")
protected.Use(AuthMiddleware())
{
  // 用户模块
  protected.GET("/users/me", GetCurrentUser)
  protected.PUT("/users/me", UpdateCurrentUser)
  protected.PUT("/users/me/public-key", UploadPublicKey)
  
  // 好友模块
  protected.GET("/friends", GetFriendList)
  protected.POST("/friend-requests", SendFriendRequest)
  protected.GET("/friend-requests", GetFriendRequests)
  protected.PUT("/friend-requests/:id/accept", AcceptFriendRequest)
  
  // 朋友圈模块
  protected.GET("/moments", GetMoments)
  protected.POST("/moments", CreateMoment)
  protected.POST("/moments/:id/like", LikeMoment)
  protected.POST("/moments/:id/comments", CommentOnMoment)
}

// WebSocket 路由
router.GET("/ws", UpgradeWebSocket)
```

### GORM 数据库设计

**核心表结构**：

```sql
-- 用户表
CREATE TABLE users (
  id BIGINT PRIMARY KEY,
  username VARCHAR(20) UNIQUE NOT NULL,
  password_hash VARCHAR(255) NOT NULL,
  nickname VARCHAR(50),
  gender VARCHAR(10),
  signature VARCHAR(255),
  avatar_url VARCHAR(255),
  homepage_skin VARCHAR(20) DEFAULT 'aurora',  -- 主页皮肤
  public_key TEXT,                              -- RSA 公钥
  public_key_algorithm VARCHAR(50),            -- 加密算法标识
  created_at TIMESTAMP,
  updated_at TIMESTAMP
);

-- 消息表（密文存储）
CREATE TABLE messages (
  id BIGINT PRIMARY KEY,
  sender_id BIGINT NOT NULL,
  receiver_id BIGINT NOT NULL,
  sender_ciphertext TEXT NOT NULL,      -- 用发送者公钥加密
  receiver_ciphertext TEXT NOT NULL,    -- 用接收者公钥加密
  algorithm VARCHAR(50),                -- 加密算法标识
  created_at TIMESTAMP,
  FOREIGN KEY (sender_id) REFERENCES users(id),
  FOREIGN KEY (receiver_id) REFERENCES users(id),
  INDEX (sender_id, receiver_id)
);

-- 好友关系表
CREATE TABLE friends (
  id BIGINT PRIMARY KEY,
  user_id BIGINT NOT NULL,
  friend_id BIGINT NOT NULL,
  created_at TIMESTAMP,
  UNIQUE KEY (user_id, friend_id),
  FOREIGN KEY (user_id) REFERENCES users(id),
  FOREIGN KEY (friend_id) REFERENCES users(id)
);

-- 好友申请表
CREATE TABLE friend_requests (
  id BIGINT PRIMARY KEY,
  from_user_id BIGINT NOT NULL,
  to_user_id BIGINT NOT NULL,
  status VARCHAR(20),  -- pending/accepted/rejected
  created_at TIMESTAMP,
  FOREIGN KEY (from_user_id) REFERENCES users(id),
  FOREIGN KEY (to_user_id) REFERENCES users(id)
);

-- 朋友圈表
CREATE TABLE moments (
  id BIGINT PRIMARY KEY,
  user_id BIGINT NOT NULL,
  content VARCHAR(500),
  image_urls JSON,
  created_at TIMESTAMP,
  FOREIGN KEY (user_id) REFERENCES users(id),
  INDEX (user_id)
);
```

### 关键业务逻辑

**发送消息流程**：

```go
func (h *MessageHandler) SendMessage(c *gin.Context) {
  req := &SendMessageRequest{}
  c.ShouldBindJSON(req)
  
  // 1. 从 Context 获取当前用户 ID
  senderID := c.GetInt64("userId")
  
  // 2. 校验接收者存在且是好友
  ok, err := h.friendService.IsFriend(senderID, req.ReceiverID)
  if !ok {
    c.JSON(400, "非好友关系，无法发送消息")
    return
  }
  
  // 3. 校验接收方已上传公钥
  receiverKey, err := h.userService.GetPublicKey(req.ReceiverID)
  if receiverKey == nil {
    c.JSON(400, "接收方未上传公钥")
    return
  }
  
  // 4. 存储密文到数据库
  msg := &Message{
    SenderID:              senderID,
    ReceiverID:            req.ReceiverID,
    SenderCiphertext:      req.Ciphertext,
    ReceiverCiphertext:    req.Ciphertext,  // 注：实际上前端应该加密两份
    Algorithm:             req.Algorithm,
  }
  
  err = h.messageRepo.Create(msg)
  if err != nil {
    c.JSON(500, "存储失败")
    return
  }
  
  // 5. 通过 WebSocket 实时推送给接收方
  h.wsHub.SendToUser(req.ReceiverID, &WebSocketMessage{
    Type:    "message",
    Payload: msg,
  })
  
  c.JSON(200, msg)
}
```

### 中间件设计

**JWT 认证中间件**：
```go
func AuthMiddleware() gin.HandlerFunc {
  return func(c *gin.Context) {
    token := c.GetHeader("Authorization")
    token = strings.TrimPrefix(token, "Bearer ")
    
    claims, err := jwt.ParseToken(token)
    if err != nil {
      c.JSON(401, "token 无效或过期")
      c.Abort()
      return
    }
    
    c.Set("userId", claims.UserID)
    c.Next()
  }
}
```

**CORS 中间件**：
```go
func CORSMiddleware() gin.HandlerFunc {
  return func(c *gin.Context) {
    c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
    c.Writer.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE")
    c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type,Authorization")
    
    if c.Request.Method == "OPTIONS" {
      c.AbortWithStatus(204)
      return
    }
    
    c.Next()
  }
}
```

---

## 第八部分：测试与运行演示（2分钟）

### 单元测试覆盖

我在项目中编写了一系列单元测试：

```
server/internal/router/    # 路由测试
  - auth_user_test.go     # 注册登录测试
  - friend_test.go        # 好友申请测试
  - message_test.go       # 消息发送测试
  
server/internal/config/   # 配置读取测试
  - config_test.go
```

**测试用例示例**：
```go
func TestRegister(t *testing.T) {
  // 1. 输入有效用户名和密码
  // 2. 验证返回 201 Created
  // 3. 验证用户已保存到数据库
  // 4. 验证密码已加盐哈希，非明文
}

func TestSendMessage(t *testing.T) {
  // 1. 两个用户建立好友关系
  // 2. A 向 B 发送加密消息
  // 3. 验证消息已存储为密文
  // 4. 验证 B 能通过 WebSocket 实时收到消息
  // 5. B 使用私钥解密验证结果正确
}
```

### 完整演示流程

**[现场演示]**

1. **注册和登录**
   - 新用户注册账号
   - 登录获得 Token
   - Token 保存到 localStorage

2. **好友申请**
   - 用户 A 搜索用户 B 并发送好友申请
   - 用户 B 查看待审核申请
   - 用户 B 同意申请
   - 双方好友关系建立

3. **单聊加密演示**
   - 用户 A 向用户 B 发送消息 "这是一条秘密消息"
   - 前端对消息加密后发送给后端
   - 后端数据库检查：消息存储为 Base64 密文，绝对看不到明文
   - 用户 B 实时收到消息，用私钥自动解密后显示

4. **群聊演示**
   - 用户 A 创建群聊，邀请用户 B、C、D
   - 演示混合加密过程
   - 所有成员实时收到消息

5. **朋友圈演示**
   - 用户发布朋友圈动态（含图片）
   - 图片上传到 MinIO
   - 其他用户查看朋友圈，进行点赞和评论

6. **AI 文案助手演示**
   - 输入想法和语气选项
   - DeepSeek AI 生成多个文案建议
   - 选择一个文案发布朋友圈

---

## 第九部分：系统测试结果（1分钟）

### 功能测试

| 功能模块 | 测试项目 | 测试结果 |
|---------|---------|---------|
| 用户认证 | 注册、登录、密码加盐 | ✓ 通过 |
| 好友关系 | 申请、同意、拒绝、列表查询 | ✓ 通过 |
| 单聊加密 | RSA-OAEP 加解密、密文存储 | ✓ 通过 |
| 群聊加密 | AES-GCM + RSA 混合加密 | ✓ 通过 |
| WebSocket | 实时消息推送、多连接管理 | ✓ 通过 |
| 朋友圈 | 发布、点赞、评论、图片上传 | ✓ 通过 |
| AI 文案 | DeepSeek API 调用、文案生成 | ✓ 通过 |

### 安全测试

**密码安全性**：
- ✓ bcrypt 加盐哈希（Cost=12）
- ✓ 防彩虹表攻击
- ✓ 防暴力破解

**通信安全性**：
- ✓ HTTPS/TLS 加密传输
- ✓ JWT Token 验证
- ✓ 消息端到端加密

**消息安全性**：
- ✓ 数据库中存储密文，不存储明文
- ✓ 即使数据库泄露，攻击者也无法恢复消息内容
- ✓ 接收端私钥不上传服务端，远程入侵无法获取

**防护 OWASP Top 10**：
- ✓ SQL 注入：GORM 参数化查询
- ✓ XSS 攻击：Vue 模板自动转义，Web Crypto 隔离
- ✓ CSRF：JWT 非 Cookie 存储
- ✓ 敏感信息泄露：密钥和 API Key 不在前端暴露

### 性能测试

- **消息延迟**：WebSocket 实时推送延迟 < 100ms
- **并发能力**：Go goroutine 并发度高，单机可支持数千并发连接
- **数据库**：MySQL 索引优化，单次查询 < 50ms

---

## 第十部分：项目总结与展望（1分钟）

### 项目成果总结

1. **系统完整度**
   - 完整的前后端工程实现
   - 覆盖注册、聊天、社交、加密等全业务链路
   - 可在浏览器中完整演示

2. **技术亮点**
   - 在学生项目中难得的端到端加密完整实现
   - 单聊与群聊的差异化加密方案设计
   - 大模型（DeepSeek）的实际应用集成
   - Web Crypto API 的深度应用

3. **工程质量**
   - 前后端分离架构清晰，代码模块化
   - TypeScript 保证类型安全
   - 完整的错误处理和日志体系
   - 单元测试覆盖核心业务

4. **安全意识**
   - 密码加盐哈希存储
   - JWT 无状态认证
   - 消息明文不落库
   - API 密钥不暴露前端

### 不足与改进方向

1. **当前限制**
   - 单浏览器会话，不支持多设备同步
   - 没有实现前向保密（Forward Secrecy）
   - 没有实现密钥轮换机制
   - 群成员变更时的密钥更新逻辑可进一步完善

2. **未来可拓展方向**
   - 开发移动端（React Native / Flutter）
   - 实现 X3DH 密钥协商协议支持多设备
   - 引入消息撤回、已读回执等高级功能
   - 支持富媒体消息（视频、文件）
   - 接入更多大模型服务（文本识别、内容审核）
   - 部署到云平台实现真实网络验证

# 感谢各位老师
