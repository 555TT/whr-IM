"""生成毕业论文中所有架构图、流程图、ER 图、类图、活动图。
全部使用 matplotlib 绘制并输出到 paper/picture/ 目录。
"""
import os
import matplotlib
matplotlib.use("Agg")
import matplotlib.pyplot as plt
import matplotlib.font_manager as fm
import matplotlib.patches as patches
from matplotlib.patches import FancyBboxPatch, FancyArrowPatch, Rectangle, Circle, Ellipse, Polygon
from matplotlib.lines import Line2D

# 注册字体: Times New Roman (英文/数字) + KaiTi_GB2312 (论文要求的楷体)
def _register_font(path):
    if os.path.exists(path):
        try:
            fm.fontManager.addfont(path)
        except Exception:
            pass

_register_font("/System/Library/Fonts/Supplemental/Times New Roman.ttf")
_register_font("/System/Library/Fonts/Supplemental/Times New Roman Bold.ttf")
# 论文指定的 KaiTi_GB2312 字体文件,与 paper/毕业论文.md 同目录
_register_font(os.path.join(os.path.dirname(__file__), "楷体_GB2312.TTF"))

# 字号:14px ≈ 五号(10.5pt)。matplotlib 默认按 pt 计算,所以设置成 10.5。
FS = 10.5

# 字体回退链: Times New Roman 处理英文与数字, 中文交给 KaiTi_GB2312
plt.rcParams["font.family"] = [
    "Times New Roman",
    "KaiTi_GB2312",
    "serif",
]
plt.rcParams["font.size"] = FS
plt.rcParams["axes.unicode_minus"] = False

PIC_DIR = os.path.join(os.path.dirname(__file__), "picture")
os.makedirs(PIC_DIR, exist_ok=True)

# 通用配色
C_BG = "#FFFFFF"
C_HEAD = "#4A6FA5"      # 蓝色标题
C_BLOCK = "#E8F0F8"     # 浅蓝
C_BLOCK_2 = "#FFF4D6"   # 浅黄
C_BLOCK_3 = "#E6F3E6"   # 浅绿
C_BLOCK_4 = "#FCE5E5"   # 浅红
C_BLOCK_5 = "#EDE5FB"   # 浅紫
C_LINE = "#333333"
C_ARROW = "#444444"

def save(fig, name, dpi=200):
    fig.savefig(os.path.join(PIC_DIR, name), dpi=dpi, bbox_inches="tight", facecolor="white")
    plt.close(fig)
    print(f"saved {name}")


def box(ax, x, y, w, h, text, color=C_BLOCK, edge=C_HEAD, fontsize=FS, weight="normal"):
    p = FancyBboxPatch((x, y), w, h, boxstyle="round,pad=0.02,rounding_size=0.08",
                       linewidth=1.2, edgecolor=edge, facecolor=color)
    ax.add_patch(p)
    ax.text(x + w / 2, y + h / 2, text, ha="center", va="center",
            fontsize=fontsize, fontweight=weight, wrap=True)


def rect(ax, x, y, w, h, text, color=C_BLOCK, edge=C_HEAD, fontsize=FS, weight="normal"):
    p = Rectangle((x, y), w, h, linewidth=1.2, edgecolor=edge, facecolor=color)
    ax.add_patch(p)
    ax.text(x + w / 2, y + h / 2, text, ha="center", va="center",
            fontsize=fontsize, fontweight=weight)


def diamond(ax, x, y, w, h, text, color=C_BLOCK_2, edge=C_HEAD, fontsize=FS):
    pts = [(x + w / 2, y + h), (x + w, y + h / 2), (x + w / 2, y), (x, y + h / 2)]
    p = Polygon(pts, closed=True, linewidth=1.2, edgecolor=edge, facecolor=color)
    ax.add_patch(p)
    ax.text(x + w / 2, y + h / 2, text, ha="center", va="center", fontsize=fontsize)


def ellipse(ax, cx, cy, w, h, text, color=C_BLOCK_3, edge=C_HEAD, fontsize=FS):
    e = Ellipse((cx, cy), w, h, linewidth=1.2, edgecolor=edge, facecolor=color)
    ax.add_patch(e)
    ax.text(cx, cy, text, ha="center", va="center", fontsize=fontsize)


def circle_node(ax, cx, cy, r, text="", color="#222", fontsize=FS):
    c = Circle((cx, cy), r, facecolor=color, edgecolor=color)
    ax.add_patch(c)
    if text:
        ax.text(cx, cy, text, ha="center", va="center", fontsize=fontsize, color="white")


def arrow(ax, x1, y1, x2, y2, text="", color=C_ARROW, style="-|>", lw=1.2, rad=0.0):
    a = FancyArrowPatch((x1, y1), (x2, y2), arrowstyle=style,
                        mutation_scale=14, color=color, lw=lw,
                        connectionstyle=f"arc3,rad={rad}")
    ax.add_patch(a)
    if text:
        ax.text((x1 + x2) / 2, (y1 + y2) / 2 + 0.06, text, ha="center", va="bottom",
                fontsize=FS, color=color, bbox=dict(facecolor="white", edgecolor="none", pad=0.5))


# ---------------------------------------------------------------------------
# 图 3-1 系统整体业务流程图
# ---------------------------------------------------------------------------
def gen_fig_3_1():
    fig, ax = plt.subplots(figsize=(14, 10))
    ax.set_xlim(0, 14); ax.set_ylim(0, 10); ax.axis("off")
    # 起点
    ellipse(ax, 7, 9.4, 1.6, 0.6, "开始", color="#333", edge="#333", fontsize=FS)
    ax.texts[-1].set_color("white")

    # 阶段 1：账号准入
    box(ax, 5.5, 8.3, 3, 0.6, "用户输入用户名/密码", C_BLOCK)
    diamond(ax, 5.7, 7.2, 2.6, 0.8, "是否已注册?", C_BLOCK_2)
    box(ax, 1.0, 7.3, 2.8, 0.6, "完成注册 → 写入 users", C_BLOCK_3)
    box(ax, 5.5, 6.1, 3, 0.6, "POST /api/auth/login", C_BLOCK)
    box(ax, 5.5, 5.1, 3, 0.6, "bcrypt 校验 + 签发 JWT", C_BLOCK)

    # 阶段 2：密钥准备
    box(ax, 5.0, 4.0, 4, 0.6, "前端检查/生成 RSA 密钥对\n私钥保存本地、公钥上传服务端", C_BLOCK_5, fontsize=FS)

    # 阶段 3：社交关系
    box(ax, 0.6, 2.9, 3.2, 0.6, "发起好友申请\n(POST /api/friend-requests)", C_BLOCK, fontsize=FS)
    box(ax, 5.3, 2.9, 3.4, 0.6, "对方处理申请\n(accept / reject)", C_BLOCK, fontsize=FS)
    box(ax, 10.0, 2.9, 3.2, 0.6, "建立双向好友关系\n写入 friends 表", C_BLOCK_3, fontsize=FS)

    # 阶段 4：实时通信
    box(ax, 0.6, 1.7, 3.5, 0.6, "单聊：RSA-OAEP 双密文加密", C_BLOCK_4, fontsize=FS)
    box(ax, 5.0, 1.7, 4, 0.6, "群聊：AES-GCM + RSA-OAEP 混合加密", C_BLOCK_4, fontsize=FS)
    box(ax, 9.8, 1.7, 3.5, 0.6, "WebSocket 实时推送密文", C_BLOCK_4, fontsize=FS)

    # 阶段 5：朋友圈
    box(ax, 1.8, 0.5, 4, 0.6, "发布朋友圈 / 点赞 / 评论", C_BLOCK_2, fontsize=FS)
    box(ax, 8.0, 0.5, 4, 0.6, "调用 DeepSeek AI 生成 / 润色文案", C_BLOCK_2, fontsize=FS)

    # 连线
    arrow(ax, 7, 9.4 - 0.3, 7, 8.9)
    arrow(ax, 7, 8.3, 7, 8.0)
    arrow(ax, 5.7, 7.6, 3.8, 7.6, text="否")
    arrow(ax, 7, 7.2, 7, 6.7, text="是")
    arrow(ax, 2.4, 7.3, 2.4, 4.6)
    arrow(ax, 2.4, 4.6, 5.0, 4.3)
    arrow(ax, 7, 6.1, 7, 5.7)
    arrow(ax, 7, 5.1, 7, 4.6)
    arrow(ax, 7, 4.0, 7, 3.5)
    arrow(ax, 5.5, 3.2, 5.3, 3.2, text="")
    arrow(ax, 3.8, 3.2, 5.3, 3.2)
    arrow(ax, 8.7, 3.2, 10.0, 3.2)
    arrow(ax, 7, 2.9, 7, 2.3)
    arrow(ax, 7, 1.7, 7, 1.1)

    save(fig, "图3-1_系统整体业务流程图.png")


# ---------------------------------------------------------------------------
# 图 4-1 系统功能架构图
# ---------------------------------------------------------------------------
def gen_fig_4_1():
    fig, ax = plt.subplots(figsize=(14, 9))
    ax.set_xlim(0, 14); ax.set_ylim(0, 9); ax.axis("off")
    # 表现层
    rect(ax, 0.3, 7.6, 13.4, 1.0, "", C_BLOCK, fontsize=FS)
    ax.text(0.5, 8.4, "表现层", fontsize=FS, fontweight="bold", color=C_HEAD)
    presents = ["登录注册页", "主聊天页", "好友申请页", "个人资料页", "好友主页", "朋友圈页", "全局导航栏"]
    for i, t in enumerate(presents):
        rect(ax, 1.6 + i * 1.72, 7.75, 1.55, 0.55, t, "#FFFFFF", "#666", fontsize=FS)

    # 业务模块层
    rect(ax, 0.3, 5.4, 13.4, 1.9, "", C_BLOCK_2, fontsize=FS)
    ax.text(0.5, 7.1, "业务模块层", fontsize=FS, fontweight="bold", color=C_HEAD)
    modules = [
        "用户认证模块", "个人资料模块", "好友主页模块",
        "好友关系模块", "单聊模块", "群聊模块",
        "朋友圈动态模块", "AI 文案助手模块", "图片上传模块"
    ]
    for i, t in enumerate(modules):
        row, col = i // 3, i % 3
        rect(ax, 1.6 + col * 4.2, 6.45 - row * 0.6, 3.95, 0.45, t,
             "#FFFFFF", "#888", fontsize=FS)

    # 公共能力层
    rect(ax, 0.3, 3.8, 13.4, 1.3, "", C_BLOCK_3, fontsize=FS)
    ax.text(0.5, 4.95, "公共能力层", fontsize=FS, fontweight="bold", color=C_HEAD)
    commons = ["JWT 鉴权中间件", "CORS 中间件", "统一错误处理",
               "WebSocket Hub", "加密工具集", "对象存储客户端",
               "DeepSeek HTTP 客户端", "bcrypt 密码哈希"]
    for i, t in enumerate(commons):
        row, col = i // 4, i % 4
        rect(ax, 1.6 + col * 3.1, 4.4 - row * 0.45, 2.9, 0.35, t,
             "#FFFFFF", "#888", fontsize=FS)

    # 数据访问层
    rect(ax, 0.3, 2.7, 13.4, 0.8, "", C_BLOCK_5, fontsize=FS)
    ax.text(0.5, 3.1, "数据访问层", fontsize=FS, fontweight="bold", color=C_HEAD)
    repos = ["UserRepo", "FriendRepo", "MessageRepo", "GroupRepo",
             "GroupMessageRepo", "MomentRepo"]
    for i, t in enumerate(repos):
        rect(ax, 1.6 + i * 2.05, 2.85, 1.92, 0.45, t, "#FFFFFF", "#888", fontsize=FS)

    # 数据持久层
    rect(ax, 0.3, 0.8, 13.4, 1.6, "", C_BLOCK_4, fontsize=FS)
    ax.text(0.5, 2.25, "数据持久层", fontsize=FS, fontweight="bold", color=C_HEAD)
    rect(ax, 2.0, 1.0, 4.5, 1.2,
         "MySQL 8.0\n(users / friends / messages /\nchat_groups / moments / ...)",
         "#FFFFFF", "#888", fontsize=FS)
    rect(ax, 7.5, 1.0, 4.5, 1.2,
         "MinIO 对象存储\n(uploads/users/{userId}/*.png)",
         "#FFFFFF", "#888", fontsize=FS)

    save(fig, "图4-1_系统功能架构图.png")


# ---------------------------------------------------------------------------
# 图 4-2 系统业务架构图
# ---------------------------------------------------------------------------
def gen_fig_4_2():
    fig, ax = plt.subplots(figsize=(14, 9))
    ax.set_xlim(0, 14); ax.set_ylim(0, 9); ax.axis("off")
    # 左边的安全保障体系
    rect(ax, 0.2, 1.0, 1.4, 7.4, "信\n息\n安\n全\n保\n障\n体\n系",
         "#FBE1B6", "#C99830", fontsize=FS, weight="bold")
    # 右边规范体系
    rect(ax, 12.4, 1.0, 1.4, 7.4, "标\n准\n规\n范\n体\n系",
         "#FBE1B6", "#C99830", fontsize=FS, weight="bold")

    # 表现层
    rect(ax, 1.8, 7.2, 10.4, 1.2, "", C_BLOCK)
    ax.text(2.0, 8.1, "表现层 (Presentation Layer)", fontsize=FS, fontweight="bold", color=C_HEAD)
    items_p = ["登录注册", "主聊天", "好友申请", "个人资料", "好友主页", "朋友圈"]
    for i, t in enumerate(items_p):
        rect(ax, 2.0 + i * 1.65, 7.32, 1.55, 0.6, t, "#FFFFFF", "#888", fontsize=FS)

    # 业务应用层
    rect(ax, 1.8, 5.4, 10.4, 1.6, "", C_BLOCK_2)
    ax.text(2.0, 6.85, "业务应用层 (Business Application Layer)", fontsize=FS, fontweight="bold", color=C_HEAD)
    items_b = ["用户认证", "个人资料", "好友主页",
               "好友关系", "单聊", "群聊",
               "朋友圈动态", "AI 文案助手", "图片上传"]
    for i, t in enumerate(items_b):
        row, col = i // 3, i % 3
        rect(ax, 2.0 + col * 3.45, 6.25 - row * 0.42, 3.3, 0.32, t,
             "#FFFFFF", "#888", fontsize=FS)

    # 应用支撑层
    rect(ax, 1.8, 3.4, 10.4, 1.6, "", C_BLOCK_3)
    ax.text(2.0, 4.85, "应用支撑层 (Application Support Layer)", fontsize=FS, fontweight="bold", color=C_HEAD)
    items_s = ["JWT 鉴权", "CORS", "错误处理",
               "WebSocket Hub", "加密工具", "MinIO 客户端",
               "DeepSeek 客户端", "日志", "配置加载"]
    for i, t in enumerate(items_s):
        row, col = i // 3, i % 3
        rect(ax, 2.0 + col * 3.45, 4.25 - row * 0.42, 3.3, 0.32, t,
             "#FFFFFF", "#888", fontsize=FS)

    # 数据资源层
    rect(ax, 1.8, 1.0, 10.4, 2.0, "", C_BLOCK_5)
    ax.text(2.0, 2.85, "数据资源层 (Data Resource Layer)", fontsize=FS, fontweight="bold", color=C_HEAD)
    rect(ax, 2.0, 1.2, 3.2, 1.5,
         "基础信息库\n用户/公钥/\n好友关系",
         "#FFFFFF", "#888", fontsize=FS)
    rect(ax, 5.5, 1.2, 3.2, 1.5,
         "业务信息库\n消息/群组/\n朋友圈",
         "#FFFFFF", "#888", fontsize=FS)
    rect(ax, 9.0, 1.2, 3.2, 1.5,
         "对象资源库\nMinIO 朋友圈\n图片",
         "#FFFFFF", "#888", fontsize=FS)

    save(fig, "图4-2_系统业务架构图.png")


# ---------------------------------------------------------------------------
# 图 4-3 系统技术架构图
# ---------------------------------------------------------------------------
def gen_fig_4_3():
    fig, ax = plt.subplots(figsize=(14, 9))
    ax.set_xlim(0, 14); ax.set_ylim(0, 9); ax.axis("off")
    # 前端展示层
    rect(ax, 0.4, 7.4, 13.2, 1.2, "", C_BLOCK)
    ax.text(0.6, 8.35, "前端展示层 (Browser)", fontsize=FS, fontweight="bold", color=C_HEAD)
    techs = ["Vue 3", "Vite 5", "TypeScript", "Pinia", "Vue Router", "Axios",
             "WebSocket API", "Web Crypto API"]
    for i, t in enumerate(techs):
        rect(ax, 0.7 + i * 1.65, 7.5, 1.55, 0.5, t, "#FFFFFF", "#666", fontsize=FS)

    # 网络通信层
    rect(ax, 0.4, 5.9, 13.2, 1.1, "", C_BLOCK_2)
    ax.text(0.6, 6.85, "网络通信层", fontsize=FS, fontweight="bold", color=C_HEAD)
    rect(ax, 2.5, 6.0, 4.0, 0.6, "HTTP / HTTPS (RESTful API)", "#FFFFFF", "#888", fontsize=FS)
    rect(ax, 7.5, 6.0, 4.0, 0.6, "WebSocket (实时推送)", "#FFFFFF", "#888", fontsize=FS)

    # 后端服务层
    rect(ax, 0.4, 3.5, 13.2, 2.0, "", C_BLOCK_3)
    ax.text(0.6, 5.3, "后端服务层 (Go)", fontsize=FS, fontweight="bold", color=C_HEAD)
    rect(ax, 1.0, 4.4, 5.5, 0.6, "Handler 层 (Gin 路由 + 参数校验)", "#FFFFFF", "#888", fontsize=FS)
    rect(ax, 7.0, 4.4, 6.0, 0.6, "Middleware 层 (JWT / CORS / Error)", "#FFFFFF", "#888", fontsize=FS)
    rect(ax, 1.0, 3.7, 5.5, 0.6, "Service 层 (业务逻辑编排)", "#FFFFFF", "#888", fontsize=FS)
    rect(ax, 7.0, 3.7, 6.0, 0.6, "WebSocket Hub (在线连接管理)", "#FFFFFF", "#888", fontsize=FS)

    # 数据持久层
    rect(ax, 0.4, 1.9, 13.2, 1.2, "", C_BLOCK_5)
    ax.text(0.6, 2.95, "数据持久层", fontsize=FS, fontweight="bold", color=C_HEAD)
    rect(ax, 1.5, 2.0, 4.5, 0.7, "GORM ORM 框架", "#FFFFFF", "#888", fontsize=FS)
    rect(ax, 6.5, 2.0, 4.5, 0.7, "MinIO Go SDK", "#FFFFFF", "#888", fontsize=FS)

    # 数据库层 + 外部服务
    rect(ax, 0.4, 0.4, 8.0, 1.1, "", C_BLOCK_4)
    ax.text(0.6, 1.35, "数据存储", fontsize=FS, fontweight="bold", color=C_HEAD)
    rect(ax, 1.0, 0.5, 3.0, 0.7, "MySQL 8.0", "#FFFFFF", "#888", fontsize=FS)
    rect(ax, 4.5, 0.5, 3.7, 0.7, "MinIO 对象存储", "#FFFFFF", "#888", fontsize=FS)

    rect(ax, 8.8, 0.4, 4.8, 1.1, "", "#FFE4E1")
    ax.text(9.0, 1.35, "外部服务", fontsize=FS, fontweight="bold", color=C_HEAD)
    rect(ax, 9.0, 0.5, 4.4, 0.7, "DeepSeek Chat Completions API", "#FFFFFF", "#888", fontsize=FS)

    save(fig, "图4-3_系统技术架构图.png")


# ---------------------------------------------------------------------------
# 图 4-4 用户登录活动图
# ---------------------------------------------------------------------------
def gen_fig_4_4():
    fig, ax = plt.subplots(figsize=(10, 12))
    ax.set_xlim(0, 10); ax.set_ylim(0, 12); ax.axis("off")
    circle_node(ax, 5, 11.4, 0.25, "", "#333")

    box(ax, 3.3, 10.3, 3.4, 0.6, "用户输入用户名密码")
    box(ax, 3.3, 9.3, 3.4, 0.6, "前端校验长度合法性")
    diamond(ax, 3.7, 8.0, 2.6, 1.0, "长度合法?")
    box(ax, 7.0, 8.3, 2.6, 0.5, "前端提示并阻断", C_BLOCK_4)

    box(ax, 3.3, 6.8, 3.4, 0.6, "POST /api/auth/login")
    box(ax, 3.3, 5.8, 3.4, 0.6, "后端查询 users 表")
    diamond(ax, 3.7, 4.5, 2.6, 1.0, "用户名存在?")
    box(ax, 7.0, 4.8, 2.6, 0.5, "返回 401", C_BLOCK_4)

    box(ax, 3.3, 3.2, 3.4, 0.6, "bcrypt 校验密码哈希")
    diamond(ax, 3.7, 1.9, 2.6, 1.0, "密码正确?")
    box(ax, 7.0, 2.2, 2.6, 0.5, "返回 401", C_BLOCK_4)

    box(ax, 3.3, 0.6, 3.4, 0.6, "签发 JWT、返回用户信息", C_BLOCK_3)
    circle_node(ax, 5, 0.1, 0.18, "", "#333")
    circle_node(ax, 5, 0.1, 0.10, "", "white")

    arrow(ax, 5, 11.15, 5, 10.9)
    arrow(ax, 5, 10.3, 5, 9.9)
    arrow(ax, 5, 9.3, 5, 9.0)
    arrow(ax, 6.3, 8.5, 7.0, 8.55, text="否")
    arrow(ax, 5, 8.0, 5, 7.4, text="是")
    arrow(ax, 5, 6.8, 5, 6.4)
    arrow(ax, 5, 5.8, 5, 5.5)
    arrow(ax, 6.3, 5.0, 7.0, 5.05, text="否")
    arrow(ax, 5, 4.5, 5, 3.8, text="是")
    arrow(ax, 5, 3.2, 5, 2.9)
    arrow(ax, 6.3, 2.4, 7.0, 2.45, text="否")
    arrow(ax, 5, 1.9, 5, 1.2, text="是")
    arrow(ax, 5, 0.6, 5, 0.3)

    save(fig, "图4-4_用户登录活动图.png")


# ---------------------------------------------------------------------------
# 图 4-5 单聊消息发送活动图
# ---------------------------------------------------------------------------
def gen_fig_4_5():
    fig, ax = plt.subplots(figsize=(14, 11))
    ax.set_xlim(0, 14); ax.set_ylim(0, 11); ax.axis("off")
    # 三个泳道：发送方前端 / 后端 / 接收方前端
    rect(ax, 0.2, 0.4, 4.5, 10.2, "", "#FAFAFA", "#888")
    rect(ax, 4.8, 0.4, 4.5, 10.2, "", "#FAFAFA", "#888")
    rect(ax, 9.4, 0.4, 4.5, 10.2, "", "#FAFAFA", "#888")
    ax.text(2.45, 10.3, "发送方浏览器", fontsize=FS, fontweight="bold", color=C_HEAD, ha="center")
    ax.text(7.05, 10.3, "服务端 (Go/Gin)", fontsize=FS, fontweight="bold", color=C_HEAD, ha="center")
    ax.text(11.65, 10.3, "接收方浏览器", fontsize=FS, fontweight="bold", color=C_HEAD, ha="center")

    # 发送方
    circle_node(ax, 2.45, 9.7, 0.22, "", "#333")
    box(ax, 0.6, 8.7, 3.7, 0.55, "用户输入明文")
    box(ax, 0.6, 7.8, 3.7, 0.55, "加载本地 RSA 私钥/好友公钥", C_BLOCK_5)
    box(ax, 0.6, 6.6, 3.7, 0.9, "encryptRSA(text, 自己公钥)\n→ senderCiphertext", C_BLOCK_5, fontsize=FS)
    box(ax, 0.6, 5.4, 3.7, 0.9, "encryptRSA(text, 好友公钥)\n→ receiverCiphertext", C_BLOCK_5, fontsize=FS)
    box(ax, 0.6, 4.4, 3.7, 0.55, "POST /api/messages")

    # 后端
    box(ax, 5.2, 8.3, 3.7, 0.6, "校验 JWT 身份", C_BLOCK_3)
    box(ax, 5.2, 7.4, 3.7, 0.6, "校验好友关系")
    box(ax, 5.2, 6.5, 3.7, 0.6, "校验算法白名单")
    box(ax, 5.2, 5.4, 3.7, 0.7, "messages 表持久化\n（仅密文与算法标识）", C_BLOCK_4, fontsize=FS)
    box(ax, 5.2, 4.4, 3.7, 0.55, "Hub.SendToUser(receiverID)")
    box(ax, 5.2, 3.4, 3.7, 0.55, "返回响应 (双份密文)")

    # 接收方
    box(ax, 9.8, 6.1, 3.7, 0.6, "WebSocket 收到 chat_message")
    box(ax, 9.8, 5.0, 3.7, 0.8, "decryptRSA(receiverCiphertext,\n本地私钥) → 明文", C_BLOCK_5, fontsize=FS)
    box(ax, 9.8, 4.1, 3.7, 0.55, "渲染到聊天窗口")

    # 发送方收到响应解密
    box(ax, 0.6, 3.3, 3.7, 0.8, "decryptRSA(senderCiphertext,\n本地私钥) → 明文", C_BLOCK_5, fontsize=FS)
    box(ax, 0.6, 2.3, 3.7, 0.55, "渲染到聊天窗口")
    circle_node(ax, 2.45, 1.7, 0.22, "", "#333")
    circle_node(ax, 2.45, 1.7, 0.10, "", "white")

    circle_node(ax, 11.65, 3.6, 0.22, "", "#333")
    circle_node(ax, 11.65, 3.6, 0.10, "", "white")

    # 箭头
    arrow(ax, 2.45, 9.5, 2.45, 9.25)
    arrow(ax, 2.45, 8.7, 2.45, 8.35)
    arrow(ax, 2.45, 7.8, 2.45, 7.5)
    arrow(ax, 2.45, 6.6, 2.45, 6.3)
    arrow(ax, 2.45, 5.4, 2.45, 4.95)
    arrow(ax, 4.3, 4.6, 5.2, 8.6, text="HTTP", rad=-0.3)
    arrow(ax, 7.05, 8.3, 7.05, 8.0)
    arrow(ax, 7.05, 7.4, 7.05, 7.1)
    arrow(ax, 7.05, 6.5, 7.05, 6.1)
    arrow(ax, 7.05, 5.4, 7.05, 4.95)
    arrow(ax, 8.9, 4.65, 9.8, 6.4, text="WS 推送", rad=0.3)
    arrow(ax, 7.05, 4.4, 7.05, 4.0)
    arrow(ax, 5.2, 3.55, 4.3, 3.7, text="响应", rad=-0.2)
    arrow(ax, 2.45, 3.3, 2.45, 2.85)
    arrow(ax, 2.45, 2.3, 2.45, 1.9)
    arrow(ax, 11.65, 6.1, 11.65, 5.8)
    arrow(ax, 11.65, 5.0, 11.65, 4.65)
    arrow(ax, 11.65, 4.1, 11.65, 3.8)

    save(fig, "图4-5_单聊消息发送活动图.png")


# ---------------------------------------------------------------------------
# 图 4-6 群聊消息发送活动图
# ---------------------------------------------------------------------------
def gen_fig_4_6():
    fig, ax = plt.subplots(figsize=(14, 12))
    ax.set_xlim(0, 14); ax.set_ylim(0, 12); ax.axis("off")
    rect(ax, 0.2, 0.4, 4.5, 11.2, "", "#FAFAFA", "#888")
    rect(ax, 4.8, 0.4, 4.5, 11.2, "", "#FAFAFA", "#888")
    rect(ax, 9.4, 0.4, 4.5, 11.2, "", "#FAFAFA", "#888")
    ax.text(2.45, 11.3, "发送方浏览器", fontsize=FS, fontweight="bold", color=C_HEAD, ha="center")
    ax.text(7.05, 11.3, "服务端 (Go/Gin)", fontsize=FS, fontweight="bold", color=C_HEAD, ha="center")
    ax.text(11.65, 11.3, "群成员浏览器 (N 端)", fontsize=FS, fontweight="bold", color=C_HEAD, ha="center")

    circle_node(ax, 2.45, 10.7, 0.22, "", "#333")
    box(ax, 0.6, 9.8, 3.7, 0.55, "用户输入群消息明文")
    box(ax, 0.6, 8.8, 3.7, 0.7, "生成一次性 AES-GCM-256\n会话密钥 + 随机 IV", C_BLOCK_5, fontsize=FS)
    box(ax, 0.6, 7.7, 3.7, 0.7, "AES 加密正文\n→ contentCiphertext", C_BLOCK_5, fontsize=FS)
    box(ax, 0.6, 6.4, 3.7, 0.9, "遍历群成员公钥\n用每位成员公钥 RSA 加密 AES key\n→ N 份 keyCiphertext", C_BLOCK_5, fontsize=FS)
    box(ax, 0.6, 5.4, 3.7, 0.55, "POST /api/groups/{id}/messages")

    box(ax, 5.2, 9.6, 3.7, 0.6, "校验发送者是群成员", C_BLOCK_3)
    box(ax, 5.2, 8.7, 3.7, 0.7, "校验 memberKeys 完整覆盖\n当前群成员")
    box(ax, 5.2, 7.7, 3.7, 0.6, "校验算法白名单")
    box(ax, 5.2, 6.6, 3.7, 0.7, "事务：写 group_messages +\nN 条 group_message_keys", C_BLOCK_4, fontsize=FS)
    box(ax, 5.2, 5.4, 3.7, 0.7, "Hub.BroadcastToGroup\n按成员分发不同 keyCiphertext")
    box(ax, 5.2, 4.3, 3.7, 0.55, "返回响应")

    box(ax, 9.8, 7.5, 3.7, 0.6, "WS 收到 group_message")
    box(ax, 9.8, 6.4, 3.7, 0.9, "decryptRSA(keyCiphertext,\n本地私钥) → AES 会话密钥", C_BLOCK_5, fontsize=FS)
    box(ax, 9.8, 5.2, 3.7, 0.9, "decryptAES(contentCiphertext,\nAES key, iv) → 明文", C_BLOCK_5, fontsize=FS)
    box(ax, 9.8, 4.3, 3.7, 0.55, "渲染到群聊窗口")
    circle_node(ax, 11.65, 3.7, 0.22, "", "#333")
    circle_node(ax, 11.65, 3.7, 0.10, "", "white")

    # 发送方本地解密
    box(ax, 0.6, 4.2, 3.7, 0.8, "用自己的 keyCiphertext +\ncontentCiphertext 本地解密", C_BLOCK_5, fontsize=FS)
    box(ax, 0.6, 3.2, 3.7, 0.55, "渲染到群聊窗口")
    circle_node(ax, 2.45, 2.6, 0.22, "", "#333")
    circle_node(ax, 2.45, 2.6, 0.10, "", "white")

    # 箭头
    for y1, y2 in [(10.5, 10.35), (9.8, 9.5), (8.8, 8.45), (7.7, 7.35), (6.4, 5.95)]:
        arrow(ax, 2.45, y1, 2.45, y2)
    arrow(ax, 4.3, 5.6, 5.2, 9.9, text="HTTP", rad=-0.3)
    for y1, y2 in [(9.6, 9.4), (8.7, 8.35), (7.7, 7.35), (6.6, 6.15), (5.4, 4.9)]:
        arrow(ax, 7.05, y1, 7.05, y2)
    arrow(ax, 8.9, 5.7, 9.8, 7.8, text="WS 推送", rad=0.3)
    arrow(ax, 7.05, 4.3, 7.05, 4.0)
    arrow(ax, 5.2, 4.5, 4.3, 4.6, text="响应", rad=-0.2)
    arrow(ax, 2.45, 4.2, 2.45, 3.75)
    arrow(ax, 2.45, 3.2, 2.45, 2.8)
    arrow(ax, 11.65, 7.5, 11.65, 7.3)
    arrow(ax, 11.65, 6.4, 11.65, 6.1)
    arrow(ax, 11.65, 5.2, 11.65, 4.85)
    arrow(ax, 11.65, 4.3, 11.65, 3.9)

    save(fig, "图4-6_群聊消息发送活动图.png")


# ---------------------------------------------------------------------------
# 图 4-7 E-R 图
# ---------------------------------------------------------------------------
def gen_fig_4_7():
    fig, ax = plt.subplots(figsize=(15, 11))
    ax.set_xlim(0, 15); ax.set_ylim(0, 11); ax.axis("off")
    # 实体（矩形）
    def entity(cx, cy, w, h, name):
        rect(ax, cx - w / 2, cy - h / 2, w, h, name, C_BLOCK, C_HEAD, fontsize=FS, weight="bold")

    # 关系（菱形）
    def rel(cx, cy, w, h, name, color=C_BLOCK_2):
        pts = [(cx, cy + h / 2), (cx + w / 2, cy), (cx, cy - h / 2), (cx - w / 2, cy)]
        p = Polygon(pts, closed=True, linewidth=1.2, edgecolor=C_HEAD, facecolor=color)
        ax.add_patch(p)
        ax.text(cx, cy, name, ha="center", va="center", fontsize=FS)

    def line(x1, y1, x2, y2, t=""):
        ax.plot([x1, x2], [y1, y2], "-", color=C_LINE, lw=1.0)
        if t:
            ax.text((x1 + x2) / 2, (y1 + y2) / 2, t, ha="center", va="center",
                    fontsize=FS, color="#900", bbox=dict(facecolor="white", edgecolor="none", pad=0.2))

    # 中央 User 实体
    entity(7.5, 5.5, 2.0, 0.8, "用户 User")

    # 周围实体
    entity(2.0, 9.5, 2.4, 0.8, "好友申请 FriendRequest")
    entity(7.5, 9.5, 2.0, 0.8, "好友关系 Friend")
    entity(13.0, 9.5, 2.0, 0.8, "群组 ChatGroup")
    entity(13.0, 6.5, 2.4, 0.8, "群成员 GroupMember")
    entity(13.0, 3.5, 2.4, 0.8, "群消息 GroupMessage")
    entity(11.0, 1.0, 2.6, 0.8, "群消息密钥 GroupMessageKey")
    entity(7.5, 1.5, 2.0, 0.8, "消息 Message")
    entity(2.0, 1.5, 2.0, 0.8, "朋友圈 Moment")
    entity(2.0, 4.5, 2.4, 0.8, "朋友圈点赞 MomentLike")
    entity(2.0, 7.0, 2.4, 0.8, "朋友圈评论 MomentComment")

    # 关系菱形
    rel(4.5, 8.0, 1.2, 0.6, "发起/接收")
    rel(7.5, 7.6, 1.0, 0.6, "拥有")
    rel(10.5, 7.6, 1.0, 0.6, "拥有")
    rel(13.0, 8.0, 1.2, 0.6, "包含")
    rel(13.0, 5.0, 1.2, 0.6, "发送/接收")
    rel(13.0, 2.2, 1.0, 0.6, "包裹")
    rel(7.5, 3.4, 1.0, 0.6, "发送/接收")
    rel(4.5, 3.0, 1.0, 0.6, "发布")
    rel(4.5, 5.5, 1.0, 0.6, "点赞")
    rel(4.5, 6.5, 1.0, 0.6, "评论")

    # 连线
    line(6.5, 5.5, 5.1, 5.5, "1")
    line(3.9, 5.5, 3.2, 4.7, "N")
    line(6.5, 5.5, 5.0, 6.5, "1")
    line(4.0, 6.5, 3.2, 6.8, "N")
    line(6.5, 5.5, 5.0, 3.0, "1")
    line(4.0, 3.0, 2.5, 1.9, "N")
    line(7.5, 4.9, 7.5, 3.7, "1")
    line(7.5, 3.1, 7.5, 1.9, "N")
    line(7.5, 6.1, 7.5, 7.3, "1")
    line(7.5, 7.9, 7.5, 9.1, "N")
    line(6.5, 5.5, 5.0, 8.0, "1")
    line(4.0, 8.0, 3.2, 9.1, "N")
    line(8.5, 5.5, 9.9, 7.6, "1")
    line(11.1, 7.6, 12.0, 7.6, "")
    line(13.0, 7.6, 13.0, 6.9, "拥有 N")
    line(13.0, 8.3, 13.0, 9.1, "1")
    line(8.5, 5.5, 11.8, 5.0, "1")
    line(13.0, 5.3, 13.0, 3.9, "N")
    line(13.0, 6.1, 13.0, 5.3, "1 (member)")
    line(13.0, 3.1, 13.0, 2.5, "1")
    line(12.5, 2.2, 12.0, 1.4, "N")

    save(fig, "图4-7_系统ER设计图.png")


# ---------------------------------------------------------------------------
# 类图生成器
# ---------------------------------------------------------------------------
def class_box(ax, x, y, w, name, attrs, methods, color=C_BLOCK):
    line_h = 0.36
    head_h = 0.45
    rect(ax, x, y, w, head_h, name, color, C_HEAD, fontsize=FS, weight="bold")
    # attributes
    y_attr = y - line_h * (len(attrs))
    rect(ax, x, y_attr, w, line_h * len(attrs), "", "#FFFFFF", C_HEAD)
    for i, a in enumerate(attrs):
        ax.text(x + 0.08, y_attr + line_h * (len(attrs) - i - 1) + line_h / 2,
                a, ha="left", va="center", fontsize=FS)
    # methods
    y_m = y_attr - line_h * (len(methods)) - 0.04
    rect(ax, x, y_m, w, line_h * len(methods), "", "#FAFAFA", C_HEAD)
    for i, m in enumerate(methods):
        ax.text(x + 0.08, y_m + line_h * (len(methods) - i - 1) + line_h / 2,
                m, ha="left", va="center", fontsize=FS)
    return (x, y_m, w, y + head_h - y_m)  # bounding


# ---------------------------------------------------------------------------
# 图 4-8 用户认证模块类图
# ---------------------------------------------------------------------------
def gen_fig_4_8():
    fig, ax = plt.subplots(figsize=(16, 11))
    ax.set_xlim(0, 16); ax.set_ylim(0, 11); ax.axis("off")
    class_box(ax, 0.5, 10.0, 4.6, "AuthHandler",
              ["- service: *AuthService"],
              ["+ Register(c *gin.Context)",
               "+ Login(c *gin.Context)",
               "+ Me(c *gin.Context)",
               "+ UpdateMe(c *gin.Context)",
               "+ UpdatePublicKey(c *gin.Context)"])

    class_box(ax, 5.6, 10.0, 4.8, "AuthService",
              ["- repo: UserRepository",
               "- jwtSecret: string"],
              ["+ Register(req) (*User, error)",
               "+ Login(req) (token, *User, err)",
               "+ UpdateProfile(uid, req) error",
               "+ UpdatePublicKey(uid, key) error",
               "+ VerifyToken(token) (uid, err)"])

    class_box(ax, 11.0, 10.0, 4.6, "UserRepository",
              ["- db: *gorm.DB"],
              ["+ Create(u *User) error",
               "+ GetByUsername(name) (*User, error)",
               "+ GetByID(id) (*User, error)",
               "+ Update(u *User) error"])

    class_box(ax, 0.5, 4.5, 4.6, "User (model)",
              ["+ ID uint64", "+ Username string",
               "+ PasswordHash string", "+ Nickname string",
               "+ Avatar string", "+ Gender int8",
               "+ Signature string", "+ HomepageSkin string",
               "+ PublicKey string", "+ PublicKeyAlgorithm string"],
              [])

    class_box(ax, 5.6, 4.5, 4.8, "AuthMiddleware",
              ["- service: *AuthService"],
              ["+ Handle(c *gin.Context)",
               "  解析 Authorization 头",
               "  校验 JWT 签名/过期",
               "  注入 c.Set(\"userId\", uid)"])

    class_box(ax, 11.0, 4.5, 4.6, "JWT 工具",
              ["+ secret []byte", "+ expire 7d"],
              ["+ Sign(uid) (token, error)",
               "+ Parse(token) (uid, error)"])

    # 关联
    arrow(ax, 5.1, 9.5, 5.6, 9.5, style="-")
    arrow(ax, 10.4, 9.5, 11.0, 9.5, style="-")
    arrow(ax, 13.3, 7.6, 13.3, 6.2, style="-")
    arrow(ax, 8.0, 7.6, 8.0, 6.2, style="-")
    arrow(ax, 2.8, 7.6, 2.8, 6.2, style="-")

    save(fig, "图4-8_用户认证模块类图.png")


# ---------------------------------------------------------------------------
# 图 4-9 单聊模块类图
# ---------------------------------------------------------------------------
def gen_fig_4_9():
    fig, ax = plt.subplots(figsize=(16, 11))
    ax.set_xlim(0, 16); ax.set_ylim(0, 11); ax.axis("off")
    class_box(ax, 0.3, 10.3, 4.8, "MessageHandler",
              ["- service: *MessageService"],
              ["+ Create(c *gin.Context)",
               "+ ListWithFriend(c *gin.Context)"])
    class_box(ax, 5.6, 10.3, 5.0, "MessageService",
              ["- repo: MessageRepository",
               "- friendSvc: *FriendService",
               "- hub: *Hub"],
              ["+ Create(senderID, req) (*Msg, error)",
               "+ ListBetween(uid, friendID) []Msg",
               "- validateAlgorithm(algo) bool"])
    class_box(ax, 11.1, 10.3, 4.7, "MessageRepository",
              ["- db: *gorm.DB"],
              ["+ Insert(m *Message) error",
               "+ ListByPair(a, b) ([]Message, error)"])

    class_box(ax, 0.3, 4.8, 4.8, "Message (model)",
              ["+ ID uint64", "+ SenderID/ReceiverID uint64",
               "+ SenderCiphertext string",
               "+ SenderAlgorithm string",
               "+ ReceiverCiphertext string",
               "+ ReceiverAlgorithm string",
               "+ CreatedAt time.Time"],
              [])

    class_box(ax, 5.6, 4.8, 5.0, "Hub (WebSocket)",
              ["- conns map[uint64]*Conn",
               "- mu sync.RWMutex"],
              ["+ Register(uid, conn)",
               "+ Unregister(uid)",
               "+ SendToUser(uid, payload)",
               "+ BroadcastToGroup(payloads)"])

    class_box(ax, 11.1, 4.8, 4.7, "Crypto (前端 TS)",
              [],
              ["+ generateRSAKeyPair()",
               "+ importPublicKey/PrivateKey()",
               "+ encryptRSA(text, pub)",
               "+ decryptRSA(ct, priv)",
               "+ generateAESKey()",
               "+ encryptAES / decryptAES",
               "+ wrap/unwrapAESKeyWithRSA"])

    arrow(ax, 5.1, 9.8, 5.6, 9.8, style="-")
    arrow(ax, 10.6, 9.8, 11.1, 9.8, style="-")
    arrow(ax, 8.1, 7.8, 8.1, 6.6, style="-")
    arrow(ax, 2.7, 7.8, 2.7, 6.6, style="-")

    save(fig, "图4-9_单聊模块类图.png")


# ---------------------------------------------------------------------------
# 图 4-10 群聊模块类图
# ---------------------------------------------------------------------------
def gen_fig_4_10():
    fig, ax = plt.subplots(figsize=(16, 16))
    ax.set_xlim(0, 16); ax.set_ylim(0, 16); ax.axis("off")
    # 第一行
    class_box(ax, 0.3, 15.3, 4.8, "GroupHandler",
              ["- service: *GroupService"],
              ["+ Create(c)", "+ ListMyGroups(c)",
               "+ GetDetail(c)", "+ Invite(c)",
               "+ Leave(c)"])
    class_box(ax, 5.6, 15.3, 5.0, "GroupService",
              ["- repo: GroupRepository",
               "- friendSvc: *FriendService"],
              ["+ Create(creator, name, ids)",
               "+ InviteMembers(gid, inviter, ids)",
               "+ Leave(gid, uid)",
               "+ GetDetail(gid, viewer)",
               "+ ListMyGroups(uid)"])
    class_box(ax, 11.1, 15.3, 4.7, "GroupRepository",
              ["- db: *gorm.DB"],
              ["+ CreateGroup(g, members) error",
               "+ FindByID(id) (*Group, error)",
               "+ ListMembers(gid) []Member",
               "+ ListGroupsOf(uid) []Group"])

    # 第二行
    class_box(ax, 0.3, 9.5, 4.8, "GroupMessageHandler",
              ["- service: *GroupMessageService"],
              ["+ Create(c)",
               "+ ListByGroup(c)"])
    class_box(ax, 5.6, 9.5, 5.0, "GroupMessageService",
              ["- repo: GroupMessageRepository",
               "- groupSvc: *GroupService",
               "- hub: *Hub"],
              ["+ Create(uid, gid, req)",
               "  1) 校验是群成员",
               "  2) 校验 memberKeys 覆盖",
               "  3) 校验算法白名单",
               "  4) 写 group_messages",
               "  5) 写 group_message_keys",
               "  6) Hub.BroadcastToGroup"])
    class_box(ax, 11.1, 9.5, 4.7, "GroupMessageRepository",
              ["- db: *gorm.DB"],
              ["+ InsertWithKeys(msg, keys) tx",
               "+ ListVisibleByUser(gid, uid)"])

    # 第三行（model）
    class_box(ax, 0.3, 3.8, 4.8, "ChatGroup",
              ["+ ID", "+ Name", "+ OwnerID",
               "+ CreatedAt"], [])
    class_box(ax, 5.6, 3.8, 5.0, "GroupMessage",
              ["+ ID", "+ GroupID",
               "+ SenderID",
               "+ ContentCiphertext",
               "+ ContentIv",
               "+ ContentAlgorithm",
               "+ CreatedAt"], [])
    class_box(ax, 11.1, 3.8, 4.7, "GroupMessageKey",
              ["+ ID", "+ MessageID",
               "+ UserID",
               "+ KeyCiphertext",
               "+ KeyAlgorithm"], [])

    # 关联
    arrow(ax, 5.1, 14.8, 5.6, 14.8, style="-")
    arrow(ax, 10.6, 14.8, 11.1, 14.8, style="-")
    arrow(ax, 5.1, 9.0, 5.6, 9.0, style="-")
    arrow(ax, 10.6, 9.0, 11.1, 9.0, style="-")

    save(fig, "图4-10_群聊模块类图.png")


# ---------------------------------------------------------------------------
# 图 4-11 朋友圈模块类图
# ---------------------------------------------------------------------------
def gen_fig_4_11():
    fig, ax = plt.subplots(figsize=(16, 14))
    ax.set_xlim(0, 16); ax.set_ylim(0, 14); ax.axis("off")
    # 第一行
    class_box(ax, 0.3, 13.3, 4.8, "MomentHandler",
              ["- service: *MomentService"],
              ["+ Create(c)", "+ ListTimeline(c)",
               "+ ListByUser(c)", "+ Delete(c)",
               "+ Like(c)", "+ Unlike(c)",
               "+ Comment(c)", "+ AIAssist(c)"])
    class_box(ax, 5.6, 13.3, 5.0, "MomentService",
              ["- repo: MomentRepository",
               "- friendSvc: *FriendService",
               "- storage: ObjectStorage",
               "- ai: *MomentAIProvider"],
              ["+ Create(uid, content, keys)",
               "+ ListTimeline(viewer)",
               "+ ListByUser(viewer, owner)",
               "+ Delete(uid, mid)",
               "+ Like / Unlike / Comment",
               "+ AIAssist(uid, req)"])
    class_box(ax, 11.1, 13.3, 4.7, "MomentRepository",
              ["- db: *gorm.DB"],
              ["+ Insert(m) error",
               "+ ListByUserIDs(ids)",
               "+ Like / Unlike",
               "+ AddComment",
               "+ Delete(mid)"])

    # 第二行
    class_box(ax, 0.3, 6.6, 4.8, "MomentAIProvider",
              ["- apiKey string", "- httpClient"],
              ["+ Assist(req) (text, error)",
               "  组装 system + user prompt",
               "  调用 DeepSeek Chat API",
               "  错误降级 (503/502)"])
    class_box(ax, 5.6, 6.6, 5.0, "UploadService",
              ["- storage: ObjectStorage"],
              ["+ Save(userID, file) (key, url)",
               "+ buildKey(userID, ext)"])
    class_box(ax, 11.1, 6.6, 4.7, "ObjectStorage «interface»",
              [],
              ["+ Put(key, reader) error",
               "+ PublicURL(key) string"])

    # 第三行
    class_box(ax, 2.0, 2.6, 5.0, "MinIOStorage",
              ["- client *minio.Client",
               "- bucket string"],
              ["+ Put / PublicURL"])
    class_box(ax, 9.0, 2.6, 5.0, "LocalStorage",
              ["- baseDir string",
               "- publicPrefix string"],
              ["+ Put / PublicURL"])

    arrow(ax, 5.1, 12.8, 5.6, 12.8, style="-")
    arrow(ax, 10.6, 12.8, 11.1, 12.8, style="-")
    arrow(ax, 10.6, 6.1, 11.1, 6.1, style="-")
    arrow(ax, 4.5, 3.5, 11.5, 5.6, style="-|>", color="#666")
    arrow(ax, 11.5, 3.5, 13.0, 5.6, style="-|>", color="#666")

    save(fig, "图4-11_朋友圈模块类图.png")


# ---------------------------------------------------------------------------
# 图 4-12 好友关系模块类图
# ---------------------------------------------------------------------------
def gen_fig_4_12():
    fig, ax = plt.subplots(figsize=(16, 11))
    ax.set_xlim(0, 16); ax.set_ylim(0, 11); ax.axis("off")
    class_box(ax, 0.3, 10.3, 4.8, "FriendHandler",
              ["- service: *FriendService"],
              ["+ SendRequest(c)",
               "+ ListIncoming(c)",
               "+ Accept(c)", "+ Reject(c)",
               "+ ListFriends(c)"])
    class_box(ax, 5.6, 10.3, 5.0, "FriendService",
              ["- reqRepo: FriendRequestRepository",
               "- frRepo: FriendRepository"],
              ["+ SendRequest(from, to, msg)",
               "  - 不能加自己 / 已是好友 / 已有 pending",
               "+ Accept(reqID, uid) tx",
               "  - 状态置 accepted",
               "  - 写双向 friends",
               "+ Reject(reqID, uid)",
               "+ ListFriends(uid) []FriendView",
               "+ AreFriends(a, b) bool"])
    class_box(ax, 11.1, 10.3, 4.7, "FriendRequestRepository",
              ["- db: *gorm.DB"],
              ["+ Insert(r) error",
               "+ UpdateStatus(id, status)",
               "+ ListIncoming(uid)",
               "+ ExistsPending(from, to) bool"])

    class_box(ax, 0.3, 3.6, 4.8, "FriendRepository",
              ["- db: *gorm.DB"],
              ["+ InsertPair(uid, fid) tx",
               "+ ListFriendIDs(uid)",
               "+ AreFriends(a, b) bool"])
    class_box(ax, 5.6, 3.6, 5.0, "FriendRequest (model)",
              ["+ ID", "+ FromUserID", "+ ToUserID",
               "+ Message string",
               "+ Status (pending/accepted/rejected)"], [])
    class_box(ax, 11.1, 3.6, 4.7, "Friend (model)",
              ["+ ID", "+ UserID", "+ FriendID",
               "  唯一索引 (user_id, friend_id)"], [])

    arrow(ax, 5.1, 9.8, 5.6, 9.8, style="-")
    arrow(ax, 10.6, 9.8, 11.1, 9.8, style="-")
    arrow(ax, 2.7, 6.0, 2.7, 5.0, style="-")
    arrow(ax, 8.1, 6.0, 8.1, 5.0, style="-")
    arrow(ax, 13.5, 6.0, 13.5, 5.0, style="-")

    save(fig, "图4-12_好友关系模块类图.png")


# ---------------------------------------------------------------------------
# 图 3-2 普通用户用例图
# ---------------------------------------------------------------------------
def gen_fig_3_2():
    fig, ax = plt.subplots(figsize=(14, 11))
    ax.set_xlim(0, 14); ax.set_ylim(0, 11); ax.axis("off")

    # actor — 火柴人
    ax_cx, ax_cy = 1.4, 5.5
    # 头
    head = Circle((ax_cx, ax_cy + 1.2), 0.28, fill=False, linewidth=1.6, edgecolor=C_LINE)
    ax.add_patch(head)
    # 身体
    ax.plot([ax_cx, ax_cx], [ax_cy + 0.9, ax_cy - 0.3], color=C_LINE, lw=1.6)
    # 手臂
    ax.plot([ax_cx - 0.6, ax_cx + 0.6], [ax_cy + 0.45, ax_cy + 0.45], color=C_LINE, lw=1.6)
    # 腿
    ax.plot([ax_cx, ax_cx - 0.5], [ax_cy - 0.3, ax_cy - 1.2], color=C_LINE, lw=1.6)
    ax.plot([ax_cx, ax_cx + 0.5], [ax_cy - 0.3, ax_cy - 1.2], color=C_LINE, lw=1.6)
    ax.text(ax_cx, ax_cy - 1.6, "普通用户", ha="center", va="center", fontsize=FS)

    # 系统边界
    sys_x, sys_y, sys_w, sys_h = 3.4, 0.4, 10.2, 10.2
    rect_sys = Rectangle((sys_x, sys_y), sys_w, sys_h, linewidth=1.4,
                         edgecolor=C_HEAD, facecolor="#FAFCFF")
    ax.add_patch(rect_sys)
    ax.text(sys_x + sys_w / 2, sys_y + sys_h - 0.25,
            "IM 即时通讯系统", ha="center", va="center",
            fontsize=FS, fontweight="bold", color=C_HEAD)

    # 用例(椭圆)
    use_cases = [
        # (cx, cy, label)
        (5.0, 9.4, "注册账号"),
        (8.2, 9.4, "登录系统"),
        (11.3, 9.4, "上传公钥"),
        (5.0, 8.4, "查看个人资料"),
        (8.2, 8.4, "修改个人资料"),
        (11.3, 8.4, "切换主页皮肤"),
        (5.0, 7.4, "发起好友申请"),
        (8.2, 7.4, "处理好友申请"),
        (11.3, 7.4, "查看好友列表"),
        (5.0, 6.4, "查看好友主页"),
        (8.2, 6.4, "发送单聊消息"),
        (11.3, 6.4, "查看单聊历史"),
        (5.0, 5.4, "创建群聊"),
        (8.2, 5.4, "邀请好友入群"),
        (11.3, 5.4, "退出群聊"),
        (5.0, 4.4, "发送群聊消息"),
        (8.2, 4.4, "查看群聊历史"),
        (11.3, 4.4, "上传朋友圈图片"),
        (5.0, 3.4, "发布朋友圈"),
        (8.2, 3.4, "浏览朋友圈时间线"),
        (11.3, 3.4, "点赞 / 取消点赞"),
        (5.0, 2.4, "评论朋友圈"),
        (8.2, 2.4, "删除朋友圈"),
        (11.3, 2.4, "调用 AI 文案助手"),
        (8.2, 1.3, "本地端到端加解密"),
    ]
    for cx, cy, label in use_cases:
        ellipse(ax, cx, cy, 2.6, 0.65, label, color=C_BLOCK, edge=C_HEAD, fontsize=FS)

    # actor → 关键用例 的关联线 (画几条代表性的关联)
    for cx, cy, _ in use_cases:
        ax.plot([ax_cx + 0.3, cx - 1.3], [ax_cy + 0.45, cy],
                color="#999", lw=0.6, alpha=0.6)

    save(fig, "图3-2_普通用户用例图.png")


if __name__ == "__main__":
    gen_fig_3_1()
    gen_fig_3_2()
    gen_fig_4_1()
    gen_fig_4_2()
    gen_fig_4_3()
    gen_fig_4_4()
    gen_fig_4_5()
    gen_fig_4_6()
    gen_fig_4_7()
    gen_fig_4_8()
    gen_fig_4_9()
    gen_fig_4_10()
    gen_fig_4_11()
    gen_fig_4_12()
    print("ALL DONE")
