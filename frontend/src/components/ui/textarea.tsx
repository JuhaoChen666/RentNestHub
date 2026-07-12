import * as React from "react";

import { cn } from "@/lib/utils";

function Textarea({ className, ...props }: React.ComponentProps<"textarea">) {
  return (
    <textarea
      data-slot="textarea"
      className={cn(
        "min-h-20 w-full rounded-md border border-[#cfd6d1] bg-white px-3 py-2 text-sm text-[var(--ink)] outline-none transition-shadow placeholder:text-[var(--muted)] focus:border-[var(--green)] focus:shadow-[0_0_0_3px_rgba(23,107,75,0.1)] disabled:cursor-not-allowed disabled:opacity-50",
        className,
      )}
      {...props}
    />
  );
}

export { Textarea };
