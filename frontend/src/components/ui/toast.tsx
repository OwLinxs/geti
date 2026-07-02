import * as React from "react";
import { X, CheckCircle2, AlertCircle, Info } from "lucide-react";
import { cn } from "@/lib/utils";

type ToastVariant = "default" | "success" | "destructive";

interface ToastItem {
  id: number;
  titulo: string;
  descricao?: string;
  variant: ToastVariant;
}

interface ToastContextValue {
  toast: (t: {
    titulo: string;
    descricao?: string;
    variant?: ToastVariant;
  }) => void;
}

const ToastContext = React.createContext<ToastContextValue | null>(null);

export function useToast() {
  const ctx = React.useContext(ToastContext);
  if (!ctx) throw new Error("useToast deve ser usado dentro de ToastProvider");
  return ctx;
}

let idSeq = 0;

export function ToastProvider({ children }: { children: React.ReactNode }) {
  const [toasts, setToasts] = React.useState<ToastItem[]>([]);

  const remove = React.useCallback((id: number) => {
    setToasts((prev) => prev.filter((t) => t.id !== id));
  }, []);

  const toast = React.useCallback<ToastContextValue["toast"]>(
    ({ titulo, descricao, variant = "default" }) => {
      const id = ++idSeq;
      setToasts((prev) => [...prev, { id, titulo, descricao, variant }]);
      window.setTimeout(() => remove(id), 5000);
    },
    [remove]
  );

  return (
    <ToastContext.Provider value={{ toast }}>
      {children}
      <div className="fixed bottom-0 right-0 z-[100] flex max-h-screen w-full flex-col gap-2 p-4 sm:max-w-sm">
        {toasts.map((t) => (
          <ToastCard key={t.id} item={t} onClose={() => remove(t.id)} />
        ))}
      </div>
    </ToastContext.Provider>
  );
}

function ToastCard({
  item,
  onClose,
}: {
  item: ToastItem;
  onClose: () => void;
}) {
  const Icon =
    item.variant === "success"
      ? CheckCircle2
      : item.variant === "destructive"
        ? AlertCircle
        : Info;
  return (
    <div
      className={cn(
        "pointer-events-auto flex items-start gap-3 rounded-md border bg-card p-4 shadow-lg animate-in slide-in-from-right-4",
        item.variant === "destructive" && "border-destructive/40",
        item.variant === "success" && "border-emerald-300"
      )}
      role="status"
    >
      <Icon
        className={cn(
          "mt-0.5 h-5 w-5 shrink-0",
          item.variant === "success" && "text-emerald-600",
          item.variant === "destructive" && "text-destructive",
          item.variant === "default" && "text-primary"
        )}
      />
      <div className="flex-1">
        <p className="text-sm font-medium text-foreground">{item.titulo}</p>
        {item.descricao && (
          <p className="mt-0.5 text-sm text-muted-foreground">
            {item.descricao}
          </p>
        )}
      </div>
      <button
        onClick={onClose}
        className="text-muted-foreground transition-colors hover:text-foreground"
        aria-label="Fechar notificação"
      >
        <X className="h-4 w-4" />
      </button>
    </div>
  );
}
