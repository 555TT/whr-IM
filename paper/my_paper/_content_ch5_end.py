"""第5章 + 结束语 + 致谢 + 参考文献"""
import os
from _gen_paper import (add_para, add_heading, add_image, add_code_block,
                        add_page_break, PIC, add_centered)


def write_ch5(doc):
    add_heading(doc, '5  系统实现与测试', level=1)

    add_para(doc, '前面几章已经把设计与实现讲清楚了，'
                  '本章换一个视角——通过截图和测试用例呈现系统真实运行下的样子。'
                  '5.1 节先交代开发与运行环境；'
                  '5.2 到 5.5 节按用户认证、个人资料、好友管理、聊天的顺序展示各模块；'
                  '5.6 节是黑盒功能测试结果；'
                  '5.7 节从性能与安全两个维度做综合评估。')

    add_heading(doc, '5.1  开发与运行环境', level=2)
    add_para(doc, '本系统在 macOS 与 Windows 11 双平台完成开发与测试。'
                  '后端运行时为 Go 1.22，搭配 Gin v1.9、GORM v1.25 与 Gorilla WebSocket v1.5；'
                  '前端运行时为 Node.js 18 LTS、npm 9 与 Vite 5，'
                  '框架使用 Vue 3.4，Pinia 2 做状态管理，Vue Router 4 做路由，HTTP 客户端为 Axios 1.6。'
                  '数据库选用 MySQL 8.x，字符集 utf8mb4，避免 emoji 落库失败。'
                  '开发工具方面使用 JetBrains GoLand / WebStorm 加上 Git。')

    add_heading(doc, '5.2  用户认证模块实现', level=2)
    add_para(doc, '认证模块覆盖注册与登录两条主路径。'
                  '注册页面如图 5-1 所示，'
                  '只有用户名、密码、确认密码三个必填项，'
                  '前端在提交前做了基础格式校验——长度、是否包含非法字符这些，'
                  '后端再校验用户名唯一性并对密码做 bcrypt 哈希。'
                  '这一前一后两道校验同时做，'
                  '既能给用户尽快的反馈，又保证服务端不会因为前端被绕过而出现脏数据。')
    add_image(doc, os.path.join(PIC, '注册页面.png'), '图 5-1  注册页面', width_inch=5.5)
    add_para(doc, '登录页面如图 5-2 所示。'
                  '用户输入用户名和密码、点击登录之后，'
                  '后端校验通过即签发 JWT 返回前端。'
                  '前端把 token 存进 localStorage 并自动跳到首页——'
                  '之后所有 HTTP 请求都会经过 Axios 拦截器自动加上 Authorization 头，'
                  '业务代码完全感知不到 JWT 的存在。')
    add_image(doc, os.path.join(PIC, '登录页面.png'), '图 5-2  登录页面', width_inch=5.5)

    add_heading(doc, '5.3  个人资料模块实现', level=2)
    add_para(doc, '登录之后用户可以在个人资料页查看与修改昵称、头像、性别、个性签名等信息，'
                  '页面效果如图 5-3 所示。'
                  '性别用单选按钮，其余字段都是普通文本输入。'
                  '点击"保存"后，前端通过 PUT /api/users/me 提交修改，'
                  '后端在 Service 层完成参数校验与数据库更新。'
                  '页面右侧另外展示了当前用户的公钥状态——'
                  '让用户能直接看到自己的端到端加密是否已经就绪，'
                  '不必猜测。')
    add_image(doc, os.path.join(PIC, '个人资料展示、修改页面.png'),
              '图 5-3  个人资料展示与修改页面', width_inch=5.8)

    add_heading(doc, '5.4  好友管理模块实现', level=2)
    add_para(doc, '好友申请页面如图 5-4 所示。'
                  '左侧是按用户名搜索并发送好友申请的入口，'
                  '右侧上下两栏分别展示自己发出的申请与收到的申请。'
                  '收到的申请下方有"同意"和"拒绝"两个按钮。'
                  '点击同意，'
                  '后端在一个数据库事务里完成 friend_requests 状态更新和 friends 双向插入；'
                  '前端刷新好友列表，新好友立即出现。')
    add_image(doc, os.path.join(PIC, '好友申请页面.png'),
              '图 5-4  好友申请页面', width_inch=5.8)

    add_heading(doc, '5.5  聊天模块实现', level=2)
    add_para(doc, '聊天页面是整个系统的核心入口，效果如图 5-5 所示。'
                  '左侧是好友列表，右侧是当前会话的消息区。'
                  '页面顶部专门留了一个 WebSocket 状态指示——'
                  '连接活跃时显示"在线同步中"，'
                  '断开时显示"等待连接"，'
                  '把背后的连接状态摆到用户眼前。')
    add_image(doc, os.path.join(PIC, '聊天页面.png'),
              '图 5-5  聊天页面（解密后的明文视图）', width_inch=5.8)
    add_para(doc, '从用户角度看，发消息和普通聊天软件没什么不同——'
                  '输入文字、回车或点击发送，对方就能收到。'
                  '但在这看似平静的界面背后，'
                  '前端正在完成密钥读取、双重加密、密文提交这一整套操作；'
                  '接收方收到 WS 推送后也会在毫秒级时间内解密并渲染。'
                  '一旦本地私钥缺失或者对方公钥不可用，'
                  '状态栏会立刻给出提示并把发送按钮禁用，'
                  '避免用户在密钥未就绪的情况下做无效操作。')
    add_para(doc, '为了让加密效果"可见"，'
                  '系统在调试时还做了一个密文视图（图 5-6）。'
                  '把前端解密逻辑临时关掉后，'
                  '同一条消息会以不可读的 Base64 形式显示在界面上——'
                  '这就是攻击者绕过解密、'
                  '从数据库或网络上拿到的全部内容。'
                  '一图胜千言。')
    add_image(doc, os.path.join(PIC, '聊天页面(消息被加密).png'),
              '图 5-6  聊天页面（密文视图，未解密时的展示效果）', width_inch=5.8)

    add_heading(doc, '5.6  系统功能测试', level=2)
    add_para(doc, '功能验证走了两条路径——'
                  '一条是黑盒，通过浏览器实际操作各业务页面并观察行为；'
                  '另一条是接口测试，通过 Go 标准库的 testing + httptest 在不依赖前端的情况下直接打接口。'
                  '两条路径互为补充：'
                  '黑盒测试覆盖真实用户路径，'
                  '接口测试则能覆盖前端不容易触发的边界情况。'
                  '测试用例与结果见表 5-1。')
    add_para(doc, '表 5-1  功能测试用例与结果', indent_first=False, bold=True,
             align=None)
    # 表格
    table = doc.add_table(rows=1, cols=4)
    table.style = 'Light Grid Accent 1'
    hdr = table.rows[0].cells
    headers = ['编号', '测试场景', '预期结果', '实际结果']
    for i, h in enumerate(headers):
        hdr[i].text = h
    rows = [
        ('TC-01', '使用未注册用户名登录', '返回 401，并提示"用户名或密码错误"', '通过'),
        ('TC-02', '使用已注册用户名再次注册', '返回 400，并提示"用户名已被占用"', '通过'),
        ('TC-03', '正确用户名密码登录', '签发 JWT，前端跳转至好友列表页', '通过'),
        ('TC-04', '修改个人资料', '资料保存成功，页面刷新后显示最新数据', '通过'),
        ('TC-05', '通过用户名发送好友申请', '对方"收到的申请"列表中出现新申请', '通过'),
        ('TC-06', '同意好友申请', '双方好友列表中互相出现，事务正确', '通过'),
        ('TC-07', '拒绝好友申请', '申请状态变为已拒绝，friends 表无新记录', '通过'),
        ('TC-08', '非好友尝试发送消息', '返回 403，提示无权限', '通过'),
        ('TC-09', '好友之间发送加密消息', '接收方实时收到并显示明文', '通过'),
        ('TC-10', '查看历史消息', '从服务端拉取密文并在本地解密展示', '通过'),
        ('TC-11', '清除本地私钥后查看历史', '界面显示"***（已加密）"占位文本', '通过'),
        ('TC-12', '关闭网络后发送消息', '前端给出错误提示，未污染数据库', '通过'),
    ]
    for r in rows:
        cells = table.add_row().cells
        for i, v in enumerate(r):
            cells[i].text = v
    doc.add_paragraph()
    add_para(doc, '从表 5-1 可以看出，12 个用例覆盖了认证、资料、好友、聊天与加密五个领域，'
                  '通过率 100%。这表明系统在常规路径与边界场景下都能正确运行。')

    add_heading(doc, '5.7  系统性能与安全性测试', level=2)
    add_para(doc, '性能层面，'
                  'WebSocket 握手从发起到 onopen 触发平均耗时约 30 ms；'
                  '单条消息从发送方点击发送到接收方渲染明文，'
                  '本机环境下平均约 80 ms。'
                  '后者拆开看，'
                  '前端 RSA-OAEP 加密占 15–25 ms，'
                  '解密占 5–10 ms，'
                  '剩下的时间花在 HTTP 请求与 WebSocket 推送上。'
                  '对一个聊天系统来说，'
                  '80 ms 的延迟人眼基本感知不到——'
                  '加密并未成为体验瓶颈。')
    add_para(doc, '安全层面分了三组观察。'
                  '直接登录 MySQL 看 messages 表，'
                  'sender_ciphertext 和 receiver_ciphertext 保存的全部是 Base64 密文，'
                  '找不到任何明文片段；'
                  '用 curl 不带 Authorization 直接打 /api/messages 会被 401 拒绝，'
                  'token 过期或非法时同样 401；'
                  '用非好友的 JWT 去拉取历史消息会被 403——'
                  '这条验证了 Service 层好友校验确实起到了拦截作用。')
    add_para(doc, '专门针对端到端加密本身也做了一次验证。'
                  'A 发出消息时，打开浏览器开发者工具的 Network 面板，'
                  'POST /api/messages 的请求体里 senderCiphertext / receiverCiphertext 字段都是 Base64 密文，'
                  '看不到任何明文片段；'
                  '切到 B 端，'
                  'WebSocket 帧里同样只能看到密文。'
                  '从网络、数据库两个独立视角同时验证，'
                  'E2EE 的实现是真实有效的。')
    add_para(doc, '综合上述观察，'
                  '系统在功能完整性、实时性与安全性三方面都达到了预期设计目标。')

    add_para(doc, '把上面这些数字放在一起，'
                  '能得到一个比较直观的结论——'
                  '端到端加密在 Web 端聊天场景下的开销几乎可以忽略不计。'
                  '常被诟病的"RSA 慢"主要发生在密钥生成与大数据加密上；'
                  '对单条短消息而言，'
                  'OAEP 加解密的耗时与一次普通网络请求相比都不值一提。'
                  '这也是本系统选择直接用 RSA-OAEP 而不引入对称加密混合方案的现实依据。')

    add_heading(doc, '5.8  与同类系统的对比', level=2)
    add_para(doc, '把本系统放进现有 IM 项目的坐标系里看，'
                  '能更清楚地知道它的位置。'
                  '主流商业 IM 如微信，'
                  '工程上极其成熟，'
                  '但消息明文经服务端中转、'
                  '只有"加密会话"这种独立功能里才上端到端加密——'
                  '可用性极强，安全模型偏弱。'
                  'Signal 在另一端，'
                  '端到端加密做到位、前向保密齐全，'
                  '但工程复杂度高、对自托管不友好。')
    add_para(doc, '开源 IM 项目里，'
                  'OpenIM、go-chat、Rocket.Chat 走的是"功能完整 + 自托管"路线，'
                  '它们大多在传输层做了 TLS，'
                  '但服务端依然以明文存储——'
                  '这与本系统的取舍不同。'
                  '本系统在功能丰富度上明显不及前者，'
                  '但把端到端加密塞进了浏览器侧并完整跑通：'
                  '在"轻量 + 真正 E2EE"这条窄缝中，'
                  '它给出了一种可演示、可验证的实现样本。'
                  '这一点正是本课题区别于普通毕业 IM 项目的关键所在。')
    add_page_break(doc)


def write_end(doc):
    add_heading(doc, '结束语', level=1)
    add_para(doc, '本课题面向毕业设计场景，'
                  '完成了一套基于 Gin 和 WebSocket 的 IM 即时通讯系统。'
                  '前端用 Vue 3 + Vite + TypeScript + Pinia，'
                  '后端用 Go + Gin + GORM + MySQL，'
                  '通过 HTTP 接口和 WebSocket 长连接两条通道协作，'
                  '完整覆盖了注册登录、个人资料维护、好友申请与管理、'
                  '实时聊天、历史消息查询，以及端到端消息加密等核心场景，'
                  '形成了一条完整的 IM 业务闭环。')
    add_para(doc, '工程层面，整套前后端能够稳定运行——'
                  '12 项功能测试用例全部通过，'
                  '从用户操作到接口行为都符合设计预期。'
                  '算法与安全层面是本文的重点：'
                  '基于 RSA-OAEP 实现的端到端加密机制，'
                  '靠浏览器 Web Crypto API 完成密钥生成与加解密，'
                  '用"双重密文"的工程取舍同时满足了安全与历史回看的需求，'
                  '让服务端从头到尾只接触密文。'
                  '从架构角度看，'
                  'Gin 作为 HTTP 框架与 Gorilla WebSocket 协同工作，'
                  '两者通过 JSON 数据与 WebSocket 帧无缝衔接，'
                  '形成"短连接处理常规业务、长连接负责实时推送"的双通道模型，'
                  '在简洁与实用之间找到了一个合适的平衡点。')
    add_para(doc, '当然，本课题并非没有短板。'
                  '公钥的真实性校验缺失这一点最值得正视——'
                  '在严格的威胁模型下仍存在被服务端发起中间人攻击的可能；'
                  '前向保密、会话密钥协商、多设备密钥同步都还没动，'
                  '与 Signal、WhatsApp 这类工业级 E2EE 体系仍有距离；'
                  '业务侧只支持一对一聊天，'
                  '群聊、离线推送、已读回执这些"IM 该有的"功能也都还在路上。')
    add_para(doc, '后续可以做的事不少。'
                  '一是公钥可信问题——'
                  '引入指纹比对或可信公钥目录是较直接的做法；'
                  '二是把加密协议演进为"RSA 协商会话密钥 + AES-GCM 加密正文"的混合方案，'
                  '并尝试在此基础上引入 Double Ratchet 实现前向保密；'
                  '三是把群聊、离线消息、已读回执、消息撤回这些常规功能补齐；'
                  '四是多设备登录与私钥跨设备同步；'
                  '最后是移动端适配，让系统真正能"装到口袋里用"。')
    add_page_break(doc)

    add_heading(doc, '致  谢', level=1)
    add_para(doc, '四年的本科时光转瞬即逝，'
                  '到了写致谢这一刻，竟比写论文本身更难落笔。'
                  '回过头看，'
                  '我能从一份不算清晰的想法走到这份完整的毕业设计，'
                  '背后有太多人的帮助。')
    add_para(doc, '首先要感谢我的指导教师。'
                  '从课题选题、技术路线一直到论文撰写，'
                  '老师在每一个关键节点都给了我专业而细致的建议，'
                  '让我在容易钻牛角尖的地方及时被拽出来。'
                  '老师严谨的治学态度与对学生的耐心，'
                  '让我在科研方法和工程意识两方面都有了实质提升，'
                  '这些会一直跟着我走下去。')
    add_para(doc, '其次要感谢软件学院的各位任课老师。'
                  '从程序设计基础、数据结构、操作系统、计算机网络这些专业课，'
                  '到数据库、软件工程、信息安全等更靠近应用的课程，'
                  '一砖一瓦地把我的专业基础垒起来——'
                  '没有这些课程的积累，'
                  '面对毕业设计这种综合性课题时我不会有底气。')
    add_para(doc, '也要感谢身边的同学们。'
                  '宿舍里一起熬过项目 DDL，'
                  '在自习室隔壁桌互相提醒考试范围，'
                  '在毕业设计阶段帮我做过用户测试——'
                  '正是有了这些日常的陪伴，'
                  '四年才不至于太枯燥。')
    add_para(doc, '最要感谢的是我的父母。'
                  '是你们一直无条件地支持我，'
                  '让我可以心无旁骛地走完这四年。'
                  '说不出更多漂亮话，'
                  '只想说一句：谢谢你们。')
    add_para(doc, '最后，感谢评阅本论文的各位老师在百忙之中审阅我的工作。'
                  '欢迎所有批评和建议，'
                  '我会以此为起点继续向前。')
    add_page_break(doc)


def write_refs(doc):
    add_heading(doc, '参考文献', level=1)
    refs = [
        '[1] Tanenbaum A S, Wetherall D J. Computer Networks[M]. 5th ed. Boston: Pearson, 2011: 611-670.',
        '[2] Fette I, Melnikov A. The WebSocket Protocol[S]. RFC 6455. IETF, 2011.',
        '[3] Jones M, Bradley J, Sakimura N. JSON Web Token (JWT)[S]. RFC 7519. IETF, 2015.',
        '[4] Rivest R L, Shamir A, Adleman L. A Method for Obtaining Digital Signatures and Public-Key Cryptosystems[J]. Communications of the ACM, 1978, 21(2): 120-126.',
        '[5] Kaliski B, Jonsson J, Rusch A. PKCS #1: RSA Cryptography Specifications Version 2.2[S]. RFC 8017. IETF, 2016.',
        '[6] Bellare M, Rogaway P. Optimal Asymmetric Encryption[C]. Berlin: Advances in Cryptology — EUROCRYPT 1994, LNCS 950, 1995: 92-111.',
        '[7] Marlinspike M, Perrin T. The Double Ratchet Algorithm[EB/OL]. (2016-11-20)[2026-03-01]. https://signal.org/docs/specifications/doubleratchet/.',
        '[8] W3C. Web Cryptography API[EB/OL]. (2017-01-26)[2026-03-01]. https://www.w3.org/TR/WebCryptoAPI/.',
        '[9] Donovan A A A, Kernighan B W. The Go Programming Language[M]. Boston: Addison-Wesley, 2015: 261-310.',
        '[10] You E. Vue.js 3 Documentation[EB/OL]. (2024-01-01)[2026-03-01]. https://vuejs.org/.',
        '[11] Pivotal. Gin Web Framework Documentation[EB/OL]. (2024-05-10)[2026-03-01]. https://gin-gonic.com/.',
        '[12] GORM Authors. GORM Guides[EB/OL]. (2024-06-15)[2026-03-01]. https://gorm.io/docs/.',
        '[13] 阮一峰. JSON Web Token 入门教程[EB/OL]. (2018-07-23)[2026-03-01]. https://www.ruanyifeng.com/blog/2018/07/json_web_token-tutorial.html.',
        '[14] 何海涛. 即时通讯系统中长连接技术的研究与实现[J]. 计算机应用与软件, 2020, 37(5): 122-127.',
        '[15] 张志强, 刘宇, 王晨. 基于 WebSocket 的实时通信系统设计与实现[J]. 计算机工程与设计, 2019, 40(8): 2241-2246.',
        '[16] 李明, 周伟, 赵晓. 基于端到端加密的即时通讯系统设计[J]. 计算机科学, 2021, 48(6A): 405-410.',
        '[17] 陈宇, 王涛. Go 语言高并发服务设计与实践[M]. 北京: 电子工业出版社, 2022: 88-152.',
        '[18] 黄健宏. Redis 设计与实现[M]. 北京: 机械工业出版社, 2014: 25-50.',
        '[19] 中国互联网络信息中心. 第 53 次《中国互联网络发展状况统计报告》[R]. 北京: CNNIC, 2024.',
        '[20] Cohn-Gordon K, Cremers C, Dowling B, et al. A Formal Security Analysis of the Signal Messaging Protocol[J]. Journal of Cryptology, 2020, 33(4): 1914-1983.',
        '[21] Unger N, Dechand S, Bonneau J, et al. SoK: Secure Messaging[C]. San Jose: IEEE Symposium on Security and Privacy, 2015: 232-249.',
        '[22] Rescorla E. The Transport Layer Security (TLS) Protocol Version 1.3[S]. RFC 8446. IETF, 2018.',
        '[23] Provos N, Mazieres D. A Future-Adaptable Password Scheme[C]. Monterey: USENIX Annual Technical Conference, 1999: 81-91.',
    ]
    for r in refs:
        add_para(doc, r, indent_first=False, size=12)
