import { useCallback, useEffect, useRef, useState } from "react";
import { FilePlus2, FlaskConical, List, Plus, X } from "lucide-react";
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
import { fetchCases, fetchCourts } from "./api.js";
import { isPlainKey, isTyping } from "./keyboard.js";

const LIST_TAB = "cases";

// 탭 막대의 탭 하나. 탭 버튼과 닫기 버튼을 나란히 둡니다(버튼 안에 버튼을 넣지 않기 위해).
function AppTab({ id, active, icon: Icon, label, dirty, onSelect, onClose }) {
  return (
    <div
      className={cn(
        "group relative flex h-9 items-center gap-1 border-b-2 text-sm transition-colors",
        active ? "border-primary text-foreground" : "border-transparent text-muted-foreground hover:text-foreground",
      )}
    >
      <button
        type="button"
        role="tab"
        id={`tab-${id}`}
        aria-selected={active}
        aria-controls={`panel-${id}`}
        onClick={onSelect}
        className="flex h-full items-center gap-1.5 px-1 font-medium outline-none focus-visible:text-primary"
      >
        <Icon className="size-4" />
        {label}
        {dirty && (
          <span className="size-1.5 rounded-full bg-amber-500" title="저장하지 않은 내용이 있습니다" aria-label="저장 안 됨" />
        )}
      </button>
      {onClose && (
        <Button variant="ghost" size="icon-xs" aria-label={`${label} 탭 닫기`} onClick={onClose}>
          <X />
        </Button>
      )}
    </div>
  );
}

// 앱 뼈대: 상단 메뉴줄과 그 아래 탭.
//   - 첫 탭 "사건목록"은 항상 있고 닫을 수 없습니다.
//   - 사건 등록 탭은 필요할 때마다 새로 열고, 여러 개를 동시에 띄울 수 있습니다.
// 다른 탭으로 옮겨도 탭 내용은 숨기기만 하므로(hidden) 입력 중인 내용과 스크롤 위치가 유지됩니다.
export default function App() {
  const { resolvedTheme } = useTheme();
  const [cases, setCases] = useState(null);
  const [casesError, setCasesError] = useState(null);
  const [courts, setCourts] = useState(null);
  const [courtsError, setCourtsError] = useState(null);

  const [formTabs, setFormTabs] = useState([]); // [{ id, seq }]
  const [tabStatus, setTabStatus] = useState({}); // id → { dirty, label }
  const [activeTab, setActiveTab] = useState(LIST_TAB);
  const [pendingClose, setPendingClose] = useState(null); // 닫기 확인 중인 탭 id
  const [openCase, setOpenCase] = useState(null);
  const seqRef = useRef(0);

  useEffect(() => {
    fetchCases().then(setCases, (e) => setCasesError(e.message));
    fetchCourts().then(setCourts, (e) => setCourtsError(e.message));
  }, []);

  const openNewCaseTab = useCallback(() => {
    const seq = ++seqRef.current;
    const id = `new-${seq}`;
    setFormTabs((tabs) => [...tabs, { id, seq }]);
    setActiveTab(id);
  }, []);

  // 탭을 닫습니다. 저장하지 않은 내용이 있으면 force가 아닌 한 확인 대화상자를 먼저 띄웁니다.
  const closeTab = (id, force = false) => {
    if (!force && tabStatus[id]?.dirty) {
      setPendingClose(id);
      return;
    }
    const idx = formTabs.findIndex((t) => t.id === id);
    const rest = formTabs.filter((t) => t.id !== id);
    setFormTabs(rest);
    // 보고 있던 탭을 닫으면 바로 왼쪽 탭(없으면 사건목록)으로 옮깁니다.
    if (activeTab === id) setActiveTab(rest[idx - 1]?.id ?? LIST_TAB);
    setTabStatus(({ [id]: _, ...others }) => others);
  };

  const onStatus = useCallback((id, status) => setTabStatus((s) => ({ ...s, [id]: status })), []);

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
    return () => window.removeEventListener("keydown", onKey);
  }, [openNewCaseTab, pendingClose]);

  // 작성 중인 등록 탭이 있으면 브라우저 창을 닫거나 새로고침할 때 경고합니다.
  const anyDirty = Object.values(tabStatus).some((s) => s.dirty);
  useEffect(() => {
    if (!anyDirty) return;
    const onBeforeUnload = (e) => {
      e.preventDefault();
      e.returnValue = "";
    };
    window.addEventListener("beforeunload", onBeforeUnload);
    return () => window.removeEventListener("beforeunload", onBeforeUnload);
  }, [anyDirty]);

  const pendingLabel = pendingClose ? (tabStatus[pendingClose]?.label ?? "새 사건") : "";
  const labelOf = (seq) => (seq === 1 ? "새 사건" : `새 사건 (${seq})`);

  return (
    <div className="flex h-svh flex-col bg-muted/40">
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
              <Plus /> 새 사건
            </Button>
          </TooltipTrigger>
          <TooltipContent>
            새 탭에서 사건 등록 <Kbd>N</Kbd>
          </TooltipContent>
        </Tooltip>
        <div className="ml-auto flex items-center gap-2">
          <Badge variant="outline" className="gap-1 text-amber-700 dark:text-amber-300">
            <FlaskConical /> 더미 데이터
          </Badge>
          <ModeToggle />
        </div>
      </header>

      <nav role="tablist" aria-label="화면" className="flex h-10 shrink-0 items-end gap-4 border-b bg-background px-4">
        <AppTab
          id={LIST_TAB}
          icon={List}
          label="사건목록"
          active={activeTab === LIST_TAB}
          onSelect={() => setActiveTab(LIST_TAB)}
        />
        {formTabs.map(({ id, seq }) => {
          const status = tabStatus[id] ?? { dirty: false, label: labelOf(seq) };
          return (
            <AppTab
              key={id}
              id={id}
              icon={FilePlus2}
              label={status.label}
              dirty={status.dirty}
              active={activeTab === id}
              onSelect={() => setActiveTab(id)}
              onClose={() => closeTab(id)}
            />
          );
        })}
      </nav>

      <main className="min-h-0 flex-1">
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
          />
        </section>
        {formTabs.map(({ id, seq }) => (
          <section
            key={id}
            role="tabpanel"
            id={`panel-${id}`}
            aria-labelledby={`tab-${id}`}
            hidden={activeTab !== id}
            className="h-full"
          >
            <CaseForm
              formId={id}
              courts={courts}
              courtsError={courtsError}
              active={activeTab === id}
              defaultLabel={labelOf(seq)}
              onStatus={(s) => onStatus(id, s)}
              onCreated={(created) => onCreated(id, created)}
              onCancel={() => closeTab(id)}
            />
          </section>
        ))}
      </main>

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
