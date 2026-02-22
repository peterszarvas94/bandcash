if (!window.__bandcashBodyLimitTester) {
  window.__bandcashBodyLimitTester = true;

  const output = document.getElementById("body-limit-output");
  const csrfInput = document.getElementById("body-limit-csrf");
  const globalOkButton = document.getElementById("body-limit-global-ok");
  const globalBigButton = document.getElementById("body-limit-global-big");
  const authOkButton = document.getElementById("body-limit-auth-ok");
  const authBigButton = document.getElementById("body-limit-auth-big");

  if (output && csrfInput && globalOkButton && globalBigButton && authOkButton && authBigButton) {
    const now = () => new Date().toLocaleTimeString();
    const csrf = csrfInput.value || "";

    const msgGlobalOK = document.getElementById("body-limit-msg-global-ok")?.textContent || "global ok";
    const msgGlobalBig = document.getElementById("body-limit-msg-global-big")?.textContent || "global too big";
    const msgAuthOK = document.getElementById("body-limit-msg-auth-ok")?.textContent || "auth ok";
    const msgAuthBig = document.getElementById("body-limit-msg-auth-big")?.textContent || "auth too big";
    const msgResult = document.getElementById("body-limit-msg-result")?.textContent || "status";

    const append = (line) => {
      output.textContent = `${line}\n${output.textContent}`;
    };

    const makeBody = (sizeBytes) => {
      const payload = {
        signals: {
          csrf,
          payload: "x".repeat(sizeBytes),
        },
      };
      return JSON.stringify(payload);
    };

    const send = async (url, body, label) => {
      const res = await fetch(url, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
          "X-CSRF-Token": csrf,
        },
        credentials: "same-origin",
        body,
      });
      append(`[${now()}] ${label} ${msgResult}: ${res.status}`);
    };

    globalOkButton.addEventListener("click", () => {
      send("/dev/body-limit/global", makeBody(8 * 1024), msgGlobalOK);
    });

    globalBigButton.addEventListener("click", () => {
      send("/dev/body-limit/global", makeBody(2 * 1024 * 1024), msgGlobalBig);
    });

    authOkButton.addEventListener("click", () => {
      send("/dev/body-limit/auth", makeBody(8 * 1024), msgAuthOK);
    });

    authBigButton.addEventListener("click", () => {
      send("/dev/body-limit/auth", makeBody(96 * 1024), msgAuthBig);
    });
  }
}
