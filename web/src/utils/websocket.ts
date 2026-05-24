const wsBaseURL = import.meta.env.VITE_WS_BASE_URL || 'ws://127.0.0.1:8080/ws'

export function createChatSocket(token: string) {
  return new WebSocket(`${wsBaseURL}?token=${token}`)
}
