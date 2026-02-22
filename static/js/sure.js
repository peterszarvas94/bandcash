const approvedSubmits = new WeakSet();
const approvedClicks = new WeakSet();
const fallbackSure = (title) => Promise.resolve(window.confirm(title || "Confirm"));

const getSureDefaults = () => {
  const popover = document.getElementById("sure-popover");
  if (!popover) {
    return {
      title: "Confirm",
      submit: "Confirm",
      cancel: "Cancel",
    };
  }

  return {
    title: popover.dataset.defaultTitle || "Confirm",
    submit: popover.dataset.defaultSubmit || "Confirm",
    cancel: popover.dataset.defaultCancel || "Cancel",
  };
};

window.sure = function(options = {}) {
  const popover = document.getElementById("sure-popover");
  if (!popover) {
    return fallbackSure(options.title);
  }

  const title = document.getElementById("sure-title");
  const message = document.getElementById("sure-message");
  const cancel = document.getElementById("sure-cancel");
  const submit = document.getElementById("sure-submit");
  if (!title || !message || !cancel || !submit) {
    return fallbackSure(options.title);
  }

  const defaults = getSureDefaults();
  title.textContent = options.title || defaults.title;
  message.textContent = options.message || "";
  submit.textContent = options.submit || defaults.submit;
  cancel.textContent = options.cancel || defaults.cancel;
  message.style.display = (options.message || "").trim() === "" ? "none" : "block";

  return new Promise((resolve) => {
    let settled = false;

    const cleanup = () => {
      document.removeEventListener("keydown", onKeydown);
      popover.removeEventListener("click", onBackdropClick);
      cancel.removeEventListener("click", onCancel);
      submit.removeEventListener("click", onSubmit);
      if (typeof popover.hidePopover === "function") {
        try {
          popover.hidePopover();
        } catch (_) {
          // ignore
        }
      }
    };

    const settle = (result) => {
      if (settled) {
        return;
      }
      settled = true;
      cleanup();
      resolve(result);
    };

    const onKeydown = (event) => {
      if (event.key === "Escape") {
        event.preventDefault();
        settle(false);
      } else if (event.key === "Enter") {
        event.preventDefault();
        settle(true);
      }
    };

    const onBackdropClick = (event) => {
      if (event.target === popover) {
        settle(false);
      }
    };

    const onCancel = () => settle(false);
    const onSubmit = () => settle(true);

    document.addEventListener("keydown", onKeydown);
    popover.addEventListener("click", onBackdropClick);
    cancel.addEventListener("click", onCancel);
    submit.addEventListener("click", onSubmit);

    if (typeof popover.showPopover === "function") {
      popover.showPopover();
    }
    submit.focus();
  });
};

const optionsFromElement = (el) => ({
  title: el.dataset.sureTitle || "",
  message: el.dataset.sureMessage || "",
  submit: el.dataset.sureSubmit || "",
  cancel: el.dataset.sureCancel || "",
});

document.addEventListener(
  "submit",
  (event) => {
    const form = event.target;
    if (!(form instanceof HTMLFormElement) || !form.dataset.sureTitle) {
      return;
    }

    if (approvedSubmits.has(form)) {
      approvedSubmits.delete(form);
      return;
    }

    event.preventDefault();
    event.stopImmediatePropagation();

    window.sure(optionsFromElement(form)).then((ok) => {
      if (!ok) {
        return;
      }
      approvedSubmits.add(form);
      if (typeof form.requestSubmit === "function") {
        form.requestSubmit();
      } else {
        form.submit();
      }
    });
  },
  true,
);

document.addEventListener(
  "click",
  (event) => {
    const target = event.target instanceof Element ? event.target.closest("[data-sure-title]") : null;
    if (!target || target.tagName === "FORM") {
      return;
    }

    if (approvedClicks.has(target)) {
      approvedClicks.delete(target);
      return;
    }

    event.preventDefault();
    event.stopImmediatePropagation();

    window.sure(optionsFromElement(target)).then((ok) => {
      if (!ok) {
        return;
      }
      approvedClicks.add(target);
      target.click();
    });
  },
  true,
);
