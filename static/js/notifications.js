const timeoutMs = 5000;
const timers = {};
const boundPopovers = new WeakSet();

function clearNotificationTimer(notificationID) {
  const timerID = timers[notificationID];
  if (!timerID) {
    return;
  }

  window.clearTimeout(timerID);
  delete timers[notificationID];
}

function onNotificationTimeout(notificationID) {
  clearNotificationTimer(notificationID);
  const currentNode = document.getElementById(
    `notification-item-${notificationID}`,
  );
  if (currentNode) {
    currentNode.remove();
  }
  syncNotifications();
}

function scheduleNotificationRemoval(notificationID, delayMs) {
  clearNotificationTimer(notificationID);
  timers[notificationID] = window.setTimeout(function timeoutHandler() {
    onNotificationTimeout(notificationID);
  }, delayMs);
}

function closeNotificationNode(node) {
  if (!node) {
    return;
  }

  const notificationID = node.getAttribute("data-notification-id");
  if (notificationID) {
    clearNotificationTimer(notificationID);
  }

  node.remove();
  syncNotifications();
}

function onPopoverClick(event) {
  const popover = event.currentTarget;
  const closeButton = event.target.closest("button");
  if (!closeButton || !popover.contains(closeButton)) {
    return;
  }

  const notificationNode = closeButton.closest("[data-notification-id]");
  if (!notificationNode || !popover.contains(notificationNode)) {
    return;
  }

  closeNotificationNode(notificationNode);
}

function bindPopoverHandlers(popover) {
  if (boundPopovers.has(popover)) {
    return;
  }

  popover.addEventListener("click", onPopoverClick);
  boundPopovers.add(popover);
}

function syncNotifications() {
  const popover = document.getElementById("notifications-popover");
  if (!popover) {
    for (const notificationID of Object.keys(timers)) {
      clearNotificationTimer(notificationID);
    }
    return;
  }

  bindPopoverHandlers(popover);

  const list = popover.querySelector("ul");
  if (!list) {
    return;
  }

  const nodes = list.querySelectorAll("[data-notification-id]");
  const nowMs = Date.now();

  for (const notificationID of Object.keys(timers)) {
    if (document.getElementById(`notification-item-${notificationID}`)) {
      continue;
    }
    clearNotificationTimer(notificationID);
  }

  for (const node of nodes) {
    const notificationID = node.getAttribute("data-notification-id");
    if (!notificationID || timers[notificationID]) {
      continue;
    }

    const createdMs = Number.parseInt(
      node.getAttribute("data-created-at-ms") || "",
      10,
    );
    const ageMs = Number.isNaN(createdMs) ? 0 : Math.max(0, nowMs - createdMs);
    const delayMs = Math.max(0, timeoutMs - ageMs);

    scheduleNotificationRemoval(notificationID, delayMs);
  }

  popover.style.display = list.children.length > 0 ? "block" : "none";
}

if (!window.__bandcashNotificationsObserver) {
  window.__bandcashNotificationsObserver = new MutationObserver(syncNotifications);
  window.__bandcashNotificationsObserver.observe(document.body, {
    childList: true,
    subtree: true,
  });
}

syncNotifications();
