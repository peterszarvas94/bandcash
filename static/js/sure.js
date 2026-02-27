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

function sure(title = "", message = "", submit = "", cancel = "", onConfirm) {
  const dialog = getDialog();
  if (!dialog) {
    return false;
  }

  dialog.title.textContent = title || "Confirm";
  dialog.message.textContent = message || "";
  dialog.message.style.display = message ? "block" : "none";
  dialog.submit.textContent = submit || "Confirm";
  dialog.cancel.textContent = cancel || "Cancel";

  function cleanup() {
    dialog.cancel.removeEventListener("click", onCancelClick);
    dialog.submit.removeEventListener("click", onSubmitClick);
    dialog.popover.removeEventListener("click", onBackdropClick);
    document.removeEventListener("keydown", onKeydown);
    if (typeof dialog.popover.hidePopover === "function") {
      dialog.popover.hidePopover();
    }
  }

  function onCancelClick() {
    cleanup();
  }

  function onSubmitClick() {
    cleanup();
    if (typeof onConfirm === "function") {
      onConfirm();
    }
  }

  function onBackdropClick(event) {
    if (event.target === dialog.popover) {
      cleanup();
    }
  }

  function onKeydown(event) {
    if (event.key === "Escape") {
      event.preventDefault();
      cleanup();
    }
  }

  dialog.cancel.addEventListener("click", onCancelClick);
  dialog.submit.addEventListener("click", onSubmitClick);
  dialog.popover.addEventListener("click", onBackdropClick);
  document.addEventListener("keydown", onKeydown);
  dialog.popover.showPopover();
  dialog.submit.focus();
  return false;
}

window.sure = sure;
