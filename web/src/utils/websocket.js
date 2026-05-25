const wsBaseURL = import.meta.env.VITE_WS_BASE_URL || '/ws';
function resolveWsUrl(base, token) {
    // 绝对地址（ws:// 或 wss://）直接拼上 token
    if (base.startsWith('ws://') || base.startsWith('wss://')) {
        return `${base}?token=${token}`;
    }
    // 相对路径：按当前页面协议构造，HTTPS 页面用 wss，HTTP 页面用 ws
    const proto = location.protocol === 'https:' ? 'wss:' : 'ws:';
    const path = base.startsWith('/') ? base : `/${base}`;
    return `${proto}//${location.host}${path}?token=${token}`;
}
export function createChatSocket(token) {
    return new WebSocket(resolveWsUrl(wsBaseURL, token));
}
