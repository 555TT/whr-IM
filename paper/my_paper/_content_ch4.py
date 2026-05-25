"""第4章 关键技术实现（重点：WebSocket + 加密）"""
import os
from _gen_paper import add_para, add_heading, add_image, add_code_block, add_page_break, DIA


def write_ch4(doc):
    add_heading(doc, '4  关键技术实现', level=1)
    add_para(doc, '4.1 节先说用户认证与 JWT，'
                  '4.2 节介绍好友关系的处理逻辑，'
                  '4.3 节进入 WebSocket 实时聊天的实现细节，'
                  '4.4 节落到端到端加密——'
                  '后两节是本文的核心亮点。')

    add_para(doc, '需要说明的是，4.1 与 4.2 节选用的是相对常规的实现路径——'
                  '它们更多展示系统的"基础设施"如何搭起来；'
                  '真正体现本文价值的，'
                  '是 4.3 节里 WebSocket 的连接管理和 4.4 节里那套端到端加密的设计取舍。')

    add_heading(doc, '4.1  用户认证与 JWT 实现', level=2)
    add_para(doc, '用户认证是整个系统的入口，'
                  '所有后续业务都建立在"这条请求来自哪个用户"这一前提之上。'
                  '后端把注册和登录拆成 AuthService 里的两个独立方法。'
                  '密码不能明文保存，这是底线；'
                  '本系统选 bcrypt——'
                  '它内置盐值，并允许调节工作因子，'
                  '能让离线穷举攻击的成本随硬件升级而水涨船高，'
                  '比裸的 MD5 或 SHA1 稳健得多。')
    add_para(doc, '登录通过后，服务端用 HMAC-SHA256 签发一个 JWT。'
                  'Payload 里塞着用户 ID 和过期时间戳，'
                  '签名密钥只保留在服务端配置里，绝不下发。'
                  '前端拿到 Token 之后存进 localStorage，'
                  'HTTP 请求里走 "Authorization: Bearer <token>" 头，'
                  'WebSocket 握手则改走 URL 查询参数——'
                  '两条路径同一份令牌，验签逻辑完全复用。')
    add_para(doc, '受保护接口的鉴权交给 AuthMiddleware 统一处理。'
                  '中间件取出请求头里的 JWT，用同样的密钥验签：'
                  '过期、伪造或缺失，全部直接 401 拒绝；'
                  '验签通过则把用户 ID 放进 gin.Context，'
                  '业务 Handler 只管 c.Get("userId") 就行。'
                  '这种"一道关卡守住全部接口"的写法，'
                  '让 Handler 里几乎看不到任何鉴权相关的样板代码。')
    add_para(doc, '关于 bcrypt 的工作因子，'
                  '本系统取默认值 10——'
                  '即一次哈希约耗时 60–80 ms。'
                  '这个量级对登录接口完全可接受，'
                  '但对暴力破解者则意味着每秒只能尝试十几次，'
                  '与 MD5 数十亿次/秒的速度形成压倒性差距。'
                  '一旦 GPU 算力升级，'
                  '只需把工作因子改成 11 或 12 重新加密，'
                  '系统的防护强度就能跟上时代。')

    add_para(doc, '关键中间件代码如下：')
    add_code_block(doc, '''func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        tokenStr := strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer ")
        claims, err := ParseJWT(tokenStr)
        if err != nil {
            c.AbortWithStatusJSON(401, gin.H{"message": "未授权"})
            return
        }
        c.Set("userId", claims.UserID)
        c.Next()
    }
}''')

    add_heading(doc, '4.2  好友关系处理', level=2)
    add_para(doc, '好友关系由两张表协同支撑——'
                  'friend_requests 记录申请，friends 记录已建立的关系。'
                  '业务流程不复杂，但有几个细节需要拿捏。')
    add_para(doc, '用户 A 向 B 发申请时，'
                  '后端要先做三道前置校验：B 是否存在、是否已经是好友、是否还有未处理的旧申请。'
                  '任何一道不通过，请求都被驳回；'
                  '全部通过才在 friend_requests 里插入一条 status=待处理 的记录。'
                  '这一步把"重复申请"挡在数据库之外，避免后续要靠唯一索引去兜底。')
    add_para(doc, 'B 在申请页处理时——同意或拒绝——会触发不同分支。'
                  '同意路径走数据库事务，'
                  '把 friend_requests 的 status 改为"已同意"的同时，'
                  '向 friends 表插入 (A, B) 和 (B, A) 两条记录。'
                  '为什么是两条？因为这样 A 查好友能看到 B、B 查好友能看到 A，'
                  '查询路径上完全对称，业务代码不需要做"双向 OR"判断，逻辑更清爽。'
                  '拒绝路径则只更新状态，friends 表保持不动。')
    add_para(doc, '聊天与历史消息接口在执行任何业务之前，'
                  '都会先调用 FriendService.IsFriend(senderId, receiverId)，'
                  '把"非好友越权发消息"这条攻击面直接堵死。'
                  '把校验放在 Service 层而不是 Handler 层，'
                  '是为了让 HTTP 和 WebSocket 两条入口都强制走同一道闸门。')

    # ========== 4.3 WebSocket 重点 ==========
    add_heading(doc, '4.3  基于 WebSocket 的实时聊天实现', level=2)
    add_para(doc, '实时聊天是本系统的两大亮点之一。'
                  '下文按"连接建立与鉴权 → 连接管理与广播 → 异常处理 → 完整时序"四个层次拆开讲。')

    add_heading(doc, '4.3.1  WebSocket 连接建立与鉴权', level=3)
    add_para(doc, 'WebSocket 在 RFC 6455 里复用 HTTP 完成握手——'
                  '客户端发一个带 "Upgrade: websocket" 头的 GET，'
                  '服务端校验后返回 101 完成协议切换。'
                  '握手本质上还是一次 HTTP 请求，'
                  '这就给鉴权留下了天然的入口：'
                  '本系统选择把 JWT 通过 URL 查询参数 token=<jwt> 传给服务端，'
                  '在握手阶段就完成验证。')
    add_para(doc, '另一种常见做法是"先放行连接、再让客户端首帧发 token"。'
                  '但这样做有两个问题：'
                  '未鉴权连接会先占用服务端资源；'
                  '逻辑上要在 WS 层再绕一层"伪鉴权握手"，反而更复杂。'
                  '把校验前移到握手阶段更直接，'
                  '代价是 token 会出现在 URL 里——'
                  '反向代理与日志可能会记录到，'
                  '因此 token 过期时间要短，敏感字段也别往 Payload 里塞。')
    add_para(doc, 'WebSocket 升级处理的关键代码如下：')
    add_code_block(doc, '''func (h *WSHandler) Serve(c *gin.Context) {
    token := c.Query("token")
    claims, err := ParseJWT(token)
    if err != nil {
        c.AbortWithStatusJSON(401, gin.H{"message": "未授权"})
        return
    }
    conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
    if err != nil {
        return
    }
    h.hub.Register(claims.UserID, conn)
    defer h.hub.Unregister(claims.UserID, conn)
    h.readLoop(claims.UserID, conn) // 阻塞读循环，连接断开后退出
}''')

    add_para(doc, '一个细节是，WebSocket 握手同样会经过 Gin 的中间件链路，'
                  '所以全局 CORS 中间件可以直接复用——跨域响应头不必额外处理。'
                  '但鉴权这一层就不能直接套用：'
                  'HTTP 接口的鉴权中间件从 Header 解析 JWT，'
                  'WebSocket 的 JWT 却躺在 URL 查询参数里——'
                  '两边来源不同，混用必然出问题。'
                  '本系统的做法是在路由注册时把 /ws 单独划成一个路由组，'
                  '不挂载 Header 鉴权中间件，'
                  '改在 Handler 内部完成 token 校验。'
                  '这样两条路径互不干扰，扩展时也更容易。')

    add_heading(doc, '4.3.2  连接管理与消息广播', level=3)
    add_para(doc, 'WebSocket Hub 是整个实时聊天模块的"调度中心"。'
                  '它的内部结构其实很朴素：一张并发安全的 map，'
                  'key 是用户 ID，value 是该用户当前持有的所有连接的列表。'
                  '新连接进来追加到对应用户的列表里，'
                  '连接断开就把它从列表中摘掉，'
                  '需要推送时取出活跃连接逐一写入。'
                  '允许一个用户同时挂多个连接，'
                  '这样多标签页、多设备同时在线就自然支持了，不必单独处理。')
    add_para(doc, '写入路径上有一个容易踩坑的地方——锁的粒度。'
                  '本系统采用"短临界区 + 锁外写"的策略：'
                  '先在读写锁保护下拷贝出目标用户的连接列表，'
                  '然后退出临界区，在锁外逐一 conn.WriteMessage。'
                  '原因很简单：网络 I/O 慢，'
                  '一旦把它放进临界区，'
                  'Hub 整张 map 会被一个慢客户端拖住，吞吐量会断崖式下降。'
                  '写入失败的连接会被主动摘除并关闭底层 TCP，'
                  '从根本上避免"僵尸连接"在 map 里越积越多。')
    add_para(doc, '为什么要允许一个用户挂多个连接？'
                  '这是 IM 系统里一个常见但容易被忽略的细节。'
                  '现实中用户可能在公司电脑、家里的笔记本和手机网页上同时登录同一账号——'
                  '如果 Hub 强行只保留最后一条连接，'
                  '前面那些客户端就会"假装在线但收不到消息"。'
                  '本系统选择追加而非覆盖，'
                  '把所有活跃连接都纳入推送范围；'
                  '代价是单条消息要向同一用户写多次，'
                  '但相比"无声的消息丢失"，这一点开销可以接受。')

    add_para(doc, 'Hub 找不到目标用户的活跃连接时怎么办？'
                  '答案是当作离线即可——'
                  '因为消息已经在 messages 表里持久化了。'
                  '离线用户下次登录，前端会调用 GET /api/messages?friendId=xxx 把历史拉回来，'
                  '最终一致性由数据库兜底。'
                  '这套设计的代价是放弃了"实时必达"，'
                  '收益是即便 WS 推送失败、网络抖动、接收方临时下线，'
                  '消息也不会丢。')

    add_heading(doc, '4.3.3  异常处理与重连机制', level=3)
    add_para(doc, '现实中网络远没有局域网那么稳定——'
                  '网络切换、Wi-Fi 抖动、服务端重启都可能导致 WS 连接突然失效。'
                  '前端 utils/websocket.ts 封装了统一的连接创建逻辑，'
                  '在异常处理上分了三个层次。')
    add_para(doc, '第一层是状态可见。'
                  'ChatView.vue 维护一个 socketConnected 响应式变量，'
                  'onopen / onclose 各自更新它，'
                  '界面顶部据此显示"在线同步中"或"等待连接"——'
                  '让用户能直接看到当前的连接状态，而不是凭感觉猜。'
                  '第二层是错误反馈，'
                  'onerror 触发时把错误写入 errorMessage，'
                  '醒目地提示"WebSocket 连接失败"，'
                  '避免用户在断网状态下还在反复点发送。'
                  '第三层是资源回收，'
                  '在 onBeforeUnmount 钩子里主动 socket.close()，'
                  '防止 Hub 残留无效连接占用内存。')
    add_para(doc, '至于自动重连，本系统刻意没做复杂的指数退避机制。'
                  '原因前面提过——消息已经持久化在数据库里，'
                  'WS 只负责实时推送这一件事。'
                  '即便连接短暂断开，'
                  '用户刷新页面或重新进入聊天时，'
                  '一次 GET /api/messages 就能把历史拉回来。'
                  '把可靠性交给数据库、把实时性交给 WS，'
                  '是一种简单又有效的解耦。')

    add_heading(doc, '4.3.4  消息收发完整时序', level=3)
    add_para(doc, '把前面几节的设计拼起来，'
                  '一条消息从 A 的浏览器走到 B 的浏览器，'
                  '完整时序如图 4-1 所示。'
                  '图里有一个细节值得特别留意——'
                  '服务端这条泳道从头到尾只接触密文，'
                  '明文只存在于两端浏览器中。')
    add_image(doc, os.path.join(DIA, '03_WebSocket聊天时序图.png'),
              '图 4-1  基于 WebSocket 的实时聊天时序图', width_inch=6.2)
    add_para(doc, 'A 在聊天页输入内容、点击发送的瞬间，'
                  '前端先调用 e2ee.ts 里的加密函数，'
                  '用 A 的公钥和 B 的公钥分别对同一段明文加密一次，'
                  '得到 sender_ciphertext 和 receiver_ciphertext 两份密文。'
                  '随后通过 POST /api/messages 把两份密文连同接收方 ID 一并提交。')
    add_para(doc, '后端 MessageService 在落库前做三道校验——'
                  '双方是否互为好友、密文是否合法、算法标识是否在支持的白名单里（rsa-oaep-sha256）。'
                  '只要任意一道不通过都会被拒绝；'
                  '全部通过才会把两份密文写入 messages 表。'
                  '落库成功后，'
                  '一方面通过 WebSocket Hub 给 B 推送一条类型为 chat_message 的 JSON 帧，'
                  '帧里载着消息 ID、发送方 ID、密文与算法标识；'
                  '另一方面给 A 返回 HTTP 201，'
                  '前端把这条消息追加到本地列表。')
    add_para(doc, 'B 端收到 WS 帧之后，'
                  'selectMessagePayloadForUser 会自动挑出 receiver_ciphertext，'
                  '用本地私钥解开并渲染。'
                  '如果解密失败——'
                  '比如对方在另一台设备生成了新密钥、本地拿到的还是老公钥——'
                  '界面会退化成"***（已加密）"的占位文本，'
                  '不会因为单条消息异常导致整个会话崩掉。')
    add_para(doc, '这个流程里 HTTP 和 WebSocket 各司其职：'
                  '前者保证消息可靠落库——'
                  '即"持久化通道"；'
                  '后者负责实时推送——'
                  '即"实时通道"。'
                  '消息的不丢失最终由数据库兜底，'
                  '消息的实时性由 WebSocket 推送保证，两者并不重叠。')

    # ========== 4.4 加密 重点 ==========
    add_heading(doc, '4.4  端到端消息加密机制', level=2)
    add_para(doc, '端到端加密是本系统的另一大亮点。'
                  '传统 IM 通常依赖 TLS 完成传输层加密——'
                  '消息在线路上安全，'
                  '但一旦到达服务端往往就以明文形式躺在数据库里。'
                  '只要服务端被攻破，或者运维/内部人员越权访问数据库，'
                  '损失是灾难性的。'
                  '本系统的解法是把加密提前到浏览器侧：'
                  '消息在发送前已经变成密文，'
                  '服务端只参与存储和转发——'
                  '从根本上让"服务端泄漏明文"这条攻击路径不再成立。')

    add_heading(doc, '4.4.1  设计目标与总体思路', level=3)
    add_para(doc, '具体到设计目标，本系统要同时满足四件事。'
                  '一是服务端不能接触消息明文——这是这套机制最核心的承诺，'
                  '否则端到端加密只剩"名义上"的安全；'
                  '二是发送方和接收方都能在本地回看历史消息，'
                  '而不是只有接收方可读——'
                  '这条与安全性同等重要，否则用户体验会断崖式下跌；'
                  '三是整个过程对用户透明，'
                  '密钥生成、上传、加解密这些步骤都在背后悄悄进行，用户感知不到；'
                  '四是算法选型必须是公开、经过同行评议、广泛使用的非对称加密算法，'
                  '避免自己造轮子。')
    add_para(doc, '为实现这四点，本系统采用"浏览器侧生成密钥对 + 前端双重加密 + '
                  '服务端密文存储与转发"的整体思路，'
                  '总体流程见图 4-2。')
    add_image(doc, os.path.join(DIA, '04_端到端加密流程图.png'),
              '图 4-2  端到端消息加密总流程图', width_inch=6.2)

    add_para(doc, '这套思路并不新颖，'
                  '本质上沿用了 PGP 时代的"接收方公钥加密、接收方私钥解密"模型——'
                  '本文的工程价值不在算法层面，'
                  '而是把它适配进现代浏览器、'
                  '并解决了"双向回看历史消息"这个体验问题。'
                  '用最少的密码学机制换最大的工程可用性，'
                  '是本系统在设计取舍上的核心立场。')
    add_para(doc, '需要明确划清两条边界。'
                  '其一，本方案处理的是消息级加密，'
                  '不涉及传输层——'
                  '前者保证服务端看不到明文，'
                  '后者保证中间节点看不到明文，二者并不冲突，'
                  '工程上应当同时启用 TLS。'
                  '其二，本方案不涉及密钥协商，'
                  '直接使用每个用户的长期 RSA 密钥对加解密——'
                  '简洁的代价是无法做到前向保密。'
                  '这两条边界划清楚，'
                  '可以避免后面读者把它误读为"完整的工业级 E2EE"。')

    add_heading(doc, '4.4.2  RSA-OAEP 密钥对生成与管理', level=3)
    add_para(doc, '密钥生成完全在浏览器侧完成，'
                  '调用的是 Web Crypto API 的 window.crypto.subtle.generateKey。'
                  '算法选 RSA-OAEP，模长 2048 位，公开指数 65537，哈希函数 SHA-256——'
                  '这是当前业界对 RSA 的标准推荐配置，'
                  'TLS 证书、JWT 签名用的都是这套参数。'
                  '生成的密钥对中，'
                  '私钥按 PKCS#8 格式导出，'
                  '公钥按 SPKI 格式导出，'
                  '再统一做一次 Base64 编码，方便序列化成字符串保存与传输。')
    add_para(doc, '前端密钥生成与导出的关键代码如下：')
    add_code_block(doc, '''export async function generateKeyPair() {
  return window.crypto.subtle.generateKey(
    {
      name: 'RSA-OAEP',
      modulusLength: 2048,
      publicExponent: new Uint8Array([1, 0, 1]),
      hash: 'SHA-256'
    },
    true,
    ['encrypt', 'decrypt']
  )
}''')
    add_para(doc, '私钥只保存在用户当前设备的 localStorage 里，'
                  '键名为 e2ee-private-key:<userId>。'
                  '这个值不通过任何接口上传服务端——'
                  '"私钥不出本机"这条规则是端到端加密的底线，'
                  '系统在工程上必须自己守住。'
                  '用户首次进入聊天功能时，'
                  '前端先检查本地是否已有私钥：'
                  '没有就生成新密钥对并上传公钥；'
                  '已有但服务端却没有对应公钥（典型情况是老用户清过缓存后又回来），'
                  '则从私钥反推出公钥再补传一次。')

    add_para(doc, '为什么选 localStorage 而不是 sessionStorage 或 IndexedDB？'
                  'sessionStorage 在刷新或关闭浏览器后就清空了，'
                  '加密会话会随之中断，体验太差；'
                  'IndexedDB 接口繁琐，'
                  '而本系统要存的只是一条 1.6 KB 左右的 RSA 私钥——'
                  '杀鸡用牛刀。'
                  'localStorage 在简单与可用之间提供了一个恰到好处的平衡。'
                  '另一个细节是，存储键里专门带上了用户 ID，'
                  '同一台浏览器即便切换账号，私钥也不会互相覆盖。')

    add_heading(doc, '4.4.3  公钥上传与好友公钥获取', level=3)
    add_para(doc, '密钥对生成后，前端通过 PUT /api/users/me/public-key 把公钥和算法标识 '
                  '(rsa-oaep-sha256) 一并上传，'
                  '后端写入 users 表的 public_key 与 public_key_algorithm 字段。'
                  'A 要给 B 发消息时，B 的公钥并不需要单独请求——'
                  '它早就随 GET /api/friends 的好友列表一起拉回来了。'
                  '这一步省掉了"每发一条消息额外查一次公钥"的开销，'
                  '前端交互次数明显减少。')
    add_para(doc, '后端在更新公钥接口里只做算法白名单校验，'
                  '不对公钥本身做解析或验证。'
                  '原因是公钥本身没有保密属性，'
                  '真正棘手的是"如何确认这把公钥真的来自对方"——'
                  '工业级方案靠指纹比对、可信目录或带外通道解决。'
                  '本系统目前假定双方信任服务端返回的公钥，'
                  '这是相对工业级 E2EE 的一项简化，留待后续工作处理。')

    add_heading(doc, '4.4.4  双重密文加密与服务端密文转发', level=3)
    add_para(doc, '用户点击发送的瞬间，'
                  '前端用接收方公钥和发送方公钥分别对同一段明文做一次 RSA-OAEP 加密，'
                  '得到两份独立的密文，'
                  '再通过 POST /api/messages 一起提交。'
                  '前端加密的关键代码如下：')
    add_code_block(doc, '''const receiverPublicKey = await importPublicKey(friend.publicKey)
const senderPublicKey = await importPublicKey(authStore.user.publicKey)
const content = draft.value.trim()
const receiverEncrypted = await encryptMessage(receiverPublicKey, content)
const senderEncrypted = await encryptMessage(senderPublicKey, content)
const { data } = await http.post('/messages', {
  receiverId: currentFriendId.value,
  senderCiphertext: senderEncrypted.ciphertext,
  senderAlgorithm: senderEncrypted.algorithm,
  receiverCiphertext: receiverEncrypted.ciphertext,
  receiverAlgorithm: receiverEncrypted.algorithm
})''')
    add_para(doc, '后端 MessageService 收到请求后，'
                  '先调用 FriendService.IsFriend 验证好友关系，'
                  '再检查两份密文非空、算法标识在 rsa-oaep-sha256 白名单内，'
                  '校验全部通过才写入 messages 表，'
                  '随后由 WebSocket Hub 推送给接收方。'
                  '从代码层面看，'
                  '后端没有调用任何解密接口，'
                  '也没有任何明文消息出现在内存或日志里——'
                  '这一点对端到端加密承诺至关重要。')
    add_para(doc, '为收发双方各保存一份密文，'
                  '是本系统工程上最关键的一次取舍。'
                  '只保存接收方密文，发送方在自己设备上就回看不了历史聊天，'
                  '体验直接崩；'
                  '让服务端保存一份明文以便双方查看，'
                  '又把端到端加密的核心承诺破坏掉了。'
                  '双密文方案是这两难局面下的折中——'
                  '安全性不打折，双向可读也保住了。')

    add_para(doc, '另一处看似不起眼的细节是 Service 层的"防御性校验"。'
                  '为了避免后端在某种边界情况下不小心把明文写进数据库，'
                  '系统对密文字段做了严格的非空校验和算法白名单校验：'
                  '任何一份密文为空或算法标识非法，'
                  '直接返回 400 并记录到错误日志。'
                  '这条防线在程序员失误或恶意构造请求的情况下都能挡住异常数据，'
                  '保证 messages 表里只可能出现合法密文。')

    add_heading(doc, '4.4.5  接收端解密与历史消息展示', level=3)
    add_para(doc, '接收端会在两种场景下做解密：'
                  '一是 WebSocket 实时收到新消息时，'
                  '二是通过 HTTP 拉取历史消息时。'
                  '两条路径的逻辑由 selectMessagePayloadForUser 与 toRenderMessage 统一封装——'
                  '前者根据当前用户 ID 决定要解的是 sender_ciphertext 还是 receiver_ciphertext，'
                  '后者拿本地私钥调用 window.crypto.subtle.decrypt 真正完成解密。'
                  '解密成功就展示明文；'
                  '失败的话（典型情况是用户换了设备、本地私钥与服务端公钥对不上）'
                  '则退化成"***（已加密）"占位文本，'
                  '同时在错误提示区给出"当前设备没有可用私钥"的友好说明，'
                  '至少让用户知道发生了什么。')
    add_para(doc, '需要承认的是，'
                  '由于私钥只保存在最初生成它的那台设备上，'
                  '本系统暂时不支持多设备解密同一份历史。'
                  '这是 RSA 简化方案的固有局限，'
                  '也是后续工作的一个明确方向。')

    add_heading(doc, '4.4.6  安全性分析', level=3)
    add_para(doc, '从安全模型角度看，'
                  '本系统已经具备端到端加密的基本特征——'
                  '消息明文只存在于两端浏览器，'
                  '服务端的职责被压缩到公钥保存、密文存储与消息转发。'
                  '即使数据库被整体脱库，'
                  '攻击者拿到的也只是一堆密文和公钥，'
                  '在没有私钥的前提下推不出任何明文。')
    add_para(doc, '不过要把话说清楚：本系统目前的实现更接近'
                  '"前端本地加密 + 服务端密文转发与存储"，'
                  '距离工业级 E2EE 体系还有距离。具体差在几处。')
    add_para(doc, '公钥真实性校验是第一个缺口。'
                  '用户拿到的对方公钥来自服务端——'
                  '一旦服务端被攻破，'
                  '它完全可以把对方公钥替换成攻击者的，'
                  '从而发起中间人攻击。'
                  'Signal 等成熟方案靠指纹比对、可信目录或带外验证来堵这个洞，'
                  '本系统目前没做。')
    add_para(doc, '前向保密是第二个缺口。'
                  'RSA 私钥是长期凭证，'
                  '一旦泄露，攻击者可以追溯解密历史上的所有密文。'
                  'Signal 通过 Double Ratchet 在每条消息上独立派生临时会话密钥，'
                  '让单次密钥泄露的影响范围被限制在极短时间窗口内——'
                  '这正是前向保密的核心收益。')
    add_para(doc, '第三个缺口在会话密钥协商。'
                  '业界通行的做法是用 RSA/ECC 协商出短期会话密钥，'
                  '再用 AES-GCM 这种对称算法加密实际正文，'
                  '兼顾安全与性能。'
                  '本系统直接用 RSA-OAEP 加密单条短消息，'
                  '受 2048 位 RSA 自身的限制——'
                  '单次最多约 190 字节明文，'
                  '在 Web 端聊天文本场景下勉强够用，'
                  '一旦要支持长文本或文件，'
                  '就必须切到"RSA 加密会话密钥 + AES-GCM 加密正文"的混合方案。')
    add_para(doc, '最后一处是多设备同步。'
                  '一个用户的私钥目前只在一个浏览器里有效，'
                  '换设备或清缓存后历史消息就无法解密——'
                  '这对真实用户体验是个明显短板，但放在毕业设计的尺度上可以接受。')
    add_para(doc, '把这些差距摊在台面上说，'
                  '不是要为现有实现辩护，'
                  '而是想表明：本系统已经把端到端加密的核心思想落到了可运行的代码里，'
                  '同时也清楚地知道下一步该往哪走——'
                  '后续工作只需在这套架构上替换密钥协商与对称加密层，'
                  '就能逼近工业级 E2EE 的能力边界。')
    add_page_break(doc)
