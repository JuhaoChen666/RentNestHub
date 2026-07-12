import * as React from "react";
import { Slot } from "@radix-ui/react-slot";
import { cva, type VariantProps } from "class-variance-authority";

import { cn } from "@/lib/utils";

const buttonVariants = cva(
  "inline-flex items-center justify-center gap-2 whitespace-nowrap rounded-md text-sm font-medium transition-colors disabled:pointer-events-none disabled:opacity-50 [&_svg]:pointer-events-none [&_svg]:size-4 [&_svg]:shrink-0",
  {
    variants: {
      variant: {
        default: "bg-[var(--green)] text-white hover:bg-[var(--green-dark)]",
        secondary:
          "bg-[var(--green-soft)] text-[var(--green)] hover:bg-[#d8ece2]",
        outline:
          "border border-[#b9d2c4] bg-transparent text-[var(--green)] hover:bg-[var(--green-soft)]",
        ghost: "text-[var(--green)] hover:bg-[var(--green-soft)]",
      },
      size: {
        default: "h-10 px-4 py-2",
        icon: "h-9 w-9",
        sm: "h-8 px-3",
      },
    },
    defaultVariants: {
      variant: "default",
      size: "default",
    },
  },
);

function Button({
  className,
  variant,
  size,
  asChild = false,
  ...props
}: React.ComponentProps<"button"> &
  VariantProps<typeof buttonVariants> & {
    asChild?: boolean;
  }) {
  const Comp = asChild ? Slot : "button";

  return (
    <Comp
      data-slot="button"
      className={cn(buttonVariants({ variant, size, className }))}
      {...props}
    />
  );
}

export { Button };
