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
    add_para(doc, '实时聊天能力是本系统工程实现的核心组成部分。'
                  '与传统的轮询或长轮询方案相比，IM 场景对“低延迟、持续在线、服务端可主动下行”'
                  '有更强要求，而 WebSocket 恰好提供了一个更贴近该场景本质的通信模型。'
                  '其关键不只是“连接长”，而在于服务端能够在无需客户端再次发起请求的前提下，'
                  '主动将新消息推送到目标连接，这一点是 HTTP 请求—响应模型难以高效模拟的。'
                  '下文按“连接建立与鉴权 → 连接管理与广播 → 异常处理 → 完整时序”四个层次展开。')

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

    add_para(doc, '一个容易被忽略的细节是，WebSocket 握手虽然最终会升级为长连接，'
                  '但其入口仍然是一条标准 HTTP 请求，因此同样会穿过 Gin 的中间件链路。'
                  '这使得跨域响应头、统一日志等横切逻辑可以直接复用。'
                  '不过鉴权策略不能照搬普通接口：HTTP 接口的 JWT 通常位于 Authorization Header 中，'
                  '而 WebSocket 握手更适合从查询参数中读取 token。若强行共用同一套中间件，'
                  '势必会造成来源冲突与责任边界混乱。'
                  '本系统因此将 /ws 单独划为一组路由，在 Handler 内部完成 JWT 校验。'
                  '这种处理方式的实质，是将“通信协议升级”和“业务身份确认”绑定在同一入口完成，'
                  '从而避免未鉴权连接进入连接管理层。')

    add_heading(doc, '4.3.2  连接管理与消息广播', level=3)
    add_para(doc, 'WebSocket Hub 是整个实时聊天模块的"调度中心"。'
                  '它的内部结构其实很朴素：一张并发安全的 map，'
                  'key 是用户 ID，value 是该用户当前持有的所有连接的列表。'
                  '新连接进来追加到对应用户的列表里，'
                  '连接断开就把它从列表中摘掉，'
                  '需要推送时取出活跃连接逐一写入。'
                  '允许一个用户同时挂多个连接，'
                  '这样多标签页、多设备同时在线就自然支持了，不必单独处理。')
    add_para(doc, '连接管理真正体现工程差异的地方，在于并发控制与广播策略。'
                  '本系统采用“短临界区 + 锁外写”的方式管理广播：先在读写锁保护下复制目标用户的连接列表，'
                  '再在锁外逐一执行 conn.WriteMessage。之所以这样设计，'
                  '是因为网络 I/O 的耗时远高于内存读写，若在持锁状态下直接写 Socket，'
                  '单个慢客户端就可能拖住整个 Hub 的广播路径，进而放大锁竞争。'
                  '换言之，该设计本质上是在“一致性开销”与“吞吐能力”之间做权衡：'
                  '允许连接列表在极短时间内与真实状态存在微小差异，换取广播路径上的整体性能稳定性。'
                  '对于写入失败的连接，系统会主动摘除并关闭底层 TCP 连接，以防无效连接长期滞留。')
    add_para(doc, '为什么要允许一个用户挂多个连接？'
                  '这是 IM 系统里一个常见但容易被忽略的细节。'
                  '现实中用户可能在公司电脑、家里的笔记本和手机网页上同时登录同一账号——'
                  '如果 Hub 强行只保留最后一条连接，'
                  '前面那些客户端就会"假装在线但收不到消息"。'
                  '本系统选择追加而非覆盖，'
                  '把所有活跃连接都纳入推送范围；'
                  '代价是单条消息要向同一用户写多次，'
                  '但相比"无声的消息丢失"，这一点开销可以接受。')

    add_para(doc, '当 Hub 找不到目标用户的活跃连接时，系统将该用户视为离线。'
                  '这里的关键不在于“是否即时送达”，而在于“消息是否最终可达”。'
                  '本系统有意将“消息存储”与“消息实时投递”拆成两个层次：'
                  'HTTP 写库负责可靠持久化，WebSocket 负责在线状态下的低延迟下发。'
                  '因此，一旦接收方离线、推送失败或网络中断，消息仍会完整保留在 messages 表中；'
                  '待用户重新进入会话页面后，前端通过 GET /api/messages?friendId=xxx 拉取历史密文，'
                  '再在本地解密展示。'
                  '这种设计事实上放弃了“绝对实时必达”的强承诺，换取了更简单、可验证且不易丢消息的实现路径。'
                  '对毕业设计场景而言，这是一个合理且必要的取舍。')

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
    add_para(doc, '端到端消息加密是本文最核心的技术亮点。'
                  '传统 IM 系统通常依赖 TLS 完成链路保护，这意味着消息在客户端与服务端之间传输时处于加密状态，'
                  '但一旦到达服务端便往往以明文形式进入数据库或业务内存。'
                  '换言之，TLS 解决的是“中间节点能否窃听”问题，'
                  '而不能解决“服务端是否天然可信”问题。'
                  '本系统的设计思路正是将安全边界前移到浏览器本地：'
                  '明文在发送前即完成加密，服务端只接触密文，'
                  '从而将消息内容从服务端明文处理链路中剥离出去。'
                  '这种做法并不能消除所有风险，但它实质上改变了攻击面，'
                  '把最常见的数据库泄露风险从“直接看到消息内容”降为“只能拿到不可读密文”。')

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
    add_para(doc, '为实现上述目标，本系统采用“浏览器侧生成密钥对 + 前端双重加密 + '
                  '服务端密文存储与转发”的整体思路，流程见图 4-2。'
                  '从工程实现角度看，这一方案并未引入复杂的会话密钥协商与双棘轮机制，'
                  '而是优先保证“原理清晰、链路闭环、实现可验证”。'
                  '其价值不在于直接追平工业级加密协议，而在于用较低实现复杂度把 E2EE 的核心思想'
                  '真实地嵌入到一个可运行的 Web IM 系统中。')
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
    add_para(doc, '密钥生成完全在浏览器侧完成，调用的是 Web Crypto API 的 '
                  'window.crypto.subtle.generateKey。算法选用 RSA-OAEP，模长 2048 位，公开指数为 65537，'
                  '哈希函数为 SHA-256。选择 RSA-OAEP 而不是更简单的 RSA-PKCS#1 v1.5，'
                  '原因在于 OAEP 通过随机填充与掩码生成函数显著提升了对选择密文攻击的抵抗能力，'
                  '属于现代密码学实践中对 RSA 加密的推荐用法。'
                  '另一方面，Web Crypto API 由浏览器原生实现提供，'
                  '相较于纯 JavaScript 加密库，在性能、稳定性和接口一致性方面更适合承担生产级别的加解密原语调用。'
                  '生成后的私钥按 PKCS#8 格式导出，公钥按 SPKI 格式导出，'
                  '再统一转换为 Base64 字符串，便于持久化保存与通过 HTTP 接口传输。')
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
    add_para(doc, '用户点击发送时，前端使用接收方公钥与发送方公钥分别对同一段明文执行一次 '
                  'RSA-OAEP 加密，得到两份彼此独立的密文，再通过 POST /api/messages 一并提交。'
                  '双重密文设计是本系统最关键的工程取舍之一：'
                  '若只保留接收方可解密的密文，则发送方将无法在本地回看自己发出的历史消息；'
                  '若服务端保留明文副本，又会直接破坏端到端加密的安全边界。'
                  '因此，分别面向收发双方生成两份可由各自私钥解密的密文，'
                  '是在“历史可读性”与“服务端不见明文”之间取得平衡的必要方案。'
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
    add_para(doc, '从安全模型角度分析，本系统已经具备端到端加密的基本特征：'
                  '消息明文仅存在于通信双方的浏览器环境中，服务端职责被压缩为公钥保存、密文存储与消息转发。'
                  '这意味着即便数据库被整体脱库，攻击者能够获取的也只是密文与公钥，'
                  '而无法在缺失私钥的前提下直接恢复聊天内容。'
                  '与传统明文存储方案相比，这一机制实质上将服务端数据泄露的后果从“明文暴露”'
                  '降低为“密文暴露”，安全后果发生了本质变化。')
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
