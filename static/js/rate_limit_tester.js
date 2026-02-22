if (!window.__bandcashRateLimitTester) {
  window.__bandcashRateLimitTester = true;

  const output = document.getElementById("rate-limit-output");
  const csrfInput = document.getElementById("rate-limit-csrf");
  const healthButton = document.getElementById("spam-health");
  const authButton = document.getElementById("spam-auth");
  if (output && csrfInput && healthButton && authButton) {
    const append = (line) => {
      output.textContent = `${line}\n${output.textContent}`;
    };

    const summarize = (statuses) => {
      const parts = [];
      const sorted = Array.from(statuses.keys()).sort((a, b) => a - b);
      for (const status of sorted) {
        parts.push(`${status}:${statuses.get(status)}`);
      }
      return parts.join(" ");
    };

    const msgStartHealth = document.getElementById("rate-limit-msg-start-health")?.textContent || "start health spam";
    const msgDoneHealth = document.getElementById("rate-limit-msg-done-health")?.textContent || "health done";
    const msgStartAuth = document.getElementById("rate-limit-msg-start-auth")?.textContent || "start auth spam";
    const msgDoneAuth = document.getElementById("rate-limit-msg-done-auth")?.textContent || "auth done";
    const now = () => new Date().toLocaleTimeString();

    healthButton.addEventListener("click", async () => {
      const total = 80;
      const statuses = new Map();
      append(`[${now()}] ${msgStartHealth} (${total})`);
      await Promise.all(
        Array.from({ length: total }, async () => {
          const res = await fetch("/health", { credentials: "same-origin" });
          statuses.set(res.status, (statuses.get(res.status) || 0) + 1);
        }),
      );
      append(`[${now()}] ${msgDoneHealth} -> ${summarize(statuses)}`);
    });

    authButton.addEventListener("click", async () => {
      const total = 15;
      const statuses = new Map();
      const csrf = csrfInput.value || "";
      append(`[${now()}] ${msgStartAuth} (${total})`);
      await Promise.all(
        Array.from({ length: total }, async (_, i) => {
          const email = `missing-${Date.now()}-${i}@example.invalid`;
          const body = {
            signals: {
              csrf,
              formData: { email },
            },
          };

          const res = await fetch("/auth/login", {
            method: "POST",
            headers: {
              "Content-Type": "application/json",
              "X-CSRF-Token": csrf,
            },
            credentials: "same-origin",
            body: JSON.stringify(body),
          });
          statuses.set(res.status, (statuses.get(res.status) || 0) + 1);
        }),
      );
      append(`[${now()}] ${msgDoneAuth} -> ${summarize(statuses)}`);
    });
  }
}
