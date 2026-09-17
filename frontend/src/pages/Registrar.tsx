import * as React from "react";
import { useNavigate } from "react-router-dom";
import { LifeBuoy } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { FormField } from "@/components/FormField";
import { Spinner } from "@/components/ui/spinner";
import { useAuth } from "@/contexts/AuthContext";
import { camposInvalidos, mensagemErro } from "@/services/api/client";

export default function Registrar() {
  const { registrar, autenticado } = useAuth();
  const navigate = useNavigate();

  const [nome, setNome] = React.useState("");
  const [email, setEmail] = React.useState("");
  const [senha, setSenha] = React.useState("");
  const [erros, setErros] = React.useState<Record<string, string>>({});
  const [erroGeral, setErroGeral] = React.useState<string | null>(null);
  const [enviando, setEnviando] = React.useState(false);

  React.useEffect(() => {
    if (autenticado) navigate("/", { replace: true });
  }, [autenticado, navigate]);

  async function onSubmit(e: React.FormEvent) {
    e.preventDefault();
    setErros({});
    setErroGeral(null);
    const novos: Record<string, string> = {};
    if (!nome.trim()) novos.nome = "Informe seu nome.";
    if (!email.trim()) novos.email = "Informe seu e-mail.";
    if (senha.length < 6) novos.senha = "A senha deve ter ao menos 6 caracteres.";
    if (Object.keys(novos).length) {
      setErros(novos);
      return;
    }
    setEnviando(true);
    try {
      await registrar(nome.trim(), email.trim(), senha);
      navigate("/", { replace: true });
    } catch (err) {
      const campos = camposInvalidos(err);
      if (campos) setErros(campos);
      else setErroGeral(mensagemErro(err, "Não foi possível criar a conta."));
    } finally {
      setEnviando(false);
    }
  }

  return (
    <div className="flex min-h-screen items-center justify-center bg-gradient-to-br from-secondary to-background p-4">
      <div className="w-full max-w-md">
        <div className="mb-6 flex flex-col items-center gap-3 text-center">
          <div className="flex h-14 w-14 items-center justify-center rounded-xl bg-primary text-primary-foreground">
            <LifeBuoy className="h-7 w-7" />
          </div>
          <div>
            <h1 className="text-2xl font-semibold tracking-tight text-foreground">
              Abertura de Chamados
            </h1>
            <p className="text-sm text-muted-foreground">
              Crie sua conta para abrir e acompanhar chamados de T.I.
            </p>
          </div>
        </div>

        <div className="rounded-lg border border-border bg-card p-6 shadow-sm">
          <h2 className="mb-5 text-lg font-semibold text-foreground">
            Criar conta
          </h2>

          <form onSubmit={onSubmit} className="space-y-4" noValidate>
            <FormField label="Nome completo" htmlFor="nome" obrigatorio erro={erros.nome}>
              <Input
                id="nome"
                value={nome}
                onChange={(e) => setNome(e.target.value)}
                placeholder="Seu nome"
                disabled={enviando}
              />
            </FormField>
            <FormField label="E-mail" htmlFor="email" obrigatorio erro={erros.email}>
              <Input
                id="email"
                type="email"
                autoComplete="username"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                placeholder="voce@prefeitura.gov.br"
                disabled={enviando}
              />
            </FormField>
            <FormField label="Senha" htmlFor="senha" obrigatorio erro={erros.senha}>
              <Input
                id="senha"
                type="password"
                autoComplete="new-password"
                value={senha}
                onChange={(e) => setSenha(e.target.value)}
                placeholder="Mínimo 6 caracteres"
                disabled={enviando}
              />
            </FormField>

            {erroGeral && (
              <div className="rounded-md border border-destructive/30 bg-destructive/5 px-3 py-2 text-sm text-destructive">
                {erroGeral}
              </div>
            )}

            <Button type="submit" className="w-full" disabled={enviando}>
              {enviando ? <Spinner className="h-4 w-4" /> : "Criar conta e entrar"}
            </Button>
          </form>

          <p className="mt-4 text-center text-sm text-muted-foreground">
            Já tem conta?{" "}
            <button
              type="button"
              className="font-medium text-primary hover:underline"
              onClick={() => navigate("/login")}
            >
              Entrar
            </button>
          </p>
        </div>

        <p className="mt-6 text-center text-xs text-muted-foreground">
          Seus dados são protegidos conforme a LGPD.
        </p>
      </div>
    </div>
  );
}
