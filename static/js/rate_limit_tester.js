function appendLine(output, line) {
  output.textContent = `${line}\n${output.textContent}`;
}

function summarizeStatuses(statuses) {
  const parts = [];
  const sorted = Array.from(statuses.keys()).sort(function(a, b) {
    return a - b;
  });

  for (const status of sorted) {
    parts.push(`${status}:${statuses.get(status)}`);
  }
  return parts.join(" ");
}

function nowTime() {
  return new Date().toLocaleTimeString();
}

async function runHealthSpam(output, msgStartHealth, msgDoneHealth) {
  const total = 80;
  const statuses = new Map();
  appendLine(output, `[${nowTime()}] ${msgStartHealth} (${total})`);

  const requests = Array.from({ length: total }, async function sendHealthRequest() {
    const res = await fetch("/health", { credentials: "same-origin" });
    statuses.set(res.status, (statuses.get(res.status) || 0) + 1);
  });
  await Promise.all(requests);

  appendLine(output, `[${nowTime()}] ${msgDoneHealth} -> ${summarizeStatuses(statuses)}`);
}

async function runAuthSpam(output, csrf, msgStartAuth, msgDoneAuth) {
  const total = 15;
  const statuses = new Map();
  appendLine(output, `[${nowTime()}] ${msgStartAuth} (${total})`);

  const requests = Array.from({ length: total }, async function sendAuthRequest(_, i) {
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
  });

  await Promise.all(requests);
  appendLine(output, `[${nowTime()}] ${msgDoneAuth} -> ${summarizeStatuses(statuses)}`);
}

function initRateLimitTester() {
  if (window.__bandcashRateLimitTester) {
    return;
  }
  window.__bandcashRateLimitTester = true;

  const output = document.getElementById("rate-limit-output");
  const csrfInput = document.getElementById("rate-limit-csrf");
  const healthButton = document.getElementById("spam-health");
  const authButton = document.getElementById("spam-auth");
  if (!output || !csrfInput || !healthButton || !authButton) {
    return;
  }

  const msgStartHealth = document.getElementById("rate-limit-msg-start-health")?.textContent || "start health spam";
  const msgDoneHealth = document.getElementById("rate-limit-msg-done-health")?.textContent || "health done";
  const msgStartAuth = document.getElementById("rate-limit-msg-start-auth")?.textContent || "start auth spam";
  const msgDoneAuth = document.getElementById("rate-limit-msg-done-auth")?.textContent || "auth done";

  healthButton.addEventListener("click", async function onHealthClick() {
    await runHealthSpam(output, msgStartHealth, msgDoneHealth);
  });

  authButton.addEventListener("click", async function onAuthClick() {
    const csrf = csrfInput.value || "";
    await runAuthSpam(output, csrf, msgStartAuth, msgDoneAuth);
  });
}

initRateLimitTester();
