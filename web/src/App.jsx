import { useCallback, useEffect, useRef, useState } from "react";
import { FilePlus2, FileText, FlaskConical, List, X } from "lucide-react";
import { toast } from "sonner";
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Kbd } from "@/components/ui/kbd";
import { Separator } from "@/components/ui/separator";
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip";
import { ModeToggle } from "@/components/mode-toggle";
import { useTheme } from "@/components/theme-provider";
import { cn } from "@/lib/utils";
import CaseListPage from "./CaseListPage.jsx";
import CaseForm from "./CaseForm.jsx";
import CaseDetailPage from "./CaseDetailPage.jsx";
import { fetchCases, fetchCourts } from "./api.js";
import { isPlainKey, isTyping } from "./keyboard.js";

// React 기초 용어(이 파일에서 자주 나옵니다):
//   - useState(초기값): 컴포넌트가 기억하는 값(상태)과 그 값을 바꾸는 함수를 [값, set함수]로 돌려줍니다.
//     set함수로 값을 바꾸면 React가 화면을 다시 그립니다. 변수에 직접 대입하면 화면이 바뀌지 않습니다.
//   - useEffect(함수, [의존값]): 화면을 그린 뒤 실행할 작업(서버 요청, 이벤트 등록 등)입니다.
//     의존값이 바뀔 때만 다시 실행되고, 돌려준 함수는 정리(clean-up) 때 실행됩니다.
//   - useCallback(함수, [의존값]): 의존값이 그대로면 같은 함수 객체를 재사용합니다(불필요한 재실행 방지).
//   - useRef(초기값): 값을 바꿔도 화면을 다시 그리지 않는 보관함입니다. .current로 읽고 씁니다.

// 항상 열려 있는 첫 탭(사건목록)의 id입니다.
const LIST_TAB = "cases";

// 탭 막대의 탭 하나. 브라우저·윈도우 탭처럼 위쪽 모서리가 둥근 상자로 그립니다.
//   - 선택된 탭: 테두리·위쪽 그림자·윗변 강조선을 두고, 바닥을 본문과 같은 색(bg-surface)으로 이어 붙입니다.
//   - 선택되지 않은 탭: 탭 막대에 묻히게 흐리게 두고, 마우스를 올리면 살짝 밝아집니다.
// 탭 버튼과 닫기 버튼은 나란히 둡니다(버튼 안에 버튼을 넣지 않기 위해).
function AppTab({ id, active, icon: Icon, label, dirty, onSelect, onClose }) {
  return (
    <div
      className={cn(
        "group relative flex h-8 max-w-60 min-w-32 items-center gap-1 rounded-t-lg border border-b-0 pr-1.5 pl-3 text-sm transition-colors",
        active
          ? [
              "z-10 -mb-px h-[calc(2rem+1px)] border-border bg-surface text-foreground",
              "shadow-[0_-3px_8px_-4px_rgb(0_0_0/0.18)]",
              "before:absolute before:inset-x-2 before:top-0 before:h-0.5 before:rounded-full before:bg-primary",
            ]
          : "border-transparent text-muted-foreground hover:bg-surface/70 hover:text-foreground",
      )}
    >
      <button
        type="button"
        role="tab"
        id={`tab-${id}`}
        aria-selected={active}
        aria-controls={`panel-${id}`}
        title={label}
        onClick={onSelect}
        className="flex h-full min-w-0 flex-1 items-center gap-1.5 font-medium outline-none focus-visible:text-primary"
      >
        <Icon className="size-4 shrink-0" />
        <span className="truncate">{label}</span>
        {dirty && (
          <span
            className="size-1.5 shrink-0 rounded-full bg-amber-500"
            title="저장하지 않은 내용이 있습니다"
            aria-label="저장 안 됨"
          />
        )}
      </button>
      {onClose && (
        <Button
          variant="ghost"
          size="icon-xs"
          aria-label={`${label} 탭 닫기`}
          className={cn(!active && "opacity-60 group-hover:opacity-100")}
          onClick={onClose}
        >
          <X />
        </Button>
      )}
    </div>
  );
}

// 앱 뼈대: 상단 메뉴줄과 그 아래 탭.
//   - 첫 탭 "사건목록"은 항상 있고 닫을 수 없습니다.
//   - 신건등록 탭은 필요할 때마다 새로 열고, 여러 개를 동시에 띄울 수 있습니다.
//   - 사건 탭은 사건 하나를 전체 화면으로 봅니다. 같은 사건은 탭 하나만 열고, 다시 열면 그 탭으로 옮깁니다.
// 다른 탭으로 옮겨도 탭 내용은 숨기기만 하므로(hidden) 입력 중인 내용과 스크롤 위치가 유지됩니다.
export default function App() {
  const { resolvedTheme } = useTheme(); // 실제로 적용 중인 테마("light" 또는 "dark")
  // 서버에서 받은 사건·법원 목록. null은 "아직 불러오는 중"을 뜻합니다.
  const [cases, setCases] = useState(null);
  const [casesError, setCasesError] = useState(null);
  const [courts, setCourts] = useState(null);
  const [courtsError, setCourtsError] = useState(null);

  // 사건목록 오른쪽에 붙는 탭들입니다. 열린 순서대로 탭 막대에 나옵니다.
  //   { id: "new-3", kind: "new", seq: 3 }          신건등록
  //   { id: "case-case-0012", kind: "case", item }   사건 보기(item은 사건 데이터)
  const [tabs, setTabs] = useState([]);
  const [tabStatus, setTabStatus] = useState({}); // id → { dirty, label }
  const [activeTab, setActiveTab] = useState(LIST_TAB);
  const [pendingClose, setPendingClose] = useState(null); // 닫기 확인 중인 탭 id
  const [openCase, setOpenCase] = useState(null); // 사건목록에서 미리보기로 열어 줄 사건(등록 직후 등)
  const seqRef = useRef(0); // 신건등록 탭 번호를 1, 2, 3...으로 매기는 계수기

  // 처음 화면이 그려질 때 한 번만([] = 의존값 없음) 사건과 법원 목록을 서버에서 받아 옵니다.
  // .then(성공 시 함수, 실패 시 함수) 형태로, 성공하면 목록을, 실패하면 오류 메시지를 상태에 넣습니다.
  useEffect(() => {
    fetchCases().then(setCases, (e) => setCasesError(e.message));
    fetchCourts().then(setCourts, (e) => setCourtsError(e.message));
  }, []);

  // 신건등록 탭을 하나 더 열고 그 탭으로 옮깁니다.
  const openNewCaseTab = useCallback(() => {
    const seq = ++seqRef.current;
    const id = `new-${seq}`;
    // set함수에 함수를 넘기면 "직전 값"을 받아 새 값을 만듭니다. [...list, 새것]은 기존 배열 뒤에 하나를 더한 새 배열입니다.
    setTabs((list) => [...list, { id, kind: "new", seq }]);
    setActiveTab(id);
  }, []);

  // 사건을 사건 탭으로 엽니다. 이미 열려 있으면 새로 만들지 않고 그 탭으로 옮깁니다.
  const openCaseTab = useCallback((item) => {
    const id = `case-${item.id}`;
    setTabs((list) => (list.some((t) => t.id === id) ? list : [...list, { id, kind: "case", item }]));
    setActiveTab(id);
  }, []);

  // 탭을 닫습니다. 저장하지 않은 내용이 있으면 force가 아닌 한 확인 대화상자를 먼저 띄웁니다.
  const closeTab = (id, force = false) => {
    if (!force && tabStatus[id]?.dirty) {
      setPendingClose(id);
      return;
    }
    const idx = tabs.findIndex((t) => t.id === id);
    const rest = tabs.filter((t) => t.id !== id);
    setTabs(rest);
    // 보고 있던 탭을 닫으면 바로 왼쪽 탭(없으면 사건목록)으로 옮깁니다.
    if (activeTab === id) setActiveTab(rest[idx - 1]?.id ?? LIST_TAB);
    // 객체에서 [id] 키만 빼고 나머지(...others)로 새 객체를 만듭니다. 닫은 탭의 상태 정보를 지우는 것입니다.
    setTabStatus(({ [id]: _, ...others }) => others);
  };

  // 등록 탭(CaseForm)이 "저장 안 됨 여부와 탭 이름"이 바뀔 때마다 알려 주는 함수입니다.
  const onStatus = useCallback((id, status) => setTabStatus((s) => ({ ...s, [id]: status })), []);

  // 사건 등록에 성공했을 때: 목록 맨 앞에 새 사건을 넣고, 등록 탭을 닫고,
  // 사건목록으로 돌아가 새 사건의 미리보기를 열고, 위쪽에 성공 알림을 띄웁니다.
  const onCreated = (id, created) => {
    setCases((list) => [created, ...(list ?? [])]);
    closeTab(id, true);
    setActiveTab(LIST_TAB);
    setOpenCase(created);
    toast.success(`${created.number} 사건을 등록했습니다.`);
  };

  // N 키로 새 사건 등록 탭을 엽니다. 한글 입력 상태에서도 동작하도록 글자가 아니라 키 위치(KeyN)로 판단합니다.
  useEffect(() => {
    const onKey = (e) => {
      if (e.code === "KeyN" && isPlainKey(e) && !e.shiftKey && !isTyping() && !pendingClose) {
        e.preventDefault();
        openNewCaseTab();
      }
    };
    window.addEventListener("keydown", onKey);
    // 정리 함수: 다음 실행 전이나 화면에서 사라질 때 등록한 이벤트를 떼어 내 중복 등록을 막습니다.
    return () => window.removeEventListener("keydown", onKey);
  }, [openNewCaseTab, pendingClose]);

  // 작성 중인 등록 탭이 있으면 브라우저 창을 닫거나 새로고침할 때 경고합니다.
  const anyDirty = Object.values(tabStatus).some((s) => s.dirty);
  useEffect(() => {
    if (!anyDirty) return;
    // beforeunload 이벤트에서 preventDefault를 부르면 브라우저가 "사이트를 나가시겠습니까?" 창을 띄웁니다.
    const onBeforeUnload = (e) => {
      e.preventDefault();
      e.returnValue = ""; // 옛 브라우저 호환용
    };
    window.addEventListener("beforeunload", onBeforeUnload);
    return () => window.removeEventListener("beforeunload", onBeforeUnload);
  }, [anyDirty]);

  const pendingLabel = pendingClose ? (tabStatus[pendingClose]?.label ?? "신건등록") : "";
  const labelOf = (seq) => (seq === 1 ? "신건등록" : `신건등록 (${seq})`);

  // 아래는 JSX입니다. HTML처럼 보이지만 JavaScript 안에서 화면 구조를 적는 문법이며,
  // {중괄호} 안에는 JavaScript 값을 넣을 수 있습니다. className은 HTML의 class(여기서는 Tailwind 클래스)입니다.
  return (
    <div className="flex h-svh flex-col bg-surface">
      <header className="flex h-14 shrink-0 items-center gap-3 border-b bg-background px-4">
        {/* 로고 파일은 저장소의 casepulse-logo/ 폴더에서 옵니다. 어두운 화면에서는 글자가 흰 로고를 씁니다. */}
        <img
          src={resolvedTheme === "dark" ? "/casepulse-logo-dark.svg" : "/casepulse-logo.svg"}
          alt="CasePulse"
          className="h-6 w-auto"
        />
        <Separator orientation="vertical" className="mx-1 h-5!" />
        <Tooltip>
          <TooltipTrigger asChild>
            <Button variant="ghost" onClick={openNewCaseTab}>
              <FilePlus2 /> 신건등록
            </Button>
          </TooltipTrigger>
          <TooltipContent>
            새 탭에서 신건등록 <Kbd>N</Kbd>
          </TooltipContent>
        </Tooltip>
        <div className="ml-auto flex items-center gap-2">
          <Badge variant="outline" className="gap-1 text-amber-700 dark:text-amber-300">
            <FlaskConical /> 더미 데이터
          </Badge>
          <ModeToggle />
        </div>
      </header>

      {/* 탭 막대: 본문보다 어두운 띠 위에 탭이 올라앉은 모양입니다. 선택된 탭이 아래 테두리를 덮어 본문과 이어집니다. */}
      <nav
        role="tablist"
        aria-label="화면"
        className="flex h-10 shrink-0 items-end gap-1 border-b bg-[color-mix(in_oklch,var(--muted),var(--foreground)_7%)] px-2 pt-2 dark:bg-background"
      >
        <AppTab
          id={LIST_TAB}
          icon={List}
          label="사건목록"
          active={activeTab === LIST_TAB}
          onSelect={() => setActiveTab(LIST_TAB)}
        />
        {/* 배열.map으로 탭마다 AppTab을 하나씩 만듭니다. key는 React가 항목을 구분하는 고유값입니다. */}
        {tabs.map((tab) => {
          const { id } = tab;
          const status =
            tab.kind === "new"
              ? (tabStatus[id] ?? { dirty: false, label: labelOf(tab.seq) })
              : { dirty: false, label: tab.item.number };
          return (
            <AppTab
              key={id}
              id={id}
              icon={tab.kind === "new" ? FilePlus2 : FileText}
              label={status.label}
              dirty={status.dirty}
              active={activeTab === id}
              onSelect={() => setActiveTab(id)}
              onClose={() => closeTab(id)}
            />
          );
        })}
      </nav>

      <main className="min-h-0 flex-1 bg-surface">
        <section
          role="tabpanel"
          id={`panel-${LIST_TAB}`}
          aria-labelledby={`tab-${LIST_TAB}`}
          hidden={activeTab !== LIST_TAB}
          className="h-full"
        >
          <CaseListPage
            cases={cases}
            error={casesError}
            active={activeTab === LIST_TAB}
            onNewCase={openNewCaseTab}
            openCase={openCase}
            onOpenCaseTab={openCaseTab}
          />
        </section>
        {tabs.map((tab) => (
          <section
            key={tab.id}
            role="tabpanel"
            id={`panel-${tab.id}`}
            aria-labelledby={`tab-${tab.id}`}
            hidden={activeTab !== tab.id}
            className="h-full"
          >
            {tab.kind === "new" ? (
              <CaseForm
                formId={tab.id}
                courts={courts}
                courtsError={courtsError}
                active={activeTab === tab.id}
                defaultLabel={labelOf(tab.seq)}
                onStatus={(s) => onStatus(tab.id, s)}
                onCreated={(created) => onCreated(tab.id, created)}
                onCancel={() => closeTab(tab.id)}
              />
            ) : (
              <CaseDetailPage item={tab.item} />
            )}
          </section>
        ))}
      </main>

      {/* 저장하지 않은 탭을 닫으려 할 때 뜨는 확인 대화상자입니다. !!값은 값을 true/false로 바꿉니다. */}
      <AlertDialog open={!!pendingClose} onOpenChange={(open) => !open && setPendingClose(null)}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>저장하지 않은 내용이 있습니다</AlertDialogTitle>
            <AlertDialogDescription>
              <b>{pendingLabel}</b> 탭을 닫으면 입력한 내용이 사라집니다.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel>계속 작성</AlertDialogCancel>
            <AlertDialogAction
              variant="destructive"
              onClick={() => {
                closeTab(pendingClose, true);
                setPendingClose(null);
              }}
            >
              닫기
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  );
}
