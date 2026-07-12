import * as React from "react";

import { cn } from "@/lib/utils";

function NativeSelect({
  className,
  ...props
}: React.ComponentProps<"select">) {
  return (
    <select
      data-slot="native-select"
      className={cn(
        "h-10 w-full appearance-none rounded-md bg-transparent text-sm text-[var(--ink)] outline-none",
        className,
      )}
      {...props}
    />
  );
}

export { NativeSelect };
