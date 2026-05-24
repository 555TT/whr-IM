function pad(value) {
    return String(value).padStart(2, '0');
}
export function formatChatMessageTime(createdAt, now = new Date()) {
    if (!createdAt)
        return '';
    const date = new Date(createdAt);
    if (Number.isNaN(date.getTime()))
        return '';
    const sameDay = date.getFullYear() === now.getFullYear() &&
        date.getMonth() === now.getMonth() &&
        date.getDate() === now.getDate();
    const time = `${pad(date.getHours())}:${pad(date.getMinutes())}`;
    if (sameDay)
        return time;
    return `${pad(date.getMonth() + 1)}-${pad(date.getDate())} ${time}`;
}
