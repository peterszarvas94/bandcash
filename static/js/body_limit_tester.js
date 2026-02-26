function nowTime() {
  return new Date().toLocaleTimeString();
}

function appendLine(output, line) {
  output.textContent = `${line}\n${output.textContent}`;
}

function makeBody(csrf, sizeBytes) {
  const payload = {
    signals: {
      csrf,
      payload: "x".repeat(sizeBytes),
    },
  };
  return JSON.stringify(payload);
}

async function sendBodyLimitRequest(output, csrf, url, body, label, msgResult) {
  const res = await fetch(url, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      "X-CSRF-Token": csrf,
    },
    credentials: "same-origin",
    body,
  });
  appendLine(output, `[${nowTime()}] ${label} ${msgResult}: ${res.status}`);
}

function initBodyLimitTester() {
  if (window.__bandcashBodyLimitTester) {
    return;
  }
  window.__bandcashBodyLimitTester = true;

  const output = document.getElementById("body-limit-output");
  const csrfInput = document.getElementById("body-limit-csrf");
  const globalOkButton = document.getElementById("body-limit-global-ok");
  const globalBigButton = document.getElementById("body-limit-global-big");
  const authOkButton = document.getElementById("body-limit-auth-ok");
  const authBigButton = document.getElementById("body-limit-auth-big");

  if (!output || !csrfInput || !globalOkButton || !globalBigButton || !authOkButton || !authBigButton) {
    return;
  }

  const csrf = csrfInput.value || "";
  const msgGlobalOK = document.getElementById("body-limit-msg-global-ok")?.textContent || "global ok";
  const msgGlobalBig = document.getElementById("body-limit-msg-global-big")?.textContent || "global too big";
  const msgAuthOK = document.getElementById("body-limit-msg-auth-ok")?.textContent || "auth ok";
  const msgAuthBig = document.getElementById("body-limit-msg-auth-big")?.textContent || "auth too big";
  const msgResult = document.getElementById("body-limit-msg-result")?.textContent || "status";

  globalOkButton.addEventListener("click", function onGlobalOkClick() {
    sendBodyLimitRequest(output, csrf, "/dev/body-limit/global", makeBody(csrf, 8 * 1024), msgGlobalOK, msgResult);
  });

  globalBigButton.addEventListener("click", function onGlobalBigClick() {
    sendBodyLimitRequest(output, csrf, "/dev/body-limit/global", makeBody(csrf, 2 * 1024 * 1024), msgGlobalBig, msgResult);
  });

  authOkButton.addEventListener("click", function onAuthOkClick() {
    sendBodyLimitRequest(output, csrf, "/dev/body-limit/auth", makeBody(csrf, 8 * 1024), msgAuthOK, msgResult);
  });

  authBigButton.addEventListener("click", function onAuthBigClick() {
    sendBodyLimitRequest(output, csrf, "/dev/body-limit/auth", makeBody(csrf, 96 * 1024), msgAuthBig, msgResult);
  });
}

initBodyLimitTester();
