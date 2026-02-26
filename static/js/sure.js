let pendingForm = null;

function resolveForm(source) {
  if (source instanceof HTMLFormElement) {
    return source;
  }
  if (source instanceof Element) {
    return source.closest("form");
  }
  return null;
}

function getDialog() {
  const popover = document.getElementById("sure-popover");
  const title = document.getElementById("sure-title");
  const message = document.getElementById("sure-message");
  const cancel = document.getElementById("sure-cancel");
  const submit = document.getElementById("sure-submit");
  if (!popover || !title || !message || !cancel || !submit) {
    return null;
  }

  return { popover, title, message, cancel, submit };
}

function closeSure(dialog) {
  pendingForm = null;
  if (typeof dialog.popover.hidePopover === "function") {
    dialog.popover.hidePopover();
  }
}

function confirmSure(dialog) {
  const form = pendingForm;
  closeSure(dialog);
  if (form) {
    form.requestSubmit();
  }
}

function sure(source, title = "", message = "", submit = "", cancel = "") {
  const form = resolveForm(source);

  if (!(form instanceof HTMLFormElement)) {
    return false;
  }

  const dialog = getDialog();
  const titleText = title || "Confirm";
  const messageText = message || "";
  if (!dialog) {
    const text = messageText ? `${titleText}\n\n${messageText}` : titleText;
    if (window.confirm(text)) {
      form.requestSubmit();
    }
    return false;
  }

  pendingForm = form;
  dialog.title.textContent = titleText;
  dialog.message.textContent = messageText;
  dialog.message.style.display = messageText ? "block" : "none";
  dialog.submit.textContent = submit || "Confirm";
  dialog.cancel.textContent = cancel || "Cancel";
  dialog.popover.showPopover();
  dialog.submit.focus();
  return false;
}

function initSureDialog() {
  const dialog = getDialog();
  if (!dialog) {
    return;
  }

  dialog.cancel.addEventListener("click", function onCancelClick() {
    closeSure(dialog);
  });

  dialog.submit.addEventListener("click", function onSubmitClick() {
    confirmSure(dialog);
  });

  dialog.popover.addEventListener("click", function onBackdropClick(event) {
    if (event.target === dialog.popover) {
      closeSure(dialog);
    }
  });

  document.addEventListener("keydown", function onSureKeydown(event) {
    if (!pendingForm) {
      return;
    }
    if (event.key === "Escape") {
      event.preventDefault();
      closeSure(dialog);
    }
  });
}

window.sure = sure;
initSureDialog();
