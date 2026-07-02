interface PageHeaderProps {
  titulo: string;
  descricao?: string;
  acao?: React.ReactNode;
}

export function PageHeader({ titulo, descricao, acao }: PageHeaderProps) {
  return (
    <div className="mb-6 flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
      <div>
        <h2 className="text-xl font-semibold tracking-tight text-foreground">
          {titulo}
        </h2>
        {descricao && (
          <p className="mt-1 text-sm text-muted-foreground">{descricao}</p>
        )}
      </div>
      {acao && <div className="flex shrink-0 gap-2">{acao}</div>}
    </div>
  );
}
