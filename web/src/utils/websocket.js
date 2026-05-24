export function createChatSocket(token) {
    return new WebSocket(`ws://127.0.0.1:8080/ws?token=${token}`);
}
