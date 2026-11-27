// Go 서버 API 호출을 모아 둔 파일입니다.

async function getJSON(url) {
  const r = await fetch(url);
  if (!r.ok) throw new Error(`HTTP ${r.status}`);
  return r.json();
}

export const fetchCases = () => getJSON("/api/cases").then((b) => b.cases);
export const fetchCourts = () => getJSON("/api/courts").then((b) => b.courts);

// 사건을 등록합니다. 성공하면 { ok: true, case }, 검사에 걸리면 { ok: false, errors: [메시지...] }를 돌려줍니다.
export async function createCase(payload) {
  let r;
  try {
    r = await fetch("/api/cases", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(payload),
    });
  } catch (e) {
    return { ok: false, errors: [`서버에 연결하지 못했습니다: ${e.message}`] };
  }
  const body = await r.json().catch(() => ({}));
  if (r.ok) return { ok: true, case: body };
  return { ok: false, errors: body.errors ?? [`HTTP ${r.status}`] };
}
