// 단축키 처리에 쓰는 도우미입니다.

// 사용자가 글자를 입력하는 중인지 알려 줍니다. 입력 중에는 단축키를 무시해야 합니다.
export function isTyping() {
  const el = document.activeElement;
  if (!el) return false;
  return el.isContentEditable || ["INPUT", "TEXTAREA", "SELECT"].includes(el.tagName);
}

// Ctrl·Cmd·Alt 없이 누른 키인지 알려 줍니다.
export function isPlainKey(e) {
  return !e.ctrlKey && !e.metaKey && !e.altKey;
}
