import { useEffect, useMemo, useRef, useState } from "react";
import { Building2, CircleAlert, FileText, Plus, Save, Trash2, TriangleAlert, Users } from "lucide-react";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import { Button } from "@/components/ui/button";
import { Card, CardAction, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Checkbox } from "@/components/ui/checkbox";
import { Input } from "@/components/ui/input";
import { Kbd, KbdGroup } from "@/components/ui/kbd";
import { Label } from "@/components/ui/label";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Spinner } from "@/components/ui/spinner";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip";
import { Combobox } from "@/components/combobox";
import { cn } from "@/lib/utils";
import { createCase } from "./api.js";
import { CATEGORIES, clientSides } from "./cases.js";
import { MOD_KEY } from "./keyboard.js";

// 사건 종류별로 지위 입력칸에 띄울 추천 값입니다. 목록에 없는 지위도 자유롭게 입력할 수 있습니다.
const ROLE_SUGGESTIONS = {
  civil: ["원고", "피고", "원고(반소피고)", "피고(반소원고)", "독립당사자참가인", "보조참가인"],
  civil_application: ["채권자", "채무자", "신청인", "피신청인", "제3채무자"],
  family: ["원고", "피고", "청구인", "상대방", "사건본인"],
  administrative: ["원고", "피고", "피고보조참가인"],
  criminal: ["피고인", "검사"],
};

// "2026가합12345"처럼 한 줄로 쓴 사건번호를 나눕니다. 형식이 아니면 null입니다.
// 정규식 풀이: (\d{4}) 숫자 4개 = 연도, ([가-힣]+) 한글 1자 이상 = 사건부호, (\d+) 숫자 1개 이상 = 일련번호.
// \s*는 사이사이 공백을 허용합니다. 괄호로 묶은 부분은 match 결과의 m[1], m[2], m[3]으로 꺼낼 수 있습니다.
const NUMBER_PATTERN = /^\s*(\d{4})\s*([가-힣]+)\s*(\d+)\s*$/;

// 당사자 행마다 붙이는 고유 번호입니다. 행을 지우거나 추가해도 React가 각 행을 정확히 구분하도록 key로 씁니다.
// (배열 위치를 key로 쓰면 중간 행을 지울 때 입력값이 엉뚱한 행으로 옮겨 보일 수 있습니다.)
let partyKeySeq = 0;
const newParty = () => ({ key: ++partyKeySeq, role: "", name: "", client: false });

// initialForm은 새 등록 탭의 빈 입력값입니다. 연도는 올해로, 당사자는 빈 행 두 개(보통 원고·피고)로 시작합니다.
// 입력칸의 값은 숫자도 문자열로 보관하고, 서버로 보낼 때 Number()로 바꿉니다.
function initialForm() {
  return {
    year: String(new Date().getFullYear()),
    code: "",
    serial: "",
    category: "",
    title: "",
    parties: [newParty(), newParty()],
    courtId: "",
    panel: "",
  };
}

// 비교용 값: 당사자 key처럼 화면에서만 쓰는 값은 빼고, 입력한 내용만 남깁니다.
// 처음 값의 snapshot과 지금 값의 snapshot이 다르면 "저장하지 않은 내용이 있음(dirty)"으로 봅니다.
const snapshot = (f) => JSON.stringify({ ...f, parties: f.parties.map(({ role, name, client }) => ({ role, name, client })) });

// Clio 서버 검사와 같은 규칙을 화면에서 먼저 확인합니다. 입력칸별 오류 메시지를 돌려줍니다.
// 화면 검사는 사용자 편의를 위한 것이고, 최종 판단은 항상 서버(Clio)가 합니다.
// 반환값: { errors: 항목별 오류 메시지, hasError: 오류가 하나라도 있는지 }
function validate(f) {
  const e = {};
  // /^\d{4}$/.test(값): 값이 처음(^)부터 끝($)까지 숫자 4개인지 확인합니다.
  if (!/^\d{4}$/.test(f.year)) e.year = "연도는 4자리 숫자입니다";
  if (!/^[가-힣]+$/.test(f.code.trim())) e.code = "사건부호는 한글로 입력합니다";
  if (!/^[1-9]\d*$/.test(f.serial)) e.serial = "일련번호는 1 이상의 숫자입니다";
  if (!f.category) e.category = "사건 종류를 고르세요";
  if (!f.title.trim()) e.title = "사건명을 입력하세요";
  e.partyRows = f.parties.map((p) => ({
    role: p.role.trim() ? null : "지위를 입력하세요",
    name: p.name.trim() ? null : "이름을 입력하세요",
  }));
  // 지위가 다른 당사자를 함께 의뢰인으로 하면 이해충돌입니다. 입력하는 즉시 알리도록 conflict에 따로 둡니다.
  const sides = clientSides(f.parties);
  if (sides.length > 1) e.conflict = `지위가 다른 당사자(${sides.join(", ")})를 함께 의뢰인으로 할 수 없습니다(이해충돌). 한쪽 의뢰인만 체크하세요.`;
  if (f.parties.length === 0) e.parties = "당사자를 한 명 이상 입력하세요";
  else if (!f.parties.some((p) => p.client)) e.parties = "의뢰인인 당사자를 한 명 이상 표시하세요";
  if (!f.courtId) e.court = "계속 법원을 고르세요";
  if (!f.panel) e.panel = "재판부를 고르세요";
  // partyRows는 항상 들어 있으므로 따로 확인하고, 나머지 키는 하나라도 있으면 오류입니다.
  const hasError =
    Object.keys(e).some((k) => k !== "partyRows") || e.partyRows.some((r) => r.role || r.name);
  return { errors: e, hasError };
}

// 입력칸 하나: 이름표, 입력 요소, 안내 또는 오류 문구.
function Field({ label, htmlFor, required, error, hint, className, children }) {
  return (
    <div className={cn("flex flex-col gap-1.5", className)}>
      <Label htmlFor={htmlFor}>
        {label}
        {required && <span className="text-xs font-normal text-muted-foreground">(필수)</span>}
      </Label>
      {children}
      {(error || hint) && (
        <p className={cn("text-xs", error ? "text-destructive" : "text-muted-foreground")}>{error || hint}</p>
      )}
    </div>
  );
}

// 사건 등록 탭의 내용입니다.
//   - onStatus({ dirty, label }): 저장하지 않은 내용이 있는지와 탭 제목을 알립니다.
//   - onCreated(사건): 등록에 성공했을 때 호출됩니다.
//   - onCancel(): 취소 버튼을 눌렀을 때 호출됩니다(저장하지 않은 내용 확인은 탭을 닫는 쪽에서 합니다).
export default function CaseForm({ formId, courts, courtsError, active, defaultLabel, onStatus, onCreated, onCancel }) {
  const [form, setForm] = useState(initialForm); // 지금 입력한 값 전체
  const [initial] = useState(() => snapshot(form)); // 처음 값(바뀌었는지 비교용, 다시 바꾸지 않음)
  const [submitted, setSubmitted] = useState(false); // 등록 버튼을 한 번이라도 눌렀는지
  const [serverErrors, setServerErrors] = useState([]); // 서버가 거절한 이유 목록
  const [saving, setSaving] = useState(false); // 서버에 보내는 중인지(버튼 잠금용)
  const yearRef = useRef(null); // 탭을 열면 연도 칸에 바로 커서를 두기 위한 참조

  // set({ title: "사기" })처럼 바꿀 칸만 넘기면 나머지 값은 그대로 두고 그 칸만 바꿉니다.
  const set = (patch) => setForm((f) => ({ ...f, ...patch }));
  // setParty(key, { name: "홍길동" }): key가 같은 당사자 행만 바꾸고 나머지 행은 그대로 둡니다.
  const setParty = (key, patch) =>
    setForm((f) => ({ ...f, parties: f.parties.map((p) => (p.key === key ? { ...p, ...patch } : p)) }));

  // 입력할 때마다(화면을 다시 그릴 때마다) 검사를 다시 계산합니다.
  const { errors, hasError } = validate(form);
  // 입력하자마자 빨간 오류가 뜨면 불편하므로, 일반 오류는 등록 버튼을 누른 뒤부터 보여 줍니다.
  // (이해충돌 경고 errors.conflict만은 show를 거치지 않고 즉시 보여 줍니다.)
  const show = (msg) => (submitted ? msg : null);
  const dirty = snapshot(form) !== initial;
  // 사건번호 세 칸이 모두 올바르면 "2026가합12345"를 만들어 탭 제목으로 씁니다.
  const numberText =
    /^\d{4}$/.test(form.year) && form.code.trim() && /^[1-9]\d*$/.test(form.serial)
      ? `${form.year}${form.code.trim()}${form.serial}`
      : "";
  const label = numberText || defaultLabel;

  // dirty나 탭 제목이 바뀔 때마다 App에 알립니다(탭의 주황 점·제목·닫기 확인에 씀).
  useEffect(() => onStatus({ dirty, label }), [dirty, label]);
  // 이 탭이 보이게 되면 연도 칸에 커서를 둡니다.
  useEffect(() => {
    if (active) yearRef.current?.focus();
  }, [active]);

  const selectedCourt = useMemo(() => courts?.find((c) => c.id === form.courtId) ?? null, [courts, form.courtId]);
  const clientCount = form.parties.filter((p) => p.client).length;
  // 등록 탭이 여러 개 동시에 열릴 수 있으므로, HTML id가 겹치지 않게 탭 id(formId)를 앞에 붙입니다.
  const roleListId = `role-options-${formId}`;
  const id = (name) => `${formId}-${name}`;

  // 사건번호 칸 어디에든 "2026가합12345"를 붙여 넣으면 세 칸으로 나눠 넣습니다.
  const onNumberPaste = (e) => {
    const m = e.clipboardData.getData("text").match(NUMBER_PATTERN);
    if (!m) return; // 형식이 아니면 평소처럼 붙여 넣습니다.
    e.preventDefault(); // 브라우저의 기본 붙여넣기를 막고 직접 세 칸에 나눠 넣습니다.
    set({ year: m[1], code: m[2], serial: m[3] });
  };

  // 등록: 화면 검사를 통과하면 서버에 보내고, 성공하면 onCreated, 실패하면 서버 오류를 보여 줍니다.
  const submit = async () => {
    setSubmitted(true); // 이제부터 입력칸 오류를 보여 줍니다.
    setServerErrors([]);
    if (hasError) return;
    setSaving(true);
    // 문자열로 보관한 숫자는 Number()로 바꾸고, 앞뒤 공백을 지워 서버 API 형식(createCaseRequest)에 맞춥니다.
    const result = await createCase({
      number: { year: Number(form.year), code: form.code.trim(), serial: Number(form.serial) },
      title: form.title.trim(),
      category: form.category,
      parties: form.parties.map(({ role, name, client }) => ({ role: role.trim(), name: name.trim(), client })),
      court: { id: form.courtId, panel: form.panel },
    });
    setSaving(false);
    if (result.ok) onCreated(result.case);
    else setServerErrors(result.errors);
  };

  // Ctrl+Enter(맥은 Cmd+Enter)로 등록합니다. 법원·재판부를 고른 직후처럼 입력칸에 포커스가 없어도
  // 이 탭이 보이는 동안에는 동작하도록 창 전체에서 키를 받습니다.
  //
  // submitRef를 쓰는 이유: 이벤트 등록(useEffect)은 active가 바뀔 때만 다시 하므로, submit을 직접 쓰면
  // 처음 등록할 때의 오래된 submit(옛 입력값)을 부르게 됩니다. 매번 ref에 최신 submit을 넣어 두고 그것을 부릅니다.
  const submitRef = useRef(submit);
  submitRef.current = submit;
  useEffect(() => {
    if (!active) return;
    const onKey = (e) => {
      if (e.key === "Enter" && (e.ctrlKey || e.metaKey)) {
        e.preventDefault();
        submitRef.current();
      }
    };
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, [active]);

  // 사건번호 세 칸의 오류 중 첫 번째 것 하나만 칸 아래에 보여 줍니다.
  const numberError = show(errors.year || errors.code || errors.serial);

  return (
    <div className="flex h-full flex-col">
      <div className="min-h-0 flex-1 overflow-y-auto p-4">
        <div className="mx-auto flex max-w-4xl flex-col gap-4">
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <FileText className="size-4" /> 사건 정보
              </CardTitle>
            </CardHeader>
            <CardContent className="grid gap-4 sm:grid-cols-2">
              <Field
                label="사건번호"
                htmlFor={id("year")}
                required
                error={numberError}
                hint="2026가합12345처럼 한 줄로 붙여 넣으면 세 칸에 나뉘어 들어갑니다"
              >
                <div className="grid grid-cols-[5rem_1fr_1fr] gap-2">
                  <Input
                    id={id("year")}
                    ref={yearRef}
                    placeholder="연도"
                    inputMode="numeric"
                    maxLength={4}
                    value={form.year}
                    aria-invalid={!!show(errors.year) || undefined}
                    // /\D/g는 숫자가 아닌 모든 글자입니다. 숫자 외의 입력은 지워서 숫자만 남깁니다.
                    onChange={(e) => set({ year: e.target.value.replace(/\D/g, "") })}
                    onPaste={onNumberPaste}
                  />
                  <Input
                    aria-label="사건부호"
                    placeholder="사건부호 (예: 가합)"
                    value={form.code}
                    aria-invalid={!!show(errors.code) || undefined}
                    onChange={(e) => set({ code: e.target.value })}
                    onPaste={onNumberPaste}
                  />
                  <Input
                    aria-label="일련번호"
                    placeholder="일련번호"
                    inputMode="numeric"
                    value={form.serial}
                    aria-invalid={!!show(errors.serial) || undefined}
                    onChange={(e) => set({ serial: e.target.value.replace(/\D/g, "") })}
                    onPaste={onNumberPaste}
                  />
                </div>
              </Field>
              <Field label="사건 종류" htmlFor={id("category")} required error={show(errors.category)}>
                <Select value={form.category} onValueChange={(v) => set({ category: v })}>
                  <SelectTrigger id={id("category")} className="w-full" aria-invalid={!!show(errors.category) || undefined}>
                    <SelectValue placeholder="선택하세요" />
                  </SelectTrigger>
                  <SelectContent>
                    {CATEGORIES.map((c) => (
                      <SelectItem key={c.value} value={c.value}>
                        {c.label}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </Field>
              <Field label="사건명" htmlFor={id("title")} required error={show(errors.title)} className="sm:col-span-2">
                <Input
                  id={id("title")}
                  placeholder="예: 손해배상(기), 사기"
                  value={form.title}
                  aria-invalid={!!show(errors.title) || undefined}
                  onChange={(e) => set({ title: e.target.value })}
                />
              </Field>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Users className="size-4" /> 당사자
              </CardTitle>
              <CardAction className="text-sm text-muted-foreground">
                {form.parties.length}명 · 의뢰인 {clientCount}명
              </CardAction>
            </CardHeader>
            <CardContent className="flex flex-col gap-2 px-0">
              <Table>
                <TableHeader>
                  <TableRow>
                    <TableHead className="w-10 pl-6 text-center">#</TableHead>
                    <TableHead className="w-[32%]">지위</TableHead>
                    <TableHead>이름</TableHead>
                    <TableHead className="w-16 text-center">의뢰인</TableHead>
                    <TableHead className="w-12 pr-6" />
                  </TableRow>
                </TableHeader>
                <TableBody>
                  {form.parties.map((p, i) => {
                    const rowErr = errors.partyRows[i];
                    return (
                      <TableRow
                        key={p.key}
                        className={cn(
                          p.client && "bg-primary/5 hover:bg-primary/10",
                          p.client && errors.conflict && "bg-destructive/10 hover:bg-destructive/15",
                        )}
                      >
                        <TableCell className="pl-6 text-center text-muted-foreground tabular-nums">{i + 1}</TableCell>
                        <TableCell>
                          <Input
                            list={roleListId}
                            aria-label={`${i + 1}번째 당사자 지위`}
                            placeholder="예: 원고, 피고1"
                            value={p.role}
                            aria-invalid={!!show(rowErr.role) || undefined}
                            onChange={(e) => setParty(p.key, { role: e.target.value })}
                          />
                        </TableCell>
                        <TableCell>
                          <Input
                            aria-label={`${i + 1}번째 당사자 이름`}
                            placeholder="이름 또는 상호"
                            value={p.name}
                            aria-invalid={!!show(rowErr.name) || undefined}
                            onChange={(e) => setParty(p.key, { name: e.target.value })}
                          />
                        </TableCell>
                        <TableCell className="text-center">
                          <Checkbox
                            checked={p.client}
                            aria-invalid={(p.client && !!errors.conflict) || undefined}
                            aria-label={`${i + 1}번째 당사자 의뢰인 여부`}
                            onCheckedChange={(v) => setParty(p.key, { client: v === true })}
                          />
                        </TableCell>
                        <TableCell className="pr-6">
                          <Tooltip>
                            <TooltipTrigger asChild>
                              <Button
                                variant="ghost"
                                size="icon-sm"
                                aria-label="당사자 삭제"
                                disabled={form.parties.length === 1}
                                onClick={() => set({ parties: form.parties.filter((x) => x.key !== p.key) })}
                              >
                                <Trash2 />
                              </Button>
                            </TooltipTrigger>
                            <TooltipContent side="left">이 당사자 삭제</TooltipContent>
                          </Tooltip>
                        </TableCell>
                      </TableRow>
                    );
                  })}
                </TableBody>
              </Table>
              {/* datalist는 입력칸(list 속성으로 연결)에 추천 값을 보여 주는 HTML 기능입니다. 사건 종류에 맞는 지위를 추천합니다. */}
              <datalist id={roleListId}>
                {(ROLE_SUGGESTIONS[form.category] ?? ROLE_SUGGESTIONS.civil).map((r) => (
                  <option key={r} value={r} />
                ))}
              </datalist>
              <div className="flex items-center gap-3 px-6">
                <Button variant="outline" size="sm" onClick={() => set({ parties: [...form.parties, newParty()] })}>
                  <Plus /> 당사자 추가
                </Button>
                <p
                  className={cn(
                    "text-xs",
                    errors.conflict || show(errors.parties) ? "text-destructive" : "text-muted-foreground",
                  )}
                >
                  {errors.conflict ??
                    show(errors.parties) ??
                    "의뢰인인 당사자에 체크하세요. 의뢰인은 같은 지위의 당사자만 될 수 있고, 같은 지위라도 의뢰인이 아닌 사람은 체크하지 않습니다."}
                </p>
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Building2 className="size-4" /> 계속 법원
              </CardTitle>
            </CardHeader>
            <CardContent className="flex flex-col gap-4">
              {courtsError && (
                <Alert variant="destructive">
                  <CircleAlert />
                  <AlertTitle>법원 목록을 불러오지 못했습니다</AlertTitle>
                  <AlertDescription>{courtsError}</AlertDescription>
                </Alert>
              )}
              <div className="grid gap-4 sm:grid-cols-2">
                <Field
                  label="법원"
                  required
                  error={show(errors.court)}
                  hint="재판부 정보가 있는 법원만 고를 수 있습니다"
                >
                  <Combobox
                    items={courts ?? []}
                    value={form.courtId}
                    // 법원을 바꾸면 이전 법원의 재판부는 맞지 않으므로 재판부 선택을 비웁니다.
                    onChange={(courtId) => set({ courtId, panel: "" })}
                    getKey={(c) => c.id}
                    getLabel={(c) => c.name}
                    getHint={(c) => `재판부 ${c.panels.length}`}
                    placeholder={courts ? "법원을 고르세요" : "법원 목록을 불러오는 중…"}
                    searchPlaceholder="법원 이름으로 검색"
                    disabled={!courts}
                    invalid={!!show(errors.court)}
                  />
                </Field>
                <Field
                  label="재판부"
                  required
                  error={show(errors.panel)}
                  hint={selectedCourt ? null : "법원을 먼저 고르세요"}
                >
                  {/* key를 법원 ID로 주면 법원이 바뀔 때 재판부 선택 상자를 새로 만들어 검색어 등이 초기화됩니다. */}
                  <Combobox
                    key={form.courtId}
                    items={selectedCourt?.panels ?? []}
                    value={form.panel}
                    onChange={(panel) => set({ panel })}
                    placeholder="재판부를 고르세요"
                    searchPlaceholder="재판부 이름으로 검색"
                    disabled={!selectedCourt}
                    invalid={!!show(errors.panel)}
                  />
                </Field>
              </div>
            </CardContent>
          </Card>
        </div>
      </div>

      <footer className="shrink-0 border-t bg-background px-4 py-3">
        <div className="mx-auto flex max-w-4xl flex-col gap-3">
          {serverErrors.length > 0 && (
            <Alert variant="destructive">
              <CircleAlert />
              <AlertTitle>등록하지 못했습니다</AlertTitle>
              <AlertDescription>
                <ul className="list-disc pl-4">
                  {serverErrors.map((m, i) => (
                    <li key={i}>{m}</li>
                  ))}
                </ul>
              </AlertDescription>
            </Alert>
          )}
          {submitted && hasError && serverErrors.length === 0 && (
            <Alert className="border-amber-300 text-amber-800 dark:border-amber-500/40 dark:text-amber-300">
              <TriangleAlert />
              <AlertTitle>빨간색으로 표시한 항목을 확인하세요.</AlertTitle>
            </Alert>
          )}
          <div className="flex items-center justify-end gap-2">
            <span className="mr-auto flex items-center gap-1 text-xs text-muted-foreground">
              <KbdGroup>
                <Kbd>{MOD_KEY}</Kbd>
                <Kbd>Enter</Kbd>
              </KbdGroup>
              로 등록
            </span>
            <Button variant="outline" onClick={onCancel}>
              취소
            </Button>
            <Button onClick={submit} disabled={saving}>
              {saving ? <Spinner /> : <Save />} 등록
            </Button>
          </div>
        </div>
      </footer>
    </div>
  );
}
