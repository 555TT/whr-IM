"""
Generate diagrams for thesis. Output 300dpi PNG.
"""
import matplotlib.pyplot as plt
import matplotlib.patches as mpatches
from matplotlib.patches import FancyBboxPatch, FancyArrowPatch, Rectangle
from matplotlib.lines import Line2D
import os

OUT = os.path.dirname(os.path.abspath(__file__))

plt.rcParams['font.sans-serif'] = ['Arial Unicode MS', 'PingFang SC', 'Heiti TC', 'STHeiti']
plt.rcParams['axes.unicode_minus'] = False

DPI = 220


def box(ax, x, y, w, h, text, fc='#E8F1FB', ec='#0071E3', lw=1.2, fs=10, bold=False):
    p = FancyBboxPatch((x, y), w, h,
                       boxstyle="round,pad=0.02,rounding_size=0.06",
                       fc=fc, ec=ec, lw=lw)
    ax.add_patch(p)
    weight = 'bold' if bold else 'normal'
    ax.text(x + w / 2, y + h / 2, text, ha='center', va='center',
            fontsize=fs, fontweight=weight, wrap=True)


def arrow(ax, x1, y1, x2, y2, text='', color='#444', style='->', lw=1.2, fs=9, offset=(0, 0.08), curve=0.0):
    a = FancyArrowPatch((x1, y1), (x2, y2),
                        arrowstyle=style, mutation_scale=14,
                        color=color, lw=lw,
                        connectionstyle=f"arc3,rad={curve}")
    ax.add_patch(a)
    if text:
        ax.text((x1 + x2) / 2 + offset[0], (y1 + y2) / 2 + offset[1], text,
                ha='center', va='center', fontsize=fs, color=color)


# ----------------- 1. 系统总体架构图 -----------------
def fig_arch():
    fig, ax = plt.subplots(figsize=(11, 7), dpi=DPI)
    ax.set_xlim(0, 11); ax.set_ylim(0, 7); ax.axis('off')

    ax.text(5.5, 6.6, '即时通讯系统总体架构图', ha='center', fontsize=14, fontweight='bold')

    # 浏览器层
    box(ax, 0.4, 4.6, 10.2, 1.6, '', fc='#F6F8FB', ec='#888', lw=1)
    ax.text(0.55, 6.0, '客户端层（浏览器）', fontsize=11, fontweight='bold', color='#444')
    box(ax, 0.7, 4.85, 2.0, 0.85, 'Vue 3\n页面视图\nViews', fc='#E8F1FB', ec='#0071E3')
    box(ax, 2.9, 4.85, 2.0, 0.85, 'Pinia\n状态管理\nStores', fc='#E8F1FB', ec='#0071E3')
    box(ax, 5.1, 4.85, 2.0, 0.85, 'Axios\nHTTP 封装', fc='#E8F1FB', ec='#0071E3')
    box(ax, 7.3, 4.85, 2.0, 0.85, 'WebSocket\n客户端', fc='#FFF1E0', ec='#F08C00')
    box(ax, 9.5, 4.85, 1.1, 0.85, 'Web\nCrypto\nE2EE', fc='#FFE3E3', ec='#D43A3A', fs=9)

    # 网络层
    ax.text(5.5, 4.4, '↓ HTTP / WebSocket（JWT 鉴权） ↓', ha='center', fontsize=10, color='#666')

    # 服务端层
    box(ax, 0.4, 1.7, 10.2, 2.5, '', fc='#F6F8FB', ec='#888', lw=1)
    ax.text(0.55, 4.0, '服务端层（Go / Gin）', fontsize=11, fontweight='bold', color='#444')

    box(ax, 0.7, 3.05, 2.0, 0.8, 'Router\n路由注册', fc='#E8F4EA', ec='#2E8B57')
    box(ax, 2.9, 3.05, 2.0, 0.8, 'Middleware\nJWT / CORS', fc='#E8F4EA', ec='#2E8B57')
    box(ax, 5.1, 3.05, 2.0, 0.8, 'Handler\nHTTP / WS', fc='#E8F4EA', ec='#2E8B57')
    box(ax, 7.3, 3.05, 2.0, 0.8, 'Service\n业务逻辑', fc='#E8F4EA', ec='#2E8B57')
    box(ax, 9.5, 3.05, 1.1, 0.8, 'WS\nHub', fc='#FFF1E0', ec='#F08C00')

    box(ax, 0.7, 1.95, 4.2, 0.8, 'Repository（GORM 数据访问层）', fc='#E8F4EA', ec='#2E8B57')
    box(ax, 5.1, 1.95, 4.2, 0.8, 'Model（User / Friend / Message）', fc='#E8F4EA', ec='#2E8B57')

    ax.text(5.5, 1.55, '↓ SQL ↓', ha='center', fontsize=10, color='#666')

    # 数据层
    box(ax, 0.4, 0.4, 10.2, 0.95, 'MySQL 8.x — users / friend_requests / friends / messages', fc='#FFF8E0', ec='#C8A100', fs=11, bold=True)

    plt.tight_layout()
    fig.savefig(os.path.join(OUT, '01_系统总体架构图.png'), bbox_inches='tight')
    plt.close(fig)


# ----------------- 2. 用例图 -----------------
def fig_usecase():
    fig, ax = plt.subplots(figsize=(10, 7), dpi=DPI)
    ax.set_xlim(0, 10); ax.set_ylim(0, 7); ax.axis('off')
    ax.text(5, 6.6, '系统用户用例图', ha='center', fontsize=14, fontweight='bold')

    # actor left (用户A)
    def actor(cx, cy, name):
        ax.plot(cx, cy + 0.35, marker='o', ms=14, mfc='white', mec='#222')
        ax.plot([cx, cx], [cy + 0.25, cy - 0.25], color='#222')
        ax.plot([cx - 0.25, cx + 0.25], [cy + 0.1, cy + 0.1], color='#222')
        ax.plot([cx, cx - 0.22], [cy - 0.25, cy - 0.6], color='#222')
        ax.plot([cx, cx + 0.22], [cy - 0.25, cy - 0.6], color='#222')
        ax.text(cx, cy - 0.85, name, ha='center', fontsize=10, fontweight='bold')

    actor(0.9, 3.5, '用户')
    actor(9.1, 3.5, '系统管理员\n（潜在）')

    # 系统边界
    sys = Rectangle((2.2, 0.6), 5.6, 5.4, fc='#F6F8FB', ec='#444', lw=1.2)
    ax.add_patch(sys)
    ax.text(5.0, 5.8, '即时通讯系统', ha='center', fontsize=11, fontweight='bold', color='#444')

    cases = [
        (3.3, 5.2, '注册账号'),
        (6.6, 5.2, '登录系统'),
        (3.3, 4.5, '查看 / 修改资料'),
        (6.6, 4.5, '上传个人公钥'),
        (3.3, 3.8, '发送好友申请'),
        (6.6, 3.8, '处理好友申请'),
        (3.3, 3.1, '查看好友列表'),
        (6.6, 3.1, '建立 WebSocket 长连接'),
        (3.3, 2.4, '发送加密消息'),
        (6.6, 2.4, '接收并解密消息'),
        (3.3, 1.7, '查看历史消息'),
        (6.6, 1.7, '查看密文展示'),
        (5.0, 1.0, '退出登录'),
    ]
    for x, y, t in cases:
        e = mpatches.Ellipse((x, y), 1.7, 0.45, fc='#E8F1FB', ec='#0071E3', lw=1)
        ax.add_patch(e)
        ax.text(x, y, t, ha='center', va='center', fontsize=9)

    # connect user to several
    for _, y, _ in cases:
        ax.plot([1.15, 2.5], [3.5, y], color='#888', lw=0.6)

    plt.tight_layout()
    fig.savefig(os.path.join(OUT, '02_系统用例图.png'), bbox_inches='tight')
    plt.close(fig)


# ----------------- 3. WebSocket 时序图 -----------------
def fig_ws_sequence():
    fig, ax = plt.subplots(figsize=(11, 8.2), dpi=DPI)
    ax.set_xlim(0, 11); ax.set_ylim(0, 9); ax.axis('off')
    ax.text(5.5, 8.6, '基于 WebSocket 的实时聊天时序图', ha='center', fontsize=14, fontweight='bold')

    # lifelines
    actors = [(1.2, 'User A\n(浏览器)'),
              (3.6, 'Frontend A\n(Vue)'),
              (6.0, 'Server\n(Gin + WS Hub)'),
              (8.4, 'Frontend B\n(Vue)'),
              (10.2, 'User B\n(浏览器)')]
    for x, name in actors:
        box(ax, x - 0.8, 7.7, 1.6, 0.55, name, fc='#F0F4F8', ec='#444', fs=10, bold=True)
        ax.plot([x, x], [0.4, 7.7], ls='--', color='#aaa', lw=0.8)

    def msg(x1, x2, y, text, dashed=False, color='#0071E3'):
        style = '->' if not dashed else '->'
        a = FancyArrowPatch((x1, y), (x2, y), arrowstyle='->', mutation_scale=12,
                            color=color, lw=1.1, ls='--' if dashed else '-')
        ax.add_patch(a)
        ax.text((x1 + x2) / 2, y + 0.13, text, ha='center', fontsize=9, color=color)

    def note(x, y, w, h, text):
        p = Rectangle((x, y), w, h, fc='#FFF8DC', ec='#999', lw=0.8)
        ax.add_patch(p)
        ax.text(x + w / 2, y + h / 2, text, ha='center', va='center', fontsize=8.5)

    y = 7.2
    msg(1.2, 3.6, y, '1. 输入消息 / 点击发送'); y -= 0.45
    note(2.6, y - 0.05, 2.0, 0.35, '前端先用 RSA-OAEP 双重加密'); y -= 0.55

    msg(3.6, 6.0, y, '2. POST /api/messages (sender_ct + receiver_ct)'); y -= 0.45
    note(4.5, y - 0.05, 3.0, 0.35, '后端 JWT 校验 + 好友关系校验 + 校验算法'); y -= 0.55
    msg(6.0, 6.0, y, '3. 落库（messages 表保存双密文）', color='#888'); y -= 0.45
    msg(6.0, 3.6, y, '4. 返回 201 Created + 消息体（密文）', dashed=True); y -= 0.45
    msg(6.0, 8.4, y, '5. WS Hub 推送 chat_message 帧（密文）', color='#F08C00'); y -= 0.45
    msg(8.4, 10.2, y, '6. 前端 B 用本地私钥解密后渲染'); y -= 0.45

    note(0.6, y - 0.4, 9.8, 0.35, '说明：服务端全程只接触密文；明文仅存在于发送方与接收方两端浏览器。'); y -= 0.7

    # 连接建立部分（顶端补充）
    y = 4.6
    ax.text(5.5, y + 0.3, '— WebSocket 长连接建立 —', ha='center', fontsize=10, color='#888')
    msg(3.6, 6.0, y, 'GET /ws?token=JWT'); y -= 0.45
    note(4.4, y - 0.05, 3.2, 0.35, '服务端校验 JWT → 升级协议 → 注册到 Hub'); y -= 0.55
    msg(6.0, 3.6, y, '101 Switching Protocols', dashed=True); y -= 0.45
    msg(6.0, 8.4, y, '101 Switching Protocols（B 端同理）', dashed=True); y -= 0.45

    plt.tight_layout()
    fig.savefig(os.path.join(OUT, '03_WebSocket聊天时序图.png'), bbox_inches='tight')
    plt.close(fig)


# ----------------- 4. 端到端消息加密流程图 -----------------
def fig_e2ee():
    fig, ax = plt.subplots(figsize=(11, 7.5), dpi=DPI)
    ax.set_xlim(0, 11); ax.set_ylim(0, 7.5); ax.axis('off')
    ax.text(5.5, 7.1, '端到端消息加密总流程图（RSA-OAEP-SHA256）', ha='center', fontsize=14, fontweight='bold')

    # Sender side
    box(ax, 0.3, 4.0, 3.5, 2.4, '', fc='#F6F8FB', ec='#888')
    ax.text(0.5, 6.2, '发送方浏览器', fontsize=11, fontweight='bold', color='#0071E3')
    box(ax, 0.5, 5.35, 3.1, 0.65, '生成 RSA-2048 密钥对\n(Web Crypto API)')
    box(ax, 0.5, 4.55, 3.1, 0.65, '私钥保存 localStorage\n公钥上传服务端')

    box(ax, 0.3, 1.4, 3.5, 2.3, '', fc='#F6F8FB', ec='#888')
    ax.text(0.5, 3.55, '发送方加密阶段', fontsize=11, fontweight='bold', color='#0071E3')
    box(ax, 0.5, 2.7, 3.1, 0.65, '获取自身公钥 + 对方公钥')
    box(ax, 0.5, 1.85, 3.1, 0.7, '同一明文加密两份\n→ sender_ct / receiver_ct', fc='#FFE3E3', ec='#D43A3A')

    # Server
    box(ax, 4.2, 2.7, 2.6, 2.0, '', fc='#FFF1E0', ec='#F08C00', lw=1.4)
    ax.text(5.5, 4.5, '服务端', ha='center', fontsize=11, fontweight='bold', color='#F08C00')
    box(ax, 4.35, 3.55, 2.3, 0.55, 'POST /messages\n仅接收密文', fc='#FFFFFF', ec='#F08C00')
    box(ax, 4.35, 2.85, 2.3, 0.55, '入库 + WS Hub 转发', fc='#FFFFFF', ec='#F08C00')

    # Receiver side
    box(ax, 7.2, 4.0, 3.5, 2.4, '', fc='#F6F8FB', ec='#888')
    ax.text(7.4, 6.2, '接收方浏览器', fontsize=11, fontweight='bold', color='#2E8B57')
    box(ax, 7.4, 5.35, 3.1, 0.65, '私钥已就绪\n(初次登录会自动生成)')
    box(ax, 7.4, 4.55, 3.1, 0.65, '收到 chat_message 帧\n(receiver_ct)')

    box(ax, 7.2, 1.4, 3.5, 2.3, '', fc='#F6F8FB', ec='#888')
    ax.text(7.4, 3.55, '接收方解密阶段', fontsize=11, fontweight='bold', color='#2E8B57')
    box(ax, 7.4, 2.7, 3.1, 0.65, '使用本地私钥解密 receiver_ct')
    box(ax, 7.4, 1.85, 3.1, 0.7, '渲染明文 / 失败显示密文占位', fc='#E8F4EA', ec='#2E8B57')

    # Arrows
    arrow(ax, 3.8, 4.95, 4.2, 4.0, '上传公钥', curve=0.1)
    arrow(ax, 3.8, 2.2, 4.35, 3.2, '提交密文', curve=-0.05)
    arrow(ax, 6.8, 3.6, 7.2, 4.9, 'WS 转发\n(receiver_ct)', curve=-0.1)
    arrow(ax, 8.95, 4.5, 8.95, 3.55, '密文')
    arrow(ax, 8.95, 2.65, 8.95, 2.55, '')

    # bottom note
    box(ax, 0.3, 0.3, 10.4, 0.85,
        '关键点：私钥不出浏览器；服务端只保存与转发密文；为发送方与接收方各保存一份密文，双方都能查看本地历史消息。',
        fc='#FFF8DC', ec='#999', fs=10)

    plt.tight_layout()
    fig.savefig(os.path.join(OUT, '04_端到端加密流程图.png'), bbox_inches='tight')
    plt.close(fig)


# ----------------- 5. 数据库 E-R 图 -----------------
def fig_er():
    fig, ax = plt.subplots(figsize=(11, 7), dpi=DPI)
    ax.set_xlim(0, 11); ax.set_ylim(0, 7); ax.axis('off')
    ax.text(5.5, 6.6, '数据库 E-R 图', ha='center', fontsize=14, fontweight='bold')

    def entity(x, y, w, h, title, fields):
        p = Rectangle((x, y), w, h, fc='#F6F8FB', ec='#0071E3', lw=1.2)
        ax.add_patch(p)
        ax.add_patch(Rectangle((x, y + h - 0.45), w, 0.45, fc='#0071E3', ec='#0071E3'))
        ax.text(x + w / 2, y + h - 0.22, title, ha='center', va='center',
                fontsize=11, color='white', fontweight='bold')
        for i, f in enumerate(fields):
            ax.text(x + 0.1, y + h - 0.72 - i * 0.28, f, fontsize=9)

    entity(0.4, 2.3, 2.8, 3.8, 'users (用户表)', [
        'id  BIGINT PK',
        'username  VARCHAR UNIQUE',
        'password  VARCHAR (bcrypt)',
        'nickname  VARCHAR',
        'avatar  VARCHAR',
        'gender  TINYINT',
        'signature  VARCHAR',
        'public_key  TEXT',
        'public_key_algorithm  VARCHAR',
        'created_at / updated_at',
    ])

    entity(4.0, 4.0, 3.0, 2.2, 'friend_requests (好友申请)', [
        'id  BIGINT PK',
        'from_user_id  BIGINT FK',
        'to_user_id    BIGINT FK',
        'status  TINYINT (待处理/同意/拒绝)',
        'created_at / updated_at',
    ])

    entity(4.0, 0.7, 3.0, 2.4, 'friends (好友关系)', [
        'id  BIGINT PK',
        'user_id   BIGINT FK',
        'friend_id BIGINT FK',
        'remark  VARCHAR',
        'created_at',
        '(user_id, friend_id) UNIQUE',
    ])

    entity(7.9, 1.8, 2.9, 3.6, 'messages (消息表)', [
        'id  BIGINT PK',
        'sender_id    BIGINT FK',
        'receiver_id  BIGINT FK',
        'sender_ciphertext   TEXT',
        'sender_algorithm    VARCHAR',
        'receiver_ciphertext TEXT',
        'receiver_algorithm  VARCHAR',
        'created_at',
    ])

    # relations
    def rel(x1, y1, x2, y2, t, side=0.2):
        arrow(ax, x1, y1, x2, y2, t, color='#2E8B57', fs=9, offset=(0, side))

    rel(3.2, 5.0, 4.0, 5.0, '1..N  发起')
    rel(3.2, 4.3, 4.0, 4.3, '1..N  接收', side=-0.18)
    rel(3.2, 2.7, 4.0, 2.0, '1..N  拥有好友')
    rel(3.2, 4.0, 7.9, 4.0, '1..N  发送 / 接收', side=0.18)

    plt.tight_layout()
    fig.savefig(os.path.join(OUT, '05_数据库ER图.png'), bbox_inches='tight')
    plt.close(fig)


# ----------------- 6. 模块结构图 -----------------
def fig_module():
    fig, ax = plt.subplots(figsize=(11, 7), dpi=DPI)
    ax.set_xlim(0, 11); ax.set_ylim(0, 7); ax.axis('off')
    ax.text(5.5, 6.6, '前后端模块结构图', ha='center', fontsize=14, fontweight='bold')

    # 前端
    box(ax, 0.3, 0.4, 5.0, 5.8, '', fc='#F6F8FB', ec='#0071E3', lw=1.2)
    ax.text(2.8, 5.95, '前端 web/', fontsize=12, fontweight='bold', color='#0071E3')

    box(ax, 0.5, 4.95, 4.6, 0.7, 'views/  注册 / 登录 / 资料 / 好友 / 聊天页面')
    box(ax, 0.5, 4.15, 4.6, 0.7, 'components/  AppNav、公用 UI 组件')
    box(ax, 0.5, 3.35, 4.6, 0.7, 'stores/  Pinia（auth 等）')
    box(ax, 0.5, 2.55, 4.6, 0.7, 'router/  Vue Router 鉴权守卫')
    box(ax, 0.5, 1.75, 4.6, 0.7, 'api/  http.ts（Axios 封装 + JWT 注入）')
    box(ax, 0.5, 0.95, 4.6, 0.7, 'utils/  e2ee.ts（加解密） + websocket.ts', fc='#FFE3E3', ec='#D43A3A')

    # 后端
    box(ax, 5.7, 0.4, 5.0, 5.8, '', fc='#F6F8FB', ec='#2E8B57', lw=1.2)
    ax.text(8.2, 5.95, '后端 server/', fontsize=12, fontweight='bold', color='#2E8B57')

    box(ax, 5.9, 4.95, 4.6, 0.7, 'cmd/main.go  入口，读取 config.yaml')
    box(ax, 5.9, 4.15, 4.6, 0.7, 'router/  路由注册（/api/* 与 /ws）')
    box(ax, 5.9, 3.35, 4.6, 0.7, 'middleware/  JWT 鉴权、CORS')
    box(ax, 5.9, 2.55, 4.6, 0.7, 'handler/  auth / user / friend / message / ws')
    box(ax, 5.9, 1.75, 4.6, 0.7, 'service/  业务逻辑（密文校验、好友关系等）')
    box(ax, 5.9, 0.95, 4.6, 0.7, 'repository/ + model/  GORM 数据访问层')

    plt.tight_layout()
    fig.savefig(os.path.join(OUT, '06_模块结构图.png'), bbox_inches='tight')
    plt.close(fig)


if __name__ == '__main__':
    fig_arch()
    fig_usecase()
    fig_ws_sequence()
    fig_e2ee()
    fig_er()
    fig_module()
    print('All diagrams generated to:', OUT)
