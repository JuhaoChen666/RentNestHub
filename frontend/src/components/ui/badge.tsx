import * as React from "react";

import { cn } from "@/lib/utils";

function Badge({ className, ...props }: React.ComponentProps<"span">) {
  return (
    <span
      data-slot="badge"
      className={cn(
        "inline-flex items-center rounded px-2 py-1 text-xs font-semibold",
        className,
      )}
      {...props}
    />
  );
}

export { Badge };
