import { Label } from "@/components/ui/label";
import { cn } from "@/lib/utils";

interface FormFieldProps {
  label: string;
  htmlFor?: string;
  obrigatorio?: boolean;
  erro?: string;
  children: React.ReactNode;
  className?: string;
}

// Campo de formulário com rótulo e mensagem de erro em PT-BR.
export function FormField({
  label,
  htmlFor,
  obrigatorio,
  erro,
  children,
  className,
}: FormFieldProps) {
  return (
    <div className={cn("space-y-1.5", className)}>
      <Label htmlFor={htmlFor}>
        {label}
        {obrigatorio && <span className="ml-0.5 text-destructive">*</span>}
      </Label>
      {children}
      {erro && <p className="text-xs text-destructive">{erro}</p>}
    </div>
  );
}
