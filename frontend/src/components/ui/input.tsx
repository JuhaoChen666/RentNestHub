import * as React from "react";

import { cn } from "@/lib/utils";

function Input({ className, type, ...props }: React.ComponentProps<"input">) {
  return (
    <input
      data-slot="input"
      type={type}
      className={cn(
        "flex h-10 w-full rounded-md border border-[#cfd6d1] bg-white px-3 py-2 text-sm text-[var(--ink)] outline-none transition-shadow file:border-0 file:bg-transparent file:text-sm file:font-medium placeholder:text-[var(--muted)] focus:border-[var(--green)] focus:shadow-[0_0_0_3px_rgba(23,107,75,0.1)] disabled:cursor-not-allowed disabled:opacity-50",
        className,
      )}
      {...props}
    />
  );
}

export { Input };
