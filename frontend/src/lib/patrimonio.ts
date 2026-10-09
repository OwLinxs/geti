// Código de controle da T.I. derivado do ID estável do registro. Ele é
// independente do número oficial de patrimônio, que pode chegar depois.
export function codigoInterno(id: number): string {
  return `TI-${String(id).padStart(6, "0")}`;
}
