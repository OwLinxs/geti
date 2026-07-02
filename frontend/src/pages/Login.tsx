import * as React from "react";
import { useLocation, useNavigate } from "react-router-dom";
import { Package } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { FormField } from "@/components/FormField";
import { Spinner } from "@/components/ui/spinner";
import { useAuth } from "@/contexts/AuthContext";
import { mensagemErro } from "@/services/api/client";

export default function Login() {
  const { entrar, autenticado } = useAuth();
  const navigate = useNavigate();
  const location = useLocation();
  const destino =
    (location.state as { from?: { pathname?: string } } | null)?.from
      ?.pathname ?? "/";

  const [email, setEmail] = React.useState("");
  const [senha, setSenha] = React.useState("");
  const [erro, setErro] = React.useState<string | null>(null);
  const [enviando, setEnviando] = React.useState(false);

  React.useEffect(() => {
    if (autenticado) navigate(destino, { replace: true });
  }, [autenticado, destino, navigate]);

  async function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    setErro(null);
    if (!email.trim() || !senha) {
      setErro("Informe e-mail e senha.");
      return;
    }
    setEnviando(true);
    try {
      await entrar(email.trim(), senha);
      navigate(destino, { replace: true });
    } catch (err) {
      setErro(mensagemErro(err, "Falha ao entrar."));
    } finally {
      setEnviando(false);
    }
  }

  return (
    <div className="flex min-h-screen items-center justify-center bg-gradient-to-br from-secondary to-background p-4">
      <div className="w-full max-w-md">
        <div className="mb-6 flex flex-col items-center gap-3 text-center">
          <div className="flex h-14 w-14 items-center justify-center rounded-xl bg-primary text-primary-foreground">
            <Package className="h-7 w-7" />
          </div>
          <div>
            <h1 className="text-2xl font-semibold tracking-tight text-foreground">
              SIGE-TI
            </h1>
            <p className="text-sm text-muted-foreground">
              Sistema de Gestão de Estoque de T.I.
            </p>
          </div>
        </div>

        <div className="rounded-lg border border-border bg-card p-6 shadow-sm">
          <h2 className="mb-1 text-lg font-semibold text-foreground">Acesso</h2>
          <p className="mb-5 text-sm text-muted-foreground">
            Entre com suas credenciais funcionais.
          </p>

          <form onSubmit={onSubmit} className="space-y-4" noValidate>
            <FormField label="E-mail" htmlFor="email" obrigatorio>
              <Input
                id="email"
                type="email"
                autoComplete="username"
                placeholder="usuario@sige-ti.local"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                disabled={enviando}
              />
            </FormField>

            <FormField label="Senha" htmlFor="senha" obrigatorio>
              <Input
                id="senha"
                type="password"
                autoComplete="current-password"
                placeholder="••••••••"
                value={senha}
                onChange={(e) => setSenha(e.target.value)}
                disabled={enviando}
              />
            </FormField>

            {erro && (
              <div className="rounded-md border border-destructive/30 bg-destructive/5 px-3 py-2 text-sm text-destructive">
                {erro}
              </div>
            )}

            <Button type="submit" className="w-full" disabled={enviando}>
              {enviando ? <Spinner className="h-4 w-4" /> : "Entrar"}
            </Button>
          </form>
        </div>

        <p className="mt-6 text-center text-xs text-muted-foreground">
          Acesso restrito a servidores autorizados. Os dados são protegidos
          conforme a LGPD.
        </p>
      </div>
    </div>
  );
}
