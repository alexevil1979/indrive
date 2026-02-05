import { cn } from "@/lib/utils";

export function Badge({
  variant = "default",
  className,
  ...props
}: React.HTMLAttributes<HTMLSpanElement> & {
  variant?: "default" | "success" | "warning" | "error" | "outline";
}) {
  return (
    <span
      className={cn(
        "inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium",
        variant === "default" && "bg-muted text-muted-foreground",
        variant === "success" && "bg-green-100 text-green-800 dark:bg-green-900/30 dark:text-green-400",
        variant === "warning" && "bg-amber-100 text-amber-800 dark:bg-amber-900/30 dark:text-amber-400",
        variant === "error" && "bg-red-100 text-red-800 dark:bg-red-900/30 dark:text-red-400",
        variant === "outline" && "border border-border bg-transparent text-foreground",
        className
      )}
      {...props}
    />
  );
}
