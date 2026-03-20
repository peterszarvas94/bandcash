const timeoutMs = 5000;
const timers = {};

function onNotificationTimeout(notificationID) {
  delete timers[notificationID];
  const currentNode = document.getElementById(`notification-item-${notificationID}`);
  if (currentNode) {
    currentNode.remove();
  }
  syncNotifications();
}

function scheduleNotificationRemoval(notificationID, delayMs) {
  timers[notificationID] = window.setTimeout(function timeoutHandler() {
    onNotificationTimeout(notificationID);
  }, delayMs);
}

function syncNotifications() {
  const popover = document.getElementById("notifications-popover");
  if (!popover) {
    return;
  }

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
    window.clearTimeout(timers[notificationID]);
    delete timers[notificationID];
  }

  for (const node of nodes) {
    const notificationID = node.getAttribute("data-notification-id");
    if (!notificationID || timers[notificationID]) {
      continue;
    }

    const createdMs = Number.parseInt(node.getAttribute("data-created-at-ms") || "", 10);
    const ageMs = Number.isNaN(createdMs) ? 0 : Math.max(0, nowMs - createdMs);
    const delayMs = Math.max(0, timeoutMs - ageMs);

    scheduleNotificationRemoval(notificationID, delayMs);
  }

  popover.style.display = list.children.length > 0 ? "block" : "none";
}

function onNotificationsMutate() {
  syncNotifications();
}

function initNotifications() {
  if (!window.__bandcashNotificationsObserver) {
    window.__bandcashNotificationsObserver = new MutationObserver(onNotificationsMutate);
    window.__bandcashNotificationsObserver.observe(document.body, {
      childList: true,
      subtree: true,
    });
  }

  syncNotifications();
}

initNotifications();
