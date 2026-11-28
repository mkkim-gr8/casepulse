// Go 서버(Aegis) API 호출을 모아 둔 파일입니다.
// 화면 컴포넌트는 fetch를 직접 쓰지 않고 이 파일의 함수만 불러 쓰므로, API 주소가 바뀌어도 여기만 고치면 됩니다.

// getJSON은 url에 GET 요청을 보내고 응답 본문을 JSON으로 읽어 돌려줍니다.
// async 함수는 Promise를 돌려주며, 안에서 await를 쓰면 응답이 올 때까지 기다렸다가 다음 줄을 실행합니다.
async function getJSON(url) {
  const r = await fetch(url); // fetch: 브라우저 내장 HTTP 요청 함수
  // r.ok는 상태 코드가 200~299일 때 true입니다. 실패하면 오류를 던져 호출한 쪽의 catch로 넘깁니다.
  if (!r.ok) throw new Error(`HTTP ${r.status}`);
  return r.json();
}

// 응답 {"cases": [...], "total": 240}에서 cases 배열만 꺼내 돌려줍니다.
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
    // 서버가 꺼져 있는 등 네트워크 자체가 실패한 경우입니다.
    return { ok: false, errors: [`서버에 연결하지 못했습니다: ${e.message}`] };
  }
  // 응답이 JSON이 아니면(예: 프록시 오류 페이지) 빈 객체로 대신합니다.
  const body = await r.json().catch(() => ({}));
  if (r.ok) return { ok: true, case: body };
  // ??는 왼쪽 값이 null/undefined일 때만 오른쪽 값을 씁니다.
  return { ok: false, errors: body.errors ?? [`HTTP ${r.status}`] };
}
