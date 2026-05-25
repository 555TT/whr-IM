"""Cover + 任务书 + 摘要 + ABSTRACT + 目录"""
from docx.enum.text import WD_ALIGN_PARAGRAPH
from docx.shared import Pt
from _gen_paper import (
    add_para,
    add_heading,
    add_centered,
    add_page_break,
    add_new_section,
    add_blank_line,
    PAPER_TITLE,
    set_run_font,
)


def _cover_line(doc, label, value='', underline_value=False):
    p = doc.add_paragraph()
    p.paragraph_format.line_spacing = 1.5
    p.paragraph_format.first_line_indent = Pt(32)
    p.paragraph_format.right_indent = Pt(10)
    r1 = p.add_run(f'{label}    ')
    set_run_font(r1, cn='宋体', size=16)
    r2 = p.add_run(value)
    set_run_font(r2, cn='黑体' if underline_value else '宋体', size=16)
    if underline_value:
        r2.font.underline = True
    # 补一段空白，使视觉效果更像模板
    r3 = p.add_run(' ' * 46)
    set_run_font(r3, cn='宋体', size=16)


def write_front(doc):
    # ========== 封面 ==========
    add_blank_line(doc)
    add_blank_line(doc)
    add_centered(doc, '郑州轻工业大学', size=22, bold=True, cn='宋体')
    add_centered(doc, '本科毕业设计(论文)', size=36, bold=False, cn='宋体')
    add_blank_line(doc)
    add_blank_line(doc)
    add_blank_line(doc)
    add_blank_line(doc)
    add_blank_line(doc)

    _cover_line(doc, '题    目', PAPER_TITLE, underline_value=True)
    _cover_line(doc, '学生姓名', '王浩然')
    _cover_line(doc, '专业班级', '软件工程22-0X')
    _cover_line(doc, '学    号', '5421220XXXX')
    _cover_line(doc, '学    院', '软件学院')
    _cover_line(doc, '指导教师（职称）', '***（***）')
    _cover_line(doc, '完成时间', '2026年5月18日')

    add_page_break(doc)

    # ========== 任务书 ==========
    add_blank_line(doc)
    add_centered(doc, '郑州轻工业大学', size=16, bold=False, cn='宋体')
    add_centered(doc, '毕业设计（论文）任务书', size=15, bold=True, cn='宋体')
    add_blank_line(doc)
    add_para(doc, f'题目：{PAPER_TITLE}', size=14, bold=True, indent_first=False, line_spacing=1.0)
    add_para(doc, '专业：软件工程22-0X    学号：5421220XXXX    姓名：王浩然', size=14, bold=True, indent_first=False, line_spacing=1.0)
    add_para(doc, '主要内容：', size=14, bold=True, indent_first=False, line_spacing=1.0)
    add_para(doc, '本课题设计并实现一套基于前后端分离架构的即时通讯系统，前端采用 Vue 3 + Vite + TypeScript 构建，'
                  '后端采用 Go 语言 + Gin 框架 + GORM + MySQL 实现。系统应支持用户注册登录、个人资料维护、'
                  '好友申请与管理、基于 WebSocket 的实时聊天以及历史消息查询等核心功能。'
                  '在此基础上，重点研究并实现基于 RSA-OAEP 非对称加密算法的端到端消息加密机制，'
                  '使聊天消息在用户浏览器中完成加密后再传至服务端，服务端只保存与转发密文，从而提升即时通讯系统中消息内容的安全性。',
             indent_first=True, line_spacing=1.5)
    add_para(doc, '基本要求：', size=14, bold=True, indent_first=False, line_spacing=1.0)
    add_para(doc, '（1）熟悉前后端分离的开发模式，能够独立完成前端工程与后端工程的搭建与联调；', line_spacing=1.5)
    add_para(doc, '（2）熟练掌握 WebSocket 协议与 JWT 鉴权机制，实现安全可靠的实时消息推送；', line_spacing=1.5)
    add_para(doc, '（3）熟悉 Web Crypto API 与 RSA-OAEP 算法原理，完成消息加密与解密功能；', line_spacing=1.5)
    add_para(doc, '（4）完成系统的需求分析、概要设计、详细设计、编码实现与测试，撰写完整的毕业论文。', line_spacing=1.5)
    add_para(doc, '思政要求：', size=14, bold=True, indent_first=False, line_spacing=1.0)
    add_para(doc, '在课题完成过程中，培养严谨求实的科学态度、自主学习与创新精神；通过研究消息加密技术，增强网络空间安全意识与社会责任感，树立正确的网络伦理观念。', line_spacing=1.5)
    add_para(doc, '主要参考资料：', size=14, bold=True, indent_first=False, line_spacing=1.0)
    add_para(doc, '[1] Tanenbaum A S, Wetherall D J. Computer Networks[M]. 5th ed. Boston: Pearson, 2011.', indent_first=False, hanging=True, line_spacing=1.5)
    add_para(doc, '[2] Donovan A A A, Kernighan B W. The Go Programming Language[M]. Boston: Addison-Wesley, 2015.', indent_first=False, hanging=True, line_spacing=1.5)
    add_para(doc, '[3] You E. Vue.js 3 Documentation[EB/OL]. (2024-01-01)[2026-03-01]. https://vuejs.org/.', indent_first=False, hanging=True, line_spacing=1.5)
    add_blank_line(doc)
    add_para(doc, '完  成  期  限：  2026年5月18日', size=14, bold=True, indent_first=False)
    add_para(doc, '指导教师签名：', size=14, bold=True, indent_first=False)
    add_para(doc, '专业负责人签名：', size=14, bold=True, indent_first=False)
    add_centered(doc, '2026年1月6日', size=14, bold=True, cn='宋体')

    # ========== 节2：摘要/ABSTRACT/目录 — 页眉=题目，页码=罗马数字 ==========
    add_new_section(doc, page_num_fmt='upperRoman', start=1,
                    header_text=PAPER_TITLE, show_page_number=True)

    # ========== 中文摘要 ==========
    add_centered(doc, '摘  要', size=15, bold=False, cn='宋体')
    add_blank_line(doc)
    add_para(doc, '即时通讯系统已成为互联网应用中的基础设施，但传统方案大多停留在“传输层加密 + 服务端明文存储”这一层面。'
                  '该模式能够保护消息在网络链路中的安全，却无法消除服务端集中持有明文所带来的泄露风险。'
                  '围绕这一问题，本文设计并实现了一套基于 Gin 和 WebSocket 的 IM 即时通讯系统，'
                  '在完成用户认证、好友关系维护、实时聊天、历史消息查询等基础功能的同时，'
                  '将研究重点放在端到端消息加密机制的工程落地上。')
    add_para(doc, '系统前端采用 Vue 3、Vite、TypeScript 与 Pinia 构建，后端采用 Go 1.22、Gin、GORM 与 MySQL 实现，'
                  '并通过 Gorilla WebSocket 建立长连接以承担实时消息下行。在安全实现上，'
                  '本文采用 RSA-OAEP 非对称加密算法与浏览器原生 Web Crypto API，由前端在本地生成密钥对，'
                  '私钥仅保存在浏览器环境中，公钥上传服务端。发送消息时，前端分别使用发送方公钥和接收方公钥'
                  '对同一条明文加密，形成两份独立密文；服务端仅负责密文的校验、存储与转发，'
                  '从而在不破坏聊天历史可读性的前提下，将消息明文排除在服务端存储边界之外。')
    add_para(doc, '测试结果表明，系统在注册登录、好友管理、实时聊天、历史消息查询、消息加密等核心场景下'
                  '均能稳定运行，12 项功能测试全部通过。对网络请求与数据库内容的观测显示，服务端链路中未出现'
                  '明文消息泄露。本文实现的方案具备端到端加密的基本特征，能够满足毕业设计场景下对工程完整性、'
                  '通信实时性与消息安全性的综合要求。与此同时，文中也对公钥真实性校验、前向保密、会话密钥协商'
                  '与多设备同步等尚未覆盖的问题进行了边界分析，为后续向工业级 E2EE 体系演进提供了清晰方向。')
    add_blank_line(doc)
    add_para(doc, '关键词：即时通讯；前后端分离；WebSocket；端到端加密；RSA-OAEP',
             indent_first=False, bold=False, line_spacing=1.5)

    add_blank_line(doc)
    add_blank_line(doc)
    add_blank_line(doc)

    # ========== ABSTRACT ==========
    add_centered(doc, 'ABSTRACT', size=14, bold=True, cn='Times New Roman')
    add_blank_line(doc)
    add_blank_line(doc)
    add_para(doc, 'Instant messaging (IM) has become one of the most important application forms in the Internet era, '
                  'and is widely used in social communication, office collaboration and online cooperation scenarios. '
                  'As users pay more and more attention to the privacy of their messages, how to provide effective '
                  'content encryption protection in a general-purpose IM system has become an important research and '
                  'engineering topic. This paper designs and implements an instant messaging system based on Gin and '
                  'WebSocket, and focuses on the engineering implementation of end-to-end message encryption while '
                  'retaining basic functions such as authentication, friend management, real-time chat and historical '
                  'message query.', en='Times New Roman', cn='Times New Roman')
    add_para(doc, 'The front-end of the system is built on Vue 3, Vite, TypeScript and Pinia, while the back-end is '
                  'implemented with Go 1.22, Gin, GORM and MySQL. Real-time message delivery is supported by Gorilla '
                  'WebSocket. In the security design, RSA-OAEP and the browser-native Web Crypto API are adopted. Key '
                  'pairs are generated locally in the browser, private keys are kept on the client side only, and public '
                  'keys are uploaded to the server. For each outgoing message, the plaintext is encrypted twice with the '
                  'sender\'s public key and the receiver\'s public key respectively, so that the server stores and forwards '
                  'ciphertext only while both communication parties can decrypt their own copies locally.', en='Times New Roman', cn='Times New Roman')
    add_para(doc, 'Experimental and functional verification results show that all 12 core test cases pass successfully, '
                  'and no plaintext message is exposed in the server-side storage or network payloads. The proposed '
                  'scheme exhibits the basic characteristics of end-to-end encryption and is suitable for an undergraduate '
                  'engineering project. At the same time, the limitations in public key authenticity verification, forward '
                  'secrecy, session key negotiation and multi-device synchronization are explicitly discussed, providing a '
                  'clear direction for future evolution toward a full industrial-grade E2EE system.', en='Times New Roman', cn='Times New Roman')
    add_blank_line(doc)
    add_para(doc, 'KEY WORDS: Instant Messaging; Front-end and Back-end Separation; WebSocket; End-to-End Encryption; RSA-OAEP',
             indent_first=False, bold=True, line_spacing=1.5, en='Times New Roman', cn='Times New Roman')

    # ========== 目录 ==========
    add_new_section(doc, page_num_fmt='upperRoman', start=1,
                    header_text=PAPER_TITLE, show_page_number=True)
    add_centered(doc, '目  录', size=16, bold=False, cn='宋体')
    add_blank_line(doc)
    toc_items = [
        ('摘  要', 'I'),
        ('ABSTRACT', 'II'),
        ('1  绪论', '1'),
        ('    1.1  研究背景与意义', '1'),
        ('    1.2  国内外研究现状', '2'),
        ('        1.2.1  国外研究现状', '2'),
        ('        1.2.2  国内研究现状', '3'),
        ('    1.3  论文主要工作', '3'),
        ('    1.4  论文组织结构', '4'),
        ('2  相关技术介绍', '5'),
        ('    2.1  前后端分离架构', '5'),
        ('    2.2  Vue 3 与前端技术栈', '5'),
        ('    2.3  Go 语言与 Gin 框架', '6'),
        ('    2.4  WebSocket 协议', '6'),
        ('    2.5  JWT 身份认证', '7'),
        ('    2.6  RSA-OAEP 公钥加密与 Web Crypto API', '7'),
        ('    2.7  MySQL 数据库与 GORM', '8'),
        ('3  系统需求分析与总体设计', '9'),
        ('    3.1  需求分析', '9'),
        ('        3.1.1  功能性需求', '9'),
        ('        3.1.2  非功能性需求', '10'),
        ('        3.1.3  用例分析', '10'),
        ('    3.2  系统总体架构', '11'),
        ('    3.3  系统功能模块设计', '12'),
        ('    3.4  数据库设计', '13'),
        ('4  关键技术实现', '15'),
        ('    4.1  用户认证与 JWT 实现', '15'),
        ('    4.2  好友关系处理', '16'),
        ('    4.3  基于 WebSocket 的实时聊天实现', '17'),
        ('        4.3.1  WebSocket 连接建立与鉴权', '17'),
        ('        4.3.2  连接管理与消息广播', '18'),
        ('        4.3.3  消息收发完整时序', '19'),
        ('    4.4  端到端消息加密机制', '20'),
        ('        4.4.1  设计目标与总体思路', '20'),
        ('        4.4.2  RSA-OAEP 密钥对生成与管理', '21'),
        ('        4.4.3  公钥上传与好友公钥获取', '22'),
        ('        4.4.4  双重密文加密与服务端密文转发', '22'),
        ('        4.4.5  接收端解密与历史消息展示', '23'),
        ('        4.4.6  安全性分析', '24'),
        ('5  系统实现与测试', '25'),
        ('    5.1  开发与运行环境', '25'),
        ('    5.2  用户认证模块实现', '25'),
        ('    5.3  个人资料模块实现', '26'),
        ('    5.4  好友管理模块实现', '27'),
        ('    5.5  聊天模块实现', '28'),
        ('    5.6  系统功能测试', '29'),
        ('    5.7  系统性能与安全性测试', '30'),
        ('    5.8  与同类系统的对比', '31'),
        ('结束语', '33'),
        ('致  谢', '34'),
        ('参考文献', '35'),
    ]
    for t, page in toc_items:
        p = doc.add_paragraph(style='Normal')
        p.paragraph_format.line_spacing = 1.5
        r = p.add_run(f'{t}')
        set_run_font(r, cn='宋体', en='Times New Roman', size=12)
        r2 = p.add_run('\t' + page)
        set_run_font(r2, cn='宋体', en='Times New Roman', size=12)

    # ========== 节3：正文（第1章起）— 页眉=题目，页码=阿拉伯数字从1重启 ==========
    add_new_section(doc, page_num_fmt='decimal', start=1,
                    header_text=PAPER_TITLE, show_page_number=True)
