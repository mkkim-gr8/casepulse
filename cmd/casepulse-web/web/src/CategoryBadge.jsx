import { Badge } from "@/components/ui/badge";
import { cn } from "@/lib/utils";
import { CATEGORY_BY_VALUE } from "./cases.js";

// 사건 종류를 색 배지로 보여 줍니다.
export default function CategoryBadge({ category, className }) {
  const c = CATEGORY_BY_VALUE[category];
  return (
    <Badge variant="secondary" className={cn("rounded-md font-medium", c?.badge, className)}>
      {c?.label ?? category}
    </Badge>
  );
}
