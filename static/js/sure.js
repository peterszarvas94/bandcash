function getDialog() {
  const popover = document.getElementById("sure-popover");
  const title = document.getElementById("sure-title");
  const message = document.getElementById("sure-message");
  const cancel = document.getElementById("sure-cancel");
  const submit = document.getElementById("sure-submit");
  return { popover, title, message, cancel, submit };
}

function sure(title = "", message = "", submit = "", cancel = "") {
  const dialog = getDialog();
  const titleText = title || "Confirm";
  const messageText = message || "";

  dialog.title.textContent = titleText;
  dialog.message.textContent = messageText;
  dialog.message.style.display = messageText.trim() ? "block" : "none";
  dialog.submit.textContent = submit;
  dialog.cancel.textContent = cancel;

  return new Promise(function(resolve) {
    function done(result) {
      document.removeEventListener("keydown", onKeydown);
      dialog.popover.removeEventListener("click", onBackdropClick);
      dialog.cancel.removeEventListener("click", onCancel);
      dialog.submit.removeEventListener("click", onSubmit);
      if (typeof dialog.popover.hidePopover === "function") {
        dialog.popover.hidePopover();
      }
      resolve(result);
    }

    function onCancel() {
      done(false);
    }

    function onSubmit() {
      done(true);
    }

    function onBackdropClick(event) {
      if (event.target === dialog.popover) {
        done(false);
      }
    }

    function onKeydown(event) {
      if (event.key === "Escape") {
        event.preventDefault();
        done(false);
      }
      if (event.key === "Enter") {
        event.preventDefault();
        done(true);
      }
    }

    document.addEventListener("keydown", onKeydown);
    dialog.popover.addEventListener("click", onBackdropClick);
    dialog.cancel.addEventListener("click", onCancel);
    dialog.submit.addEventListener("click", onSubmit);

    if (typeof dialog.popover.showPopover === "function") {
      dialog.popover.showPopover();
    }
    dialog.submit.focus();
  });
}

window.sure = sure;
