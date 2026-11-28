import { Badge } from "@/components/ui/badge";
import { cn } from "@/lib/utils";
import { CATEGORY_BY_VALUE } from "./cases.js";

// 사건 종류를 색 배지로 보여 줍니다.
// React 컴포넌트는 props(속성) 객체를 받아 화면에 그릴 JSX를 돌려주는 함수입니다.
// { category, className }은 props에서 두 값을 꺼내는 구조 분해 문법입니다.
export default function CategoryBadge({ category, className }) {
  const c = CATEGORY_BY_VALUE[category];
  // c?.badge의 ?.는 c가 없으면(알 수 없는 종류) 오류 대신 undefined를 돌려줍니다.
  // cn은 여러 Tailwind 클래스 문자열을 합치고, 서로 충돌하면 뒤의 것을 남기는 도우미입니다(lib/utils.js).
  return (
    <Badge variant="secondary" className={cn("rounded-md font-medium", c?.badge, className)}>
      {c?.label ?? category}
    </Badge>
  );
}
