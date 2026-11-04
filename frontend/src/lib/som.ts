// Som de notificação gerado via Web Audio API — sem depender de arquivo de
// áudio. Toca um "ding" curto (dois tons) quando chega mensagem/notificação.
//
// Observação sobre autoplay: navegadores só liberam áudio após alguma
// interação do usuário com a página. Como o app exige login (cliques), o
// AudioContext costuma estar liberado. Tudo é best-effort: qualquer falha é
// silenciosa para nunca quebrar a UI.

let ctx: AudioContext | null = null;

function obterContexto(): AudioContext | null {
  try {
    if (!ctx) {
      const AC =
        window.AudioContext ||
        (window as unknown as { webkitAudioContext?: typeof AudioContext })
          .webkitAudioContext;
      if (!AC) return null;
      ctx = new AC();
    }
    // Pode estar suspenso até a 1ª interação; tenta retomar.
    if (ctx.state === "suspended") void ctx.resume();
    return ctx;
  } catch {
    return null;
  }
}

/** Toca um "ding" curto de notificação. Best-effort (silencioso em falha). */
export function tocarSomNotificacao() {
  const audio = obterContexto();
  if (!audio) return;
  try {
    const agora = audio.currentTime;
    // Dois tons rápidos (G5 -> C6), tipo notificação.
    const notas = [
      { freq: 784, inicio: 0, dur: 0.12 },
      { freq: 1047, inicio: 0.1, dur: 0.16 },
    ];
    for (const n of notas) {
      const osc = audio.createOscillator();
      const ganho = audio.createGain();
      osc.type = "sine";
      osc.frequency.value = n.freq;
      // Envelope para não "estalar".
      const t0 = agora + n.inicio;
      ganho.gain.setValueAtTime(0, t0);
      ganho.gain.linearRampToValueAtTime(0.25, t0 + 0.02);
      ganho.gain.exponentialRampToValueAtTime(0.0001, t0 + n.dur);
      osc.connect(ganho);
      ganho.connect(audio.destination);
      osc.start(t0);
      osc.stop(t0 + n.dur + 0.02);
    }
  } catch {
    /* best-effort */
  }
}
