// 단축키 처리에 쓰는 도우미입니다.

// 사용자가 글자를 입력하는 중인지 알려 줍니다. 입력 중에는 단축키를 무시해야 합니다.
// 예: 검색창에 "n"을 입력할 때 신건등록 단축키(N)가 실행되면 안 됩니다.
export function isTyping() {
  const el = document.activeElement; // 지금 키보드 입력을 받고 있는(포커스된) 요소
  if (!el) return false;
  return el.isContentEditable || ["INPUT", "TEXTAREA", "SELECT"].includes(el.tagName);
}

// 보조 키 이름. 맥에서는 Cmd(⌘), 그 밖에는 Ctrl입니다.
// 맥에서 Ctrl+클릭은 오른쪽 클릭(메뉴)이므로, 클릭과 함께 쓰는 키는 맥에서 반드시 ⌘로 안내해야 합니다.
export const IS_MAC = /Mac|iPhone|iPad/.test(navigator.platform);
export const MOD_KEY = IS_MAC ? "⌘" : "Ctrl";

// Ctrl·Cmd·Alt 없이 누른 키인지 알려 줍니다.
export function isPlainKey(e) {
  return !e.ctrlKey && !e.metaKey && !e.altKey;
}
