"""第3章 系统需求分析与总体设计"""
import os
from _gen_paper import add_para, add_heading, add_image, add_page_break, DIA


def write_ch3(doc):
    add_heading(doc, '3  系统需求分析与总体设计', level=1)

    add_heading(doc, '3.1  需求分析', level=2)
    add_para(doc, '动手写代码之前，先把"系统要做什么、做到什么程度"理清楚。'
                  '下面分别从功能、非功能、用例三个角度展开。'
                  '需要说明的是，'
                  '本课题的需求并不来自外部客户，'
                  '而是从典型 IM 系统的业务闭环倒推而来——'
                  '以"用户能完成怎样的一次完整聊天"作为最高目标，'
                  '逐步拆解到每一个具体的接口与页面。'
                  '这种自顶向下的拆法对中小型项目最为合适，'
                  '能够避免把精力浪费在用不到的功能上。')

    add_heading(doc, '3.1.1  功能性需求', level=3)
    add_para(doc, '结合典型 IM 的业务闭环和本课题的加密目标，'
                  '功能性需求大致落在六个模块上。')
    add_para(doc, '用户认证：支持用户名 + 密码注册与登录；'
                  '注册时校验用户名唯一，密码必须哈希存储；'
                  '登录成功后服务端签发 JWT，前端后续请求携带令牌。')
    add_para(doc, '个人资料：登录用户可以查看与修改昵称、头像、性别、个性签名等基本信息，'
                  '并能看到自己当前的公钥是否已经就绪。')
    add_para(doc, '好友关系：通过用户名发起好友申请，对方可以选择同意或拒绝；'
                  '一旦同意，双方互相出现在好友列表中。')
    add_para(doc, '实时聊天：好友之间走 WebSocket 长连接收发消息，'
                  '系统需要妥善处理双方都在线、单方在线、双方都不在线这几种情形。')
    add_para(doc, '消息加密：用户首次进入聊天功能时，前端自动生成 RSA 密钥对——'
                  '私钥留在本地，公钥上传服务端；'
                  '发送时前端做双重加密再上传，'
                  '服务端只接收密文，接收方在浏览器里用私钥解开。')
    add_para(doc, '历史消息：用户可以回看与某位好友的聊天记录，'
                  '服务端给出密文，前端本地解密后展示。')

    add_heading(doc, '3.1.2  非功能性需求', level=3)
    add_para(doc, '非功能层面，本系统给自己设了 5 条线。'
                  '安全是第一位的——'
                  '密码必须用 bcrypt 等单向哈希保存，'
                  '所有受保护接口必须通过 JWT 鉴权，'
                  '消息明文不得在服务端落地，私钥不得上传。'
                  '实时性方面，'
                  '在双方都在线的本机/局域网环境下，'
                  '端到端延迟（点击发送→对端解密渲染）控制在 100 ms 以内。'
                  '可扩展性方面，'
                  '采用前后端分离与分层架构，'
                  '业务、数据访问、连接管理之间用清晰接口隔开，'
                  '为后续接入群聊、离线消息、已读回执等功能留出余地。'
                  '此外，前端交互要贴近常见社交软件的使用习惯——'
                  '连接状态、加密状态都应当被用户看见；'
                  '代码层面命名规范、层次分明，'
                  '关键接口要有单元测试和集成测试覆盖。')

    add_heading(doc, '3.1.3  用例分析', level=3)
    add_para(doc, '系统里只有一类典型角色——"已注册的普通用户"，'
                  '其与系统之间的交互如图 3-1 所示。'
                  '从注册、登录、查看与修改资料、上传公钥，'
                  '到发送好友申请、处理好友申请、查看好友列表，'
                  '再到建立 WebSocket 连接、收发加密消息、查看历史消息，'
                  '一张图把所有用例串起来。'
                  '这张图直接决定了后面模块怎么切、接口怎么排。')
    add_image(doc, os.path.join(DIA, '02_系统用例图.png'), '图 3-1  系统用户用例图', width_inch=5.4)
    add_para(doc, '其中"上传公钥"这个用例值得单独说一下——'
                  '它对用户透明，'
                  '完全由前端在首次进入聊天功能时自动触发，'
                  '却是整个加密体系能否工作的前提。'
                  '在用例图里把它显式画出来，'
                  '是想让读者意识到：'
                  '哪怕是看似不起眼的"自动行为"，'
                  '在系统设计阶段也值得作为独立用例对待。')

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
    add_para(doc, '数据库层选 MySQL 8.x，'
                  '保存用户、好友申请、好友关系与消息四类数据。'
                  '需要强调的是，messages 表里存的全部是密文与算法标识，'
                  '不含任何明文。')
    add_para(doc, '至于为何不引入 Redis 或消息队列——'
                  '本系统的目标用户规模在课程项目层面，'
                  'WebSocket 单机维护几千条连接 + MySQL 处理读写已经足够。'
                  '过早引入中间件会增加部署成本与论文外的依赖，'
                  '反而模糊了课题真正想讲的事——'
                  '即"如何把端到端加密做进 IM"。'
                  '换句话说，'
                  '架构的克制本身也是一种设计。')
    add_image(doc, os.path.join(DIA, '01_系统总体架构图.png'), '图 3-2  系统总体架构图', width_inch=6.0)

    add_heading(doc, '3.3  系统功能模块设计', level=2)
    add_para(doc, '按高内聚、低耦合切分，系统划成 5 个核心模块，'
                  '前后端的对应关系如图 3-3 所示。')
    add_image(doc, os.path.join(DIA, '06_模块结构图.png'), '图 3-3  前后端模块结构图', width_inch=6.0)

    add_para(doc, '用户认证模块负责注册、登录、令牌签发、令牌校验与登出，'
                  '后端落在 AuthHandler/AuthService，前端落在 LoginView、RegisterView 以及 auth store。'
                  '个人资料模块处理昵称、头像、性别、签名等基础信息的查改，'
                  '也包括公钥的上传——后端是 UserHandler/UserService，前端则是 ProfileView。')
    add_para(doc, '好友关系模块覆盖申请的发送、接收、同意、拒绝以及好友列表查询，'
                  '后端 FriendHandler/FriendService，前端 FriendRequestsView 和 FriendsView。'
                  '实时聊天模块是 HTTP 与 WebSocket 协作的舞台：'
                  'MessageHandler 负责密文入库，'
                  'MessageService 校验业务规则，'
                  'WebSocketHandler 完成实时推送。')
    add_para(doc, '消息加密模块独立成一个文件——前端 utils/e2ee.ts。'
                  '它把 RSA-OAEP 的密钥生成、导入导出、加解密、本地持久化、'
                  '历史消息密文选择这些零碎逻辑全部封装起来，'
                  '上层视图只需要"获取密钥/加密/解密"三个动作就能用上端到端加密能力。'
                  '这种集中封装让加密细节不会渗透到业务代码中，也方便后续替换算法。')

    add_heading(doc, '3.4  数据库设计', level=2)
    add_para(doc, '从需求里抽出的实体落到 4 张核心表：'
                  'users、friend_requests、friends、messages，'
                  'E-R 关系如图 3-4 所示。'
                  'users 与后三张表都是一对多——'
                  '一个用户可以发出多条好友申请、拥有多个好友、收发多条消息。')
    add_image(doc, os.path.join(DIA, '05_数据库ER图.png'), '图 3-4  数据库 E-R 图', width_inch=6.0)

    add_para(doc, 'users 表除常规字段外，额外增加 public_key 与 public_key_algorithm 两列，'
                  '保存用户在浏览器侧生成的公钥及算法标识。'
                  '其他用户向该用户发消息时，'
                  '正是从这两列拿到加密用的公钥。')
    add_para(doc, 'friend_requests 表记录申请本身——'
                  'from_user_id、to_user_id 加一个 status 标识申请处于"待处理/同意/拒绝"哪一种状态。'
                  '一旦 status 被改为"同意"，业务层会在 friends 表里插入双向记录，'
                  '保证 A 查好友能看到 B、B 查好友也能看到 A。'
                  'friends 表用 (user_id, friend_id) 联合唯一索引避免重复落库。')
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
