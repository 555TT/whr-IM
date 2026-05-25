"""Cover + 任务书 + 摘要 + ABSTRACT + 目录"""
from docx.enum.text import WD_ALIGN_PARAGRAPH
from docx.shared import Pt
from _gen_paper import (add_para, add_heading, add_centered, add_page_break,
                        add_new_section, PAPER_TITLE)


def write_front(doc):
    # ========== 封面 ==========
    for _ in range(3):
        doc.add_paragraph()
    add_centered(doc, '郑州轻工业大学', size=22, bold=True, space_after=18)
    add_centered(doc, '本科毕业设计（论文）', size=20, bold=True, space_after=60)
    for _ in range(3):
        doc.add_paragraph()
    items = [
        ('题    目', '基于Gin和WebSocket的IM即时通讯系统'),
        ('学生姓名', '王浩然'),
        ('专业班级', '软件工程22-0X'),
        ('学    号', '5421220XXXX'),
        ('学    院', '软件学院'),
        ('指导教师（职称）', '***（***）'),
        ('完成时间', '2026年5月18日'),
    ]
    for k, v in items:
        p = doc.add_paragraph()
        p.alignment = WD_ALIGN_PARAGRAPH.CENTER
        p.paragraph_format.space_after = Pt(10)
        from _gen_paper import set_run_font
        r1 = p.add_run(f'{k}    ')
        set_run_font(r1, size=14, bold=True)
        r2 = p.add_run(v)
        set_run_font(r2, size=14, bold=False)
        r3 = p.add_run('    ')
        set_run_font(r3, size=14)
    add_page_break(doc)

    # ========== 任务书 ==========
    add_centered(doc, '郑州轻工业大学', size=18, bold=True, space_after=10)
    add_centered(doc, '毕业设计（论文）任务书', size=18, bold=True, space_after=20)
    add_para(doc, '题目：基于Gin和WebSocket的IM即时通讯系统', indent_first=False, bold=True)
    add_para(doc, '专业：软件工程    学号：5421220XXXX    姓名：王浩然', indent_first=False)
    add_heading(doc, '主要内容：', level=3)
    add_para(doc, '本课题设计并实现一套基于前后端分离架构的即时通讯系统，前端采用 Vue 3 + Vite + TypeScript 构建，'
                  '后端采用 Go 语言 + Gin 框架 + GORM + MySQL 实现。系统应支持用户注册登录、个人资料维护、'
                  '好友申请与管理、基于 WebSocket 的实时聊天以及历史消息查询等核心功能。'
                  '在此基础上，重点研究并实现基于 RSA-OAEP 非对称加密算法的端到端消息加密机制，'
                  '使聊天消息在用户浏览器中完成加密后再传至服务端，服务端只保存与转发密文，'
                  '从而提升即时通讯系统中消息内容的安全性。')
    add_heading(doc, '基本要求：', level=3)
    add_para(doc, '（1）熟悉前后端分离的开发模式，能够独立完成前端工程与后端工程的搭建与联调；')
    add_para(doc, '（2）熟练掌握 WebSocket 协议与 JWT 鉴权机制，实现安全可靠的实时消息推送；')
    add_para(doc, '（3）熟悉 Web Crypto API 与 RSA-OAEP 算法原理，完成消息加密与解密功能；')
    add_para(doc, '（4）完成系统的需求分析、概要设计、详细设计、编码实现与测试，撰写完整的毕业论文。')
    add_heading(doc, '思政要求：', level=3)
    add_para(doc, '在课题完成过程中，培养严谨求实的科学态度、自主学习与创新精神；'
                  '通过研究消息加密技术，增强网络空间安全意识与社会责任感，树立正确的网络伦理观念。')
    add_heading(doc, '主要参考资料：', level=3)
    add_para(doc, '[1] Tanenbaum A S, Wetherall D J. Computer Networks[M]. 5th ed. Boston: Pearson, 2011.', indent_first=False)
    add_para(doc, '[2] Donovan A A A, Kernighan B W. The Go Programming Language[M]. Boston: Addison-Wesley, 2015.', indent_first=False)
    add_para(doc, '[3] You E. Vue.js 3 Documentation[EB/OL]. (2024-01-01)[2026-03-01]. https://vuejs.org/.', indent_first=False)
    add_para(doc, '完  成  期  限：2026年5月18日', indent_first=False, space_before=12)
    add_para(doc, '指导教师签名：', indent_first=False)
    add_para(doc, '专业负责人签名：', indent_first=False)
    add_para(doc, '2026年1月6日', indent_first=False)

    # ========== 节2：摘要/ABSTRACT/目录 — 页眉=题目，页码=罗马数字 ==========
    add_new_section(doc, page_num_fmt='upperRoman', start=1,
                    header_text=PAPER_TITLE, show_page_number=True)

    # ========== 中文摘要 ==========
    add_centered(doc, '摘  要', size=16, bold=True, space_after=12)
    add_para(doc, '即时通讯几乎已经渗透到当代互联网用户的全部沟通场景，社交、办公、'
                  '远程协作皆离不开它。问题在于，消息一旦到达服务端往往以明文形式落库，'
                  '一旦数据库被攻破，后果难以挽回。本文围绕这一痛点，'
                  '设计并实现了一个前后端分离的即时通讯系统，'
                  '并把工作重点放在端到端消息加密上：'
                  '采用 RSA-OAEP 非对称算法，让消息在用户浏览器中完成加密再上传，'
                  '服务端只触碰密文。')
    add_para(doc, '系统前端使用 Vue 3 + Vite + TypeScript，状态管理交给 Pinia，'
                  'HTTP 请求由 Axios 统一封装，密钥生成与消息加解密则借助浏览器原生的 Web Crypto API；'
                  '后端选用 Go 1.22 + Gin + GORM + MySQL 这一组合，'
                  'WebSocket 部分基于 Gorilla 库实现长连接，承担实时消息的推送。'
                  '身份认证统一走 JWT。'
                  '为了让发送方在自己的设备上也能回看历史聊天，'
                  '系统对同一条明文做两次加密——分别使用收发双方的公钥——'
                  '这样数据库里只有密文，双方却各自持有可解密的副本。')
    add_para(doc, '功能测试覆盖了注册登录、好友管理、实时聊天、历史消息查询、'
                  '消息加密等所有核心场景，共 12 例，全部通过；'
                  '抓包与数据库直查亦未发现明文外泄，'
                  '表明系统达到了预期的安全目标，具有工程实践价值。')
    add_para(doc, '关键词：即时通讯；前后端分离；WebSocket；端到端加密；RSA-OAEP',
             indent_first=False, bold=True, space_before=12)
    add_page_break(doc)

    # ========== ABSTRACT ==========
    add_centered(doc, 'ABSTRACT', size=16, bold=True, space_after=12)
    add_para(doc, 'Instant messaging (IM) has become one of the most important application forms in the Internet era, '
                  'and is widely used in social communication, office collaboration and online cooperation scenarios. '
                  'As users pay more and more attention to the privacy of their messages, '
                  'how to provide effective content encryption protection in a general-purpose IM system '
                  'has become an important research and engineering topic. This paper designs and implements '
                  'an instant messaging system based on a front-end and back-end separated architecture. '
                  'On the basis of basic business functions, this paper focuses on the design and implementation '
                  'of an end-to-end message encryption mechanism based on the RSA-OAEP asymmetric encryption algorithm, '
                  'so that messages are encrypted in the user browser before being submitted to the server, '
                  'and the server only participates in the storage and forwarding of ciphertext.',
             en='Times New Roman', cn='Times New Roman')
    add_para(doc, 'The front-end of the system is built on Vue 3, Vite and TypeScript, '
                  'using Pinia for state management, interacting with the back-end HTTP interface via Axios, '
                  'and using the browser Web Crypto API to generate key pairs and encrypt or decrypt messages. '
                  'The back-end is implemented based on Go 1.22, Gin, GORM and MySQL, '
                  'and uses the Gorilla WebSocket library to build long connections '
                  'for real-time message push and forwarding. The system implements identity authentication through JWT, '
                  'and encrypts one copy of ciphertext for the sender and the receiver respectively, '
                  'so that both chat parties can decrypt and view historical messages locally. '
                  'Functional test results show that the core functions such as registration and login, '
                  'friend management, real-time chat, historical message query and message encryption '
                  'can all run stably and meet the design objectives of the system, '
                  'showing good engineering practical value.',
             en='Times New Roman', cn='Times New Roman')
    add_para(doc, 'KEY WORDS: Instant Messaging; Front-end and Back-end Separation; WebSocket; '
                  'End-to-End Encryption; RSA-OAEP',
             indent_first=False, bold=True, space_before=12, en='Times New Roman', cn='Times New Roman')
    add_page_break(doc)

    # ========== 目录 ==========
    add_centered(doc, '目  录', size=16, bold=True, space_after=12)
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
        ('结束语', '32'),
        ('致  谢', '33'),
        ('参考文献', '34'),
    ]
    for t, page in toc_items:
        p = doc.add_paragraph()
        p.paragraph_format.line_spacing = 1.5
        from _gen_paper import set_run_font
        r = p.add_run(f'{t}')
        set_run_font(r, size=12)
        r2 = p.add_run('   ' + '·' * max(2, 60 - len(t) * 2) + '   ' + page)
        set_run_font(r2, size=12)

    # ========== 节3：正文（第1章起）— 页眉=题目，页码=阿拉伯数字从1重启 ==========
    add_new_section(doc, page_num_fmt='decimal', start=1,
                    header_text=PAPER_TITLE, show_page_number=True)
