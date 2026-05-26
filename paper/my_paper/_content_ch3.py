"""第3章 系统需求分析与总体设计"""
import os
from _gen_paper import add_para, add_heading, add_image, add_page_break, DIA


def write_ch3(doc):
    add_heading(doc, '3  系统需求分析与总体设计', level=1)

    add_heading(doc, '3.1  需求分析', level=2)
    add_para(doc, '写代码之前先要把"系统要做什么、做到什么程度"想清楚。'
                  '下面分别从功能、非功能和用例三个方向展开。'
                  '需要交代一句：本课题没有外部客户，'
                  '需求是从典型 IM 系统的业务闭环倒推出来的——'
                  '以"用户能完成一次完整聊天"作为最高目标，'
                  '再逐步拆到每一个具体接口和页面。'
                  '这种自顶向下的拆法在中小型项目里最实用，'
                  '能避免把精力浪费在用不上的功能上。')

    add_heading(doc, '3.1.1  功能性需求', level=3)
    add_para(doc, '结合典型 IM 的业务闭环和本课题的加密目标，'
                  '功能性需求大致落在六个模块上。')
    add_para(doc, '用户认证：注册和登录都用"用户名 + 密码"走，'
                  '注册时校验用户名唯一，密码必须哈希后再落库；'
                  '登录成功后由服务端签发 JWT，前端后续请求把令牌带上即可。')
    add_para(doc, '个人资料：登录用户可以查看与修改昵称、头像、性别、个性签名等基本信息，'
                  '并能看到自己当前的公钥是否已经就绪。')
    add_para(doc, '好友关系：通过用户名发起好友申请，对方可以选择同意或拒绝；'
                  '一旦同意，双方互相出现在好友列表中。')
    add_para(doc, '实时聊天：好友之间走 WebSocket 长连接收发消息，'
                  '系统需要妥善处理双方都在线、单方在线、双方都不在线这几种情形。')
    add_para(doc, '消息加密：用户第一次进入聊天功能时，前端自动生成 RSA 密钥对'
                  '——私钥留在本地，公钥推到服务端；'
                  '发送一条消息时前端做双重加密后上传，'
                  '服务端只接收密文，接收方在浏览器里用私钥解开。')
    add_para(doc, '历史消息：用户可以回看与某位好友的聊天记录，'
                  '服务端给出密文，前端本地解密后展示。')

    add_heading(doc, '3.1.2  非功能性需求', level=3)
    add_para(doc, '几条非功能性约束不是同等分量放在那里的——把它们排好序很重要。'
                  '安全那条最不能让步：密码绝不能明文落库（用 bcrypt 这类单向哈希过一遍是底线），'
                  '受保护接口一律走 JWT，消息明文不允许出现在服务端任何位置，私钥则永远只能留在浏览器本地。'
                  '实时性紧跟其后，目标比较具体——双方在线、网络正常的本机或局域网环境下，'
                  '从点击发送到对端解密并把明文渲染出来，整条链路最好压在 100 ms 以内。'
                  '可扩展性的考量主要落在结构上：前后端分离 + 分层架构，模块之间用清晰的接口隔开，'
                  '日后想加群聊、离线消息、已读回执，至少不必把现有代码推倒重来。'
                  '剩下的两条权重更低些。一个是用户体验——WebSocket 连接状态、端到端加密状态这种"系统内部状态"'
                  '应当被用户看见，而不是闷在底层；另一个是代码本身，命名、分层这些约束之外，关键接口要尽量有单元测试或集成测试盯着。')

    add_heading(doc, '3.1.3  用例分析', level=3)
    add_para(doc, '本系统里只有一类典型角色——"已注册的普通用户"，'
                  '他与系统之间的交互如图 3-1 所示。'
                  '注册、登录、查看与修改资料、上传公钥；'
                  '发送和处理好友申请、查看好友列表；'
                  '建立 WebSocket 连接、收发加密消息、查看历史消息——'
                  '一张图把所有用例串起来了。'
                  '后面模块怎么切、接口怎么排，几乎都是从这张图里读出来的。')
    add_image(doc, os.path.join(DIA, '02_系统用例图.png'), '图 3-1  系统用户用例图', width_inch=5.4)
    add_para(doc, '其中"上传公钥"这个用例值得单独说一句——'
                  '它对用户是完全透明的，'
                  '首次进入聊天功能时由前端自动触发，'
                  '可它又是整套加密体系能否工作的前提。'
                  '在用例图里把它显式画出来，'
                  '是想提醒读者：'
                  '哪怕是看起来不起眼的"自动行为"，'
                  '在系统设计阶段也应当被当作独立用例对待。')

    add_heading(doc, '3.2  系统总体架构', level=2)
    add_para(doc, '总体架构沿用最典型的"前端 + 后端 + 数据库"三段式，'
                  '如图 3-2 所示。前端跑在浏览器里，'
                  '业务请求走 HTTP、实时消息走 WebSocket，'
                  '加解密由 Web Crypto API 在本地完成。')
    add_para(doc, '后端是这套系统真正的"主战场"。'
                  '从外到内分为六层——'
                  '路由层把 HTTP 请求和 WebSocket 升级请求分发到 Handler；'
                  '中间件层统一处理 CORS、JWT 鉴权与日志；'
                  'Handler 层做参数绑定与响应组装；'
                  'Service 层承载真正的业务逻辑；'
                  'Repository 层借助 GORM 与数据库打交道；'
                  'WebSocket Hub 层独立出来，'
                  '专门维护在线连接并按用户 ID 推送下行消息。'
                  '各层职责清晰，调用关系单向向下，避免循环依赖。')
    add_para(doc, '数据库选 MySQL 8.x，保存用户、好友申请、好友关系与消息四类数据。'
                  '有一点要强调：messages 表里存的全部是密文加上算法标识，'
                  '没有任何明文。')
    add_para(doc, '为什么不引入 Redis 或消息队列？'
                  '本系统的目标用户规模在课程项目量级，'
                  'WebSocket 单机维护几千条连接、MySQL 处理读写完全够用。'
                  '过早引入中间件只会增加部署成本和论文外的额外依赖，'
                  '反而模糊本课题真正想讲的事——"把端到端加密做进 IM"。'
                  '换个说法，架构上的克制本身就是一种设计选择。')
    add_image(doc, os.path.join(DIA, '01_系统总体架构图.png'), '图 3-2  系统总体架构图', width_inch=6.0)

    add_heading(doc, '3.3  系统功能模块设计', level=2)
    add_para(doc, '按高内聚、低耦合切分，系统划成 5 个核心模块，'
                  '前后端的对应关系如图 3-3 所示。')
    add_image(doc, os.path.join(DIA, '06_模块结构图.png'), '图 3-3  前后端模块结构图', width_inch=6.0)

    add_para(doc, '用户认证模块承担注册到登出整条链路上跟 token 有关的活：令牌签发、令牌校验，包括登出时怎么处置它。'
                  '后端代码落在 AuthHandler 与 AuthService 这两层，前端则分布在 LoginView、RegisterView 以及 auth store 里。'
                  '个人资料模块的边界比较窄，做的是昵称、头像、性别、签名这些字段的读写，再加上一件不那么显眼的事——把用户公钥推到服务端。'
                  '它由后端的 UserHandler / UserService 配合前端 ProfileView 完成。')
    add_para(doc, '好友关系模块跟"申请"这件事强相关，发申请、接申请、同意或拒绝、再到拉取自己的好友列表都归它管，'
                  '对应后端 FriendHandler / FriendService 与前端 FriendRequestsView、FriendsView。实时聊天模块的情况要复杂一些，'
                  '它实际上是 HTTP 通道和 WebSocket 通道协作的地方：密文走 HTTP 由 MessageHandler 写入数据库，'
                  '业务规则的校验放在 MessageService，实时推送则交给独立的 WebSocketHandler。')
    add_para(doc, '消息加密相关逻辑被单独抽到一个文件里——前端 utils/e2ee.ts。'
                  '它把 RSA-OAEP 的密钥生成、导入导出、加解密、本地持久化、'
                  '历史消息密文选择这些零碎逻辑全部封装起来，'
                  '上层视图只需要"获取密钥、加密、解密"这三个动作就能用上端到端加密能力。'
                  '这样做的好处是加密细节不会渗透进业务代码，将来想换算法也更省事。')

    add_heading(doc, '3.4  数据库设计', level=2)
    add_para(doc, '从需求里抽出的实体落到 4 张核心表：'
                  'users、friend_requests、friends、messages，'
                  'E-R 关系如图 3-4 所示。'
                  'users 与后三张表都是一对多——'
                  '一个用户可以发出多条好友申请、拥有多个好友、收发多条消息。')
    add_image(doc, os.path.join(DIA, '05_数据库ER图.png'), '图 3-4  数据库 E-R 图', width_inch=6.0)

    add_para(doc, 'users 表里除了账户必需的几个字段，专门多挂了 public_key 和 public_key_algorithm 两列，'
                  '存的就是用户在自己浏览器里生成出来的那把公钥，以及它对应的算法名。'
                  '别人想给这位用户发加密消息时，加密用的公钥就是从这里取的——没有这两列，整条加密链路就接不起来。')
    add_para(doc, 'friend_requests 表存放申请本身的信息，结构很直白——from_user_id、to_user_id，再加一个 status 标记现在是处于'
                  '"待处理""同意"还是"拒绝"中的哪一种。当 status 被翻成"同意"那一刻，业务层会顺势往 friends 表里插两条对称记录，'
                  '这样无论 A 查自己的好友、还是 B 查自己的好友，对方都能正常出现，不必在 SQL 里拼 OR。重复落库这件事则交给 '
                  '(user_id, friend_id) 联合唯一索引兜底。')
    add_para(doc, '索引方面专门做过取舍。'
                  'users 给 username 加唯一索引，登录与防重复注册一并解决；'
                  'friend_requests 给 (to_user_id, status) 加复合索引，'
                  '专门为"我收到的待处理申请"这条高频查询服务；'
                  'friends 给 user_id 建普通索引，加速好友列表；'
                  'messages 建 (sender_id, receiver_id, created_at) 三列复合索引，'
                  '把历史消息的范围扫描压在 O(log n) 量级。'
                  '索引不是越多越好，写多读少的字段就没必要建——这一点在落地时也做了权衡。')
    add_para(doc, 'messages 表是整个安全设计的落地点。'
                  '除 sender_id、receiver_id、created_at 外，'
                  '专门设了 sender_ciphertext / sender_algorithm 与 receiver_ciphertext / receiver_algorithm 两组字段。'
                  'sender_ciphertext 用发送方公钥加密，'
                  'receiver_ciphertext 用接收方公钥加密——'
                  '同一段明文加密两次，'
                  '收发双方各拿一份能用自己私钥解开的副本，'
                  '历史消息双向可读的同时，服务端依然看不见明文。')
    add_page_break(doc)
