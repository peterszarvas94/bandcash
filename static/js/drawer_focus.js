const mobileQuery = window.matchMedia("(max-width: 1020px)");

function isMobile() {
  return mobileQuery.matches;
}

function isFocusableElement(el) {
  return !el.hasAttribute("inert") && el.offsetParent !== null;
}

function focusableElements(container) {
  if (!container) {
    return [];
  }

  const selector =
    'a[href], button:not([disabled]), input:not([disabled]):not([type="hidden"]), select:not([disabled]), textarea:not([disabled]), [tabindex]:not([tabindex="-1"])';
  const nodes = Array.from(container.querySelectorAll(selector));
  return nodes.filter(isFocusableElement);
}

function focusFirst(container) {
  const items = focusableElements(container);
  if (items.length > 0) {
    items[0].focus();
  }
}

function isVisible(el) {
  return Boolean(el && el.offsetParent !== null);
}

function isCancelButton(button) {
  if (!isVisible(button) || button.disabled) {
    return false;
  }

  const onClick = button.getAttribute("data-on:click") || "";
  return (
    onClick.includes("$formState = ''") ||
    onClick.includes('$formState = ""') ||
    onClick.includes("$eventFormState = ''") ||
    onClick.includes('$eventFormState = ""')
  );
}

function findCancelButton(panel) {
  if (!panel) {
    return null;
  }

  const buttons = Array.from(panel.querySelectorAll("button"));
  for (const button of buttons) {
    if (isCancelButton(button)) {
      return button;
    }
  }
  return null;
}

function hasVisiblePanelContent(shell) {
  const panel = shell.querySelector(".app-shell-panel");
  if (!panel) {
    return false;
  }
  return focusableElements(panel).length > 0;
}

function applyState(shell) {
  const navToggle = shell.querySelector('input[id$="-nav-toggle"]');
  const panelToggle = shell.querySelector('input[id$="-panel-toggle"]');
  const top = shell.querySelector(".app-shell-top");
  const nav = shell.querySelector(".app-shell-nav");
  const panel = shell.querySelector(".app-shell-panel");
  const main = shell.querySelector(".app-shell-main");
  const panelVisible = hasVisiblePanelContent(shell);

  shell.classList.toggle("panel-open", panelVisible);

  if (!isMobile()) {
    if (navToggle) {
      navToggle.checked = false;
    }
    if (panelToggle) {
      panelToggle.checked = false;
    }

    [top, nav, main].forEach(function(el) {
      if (el) {
        el.removeAttribute("inert");
      }
    });

    if (panel) {
      panel.toggleAttribute("inert", !panelVisible);
    }
    return;
  }

  if (panelToggle) {
    const shouldOpenPanel = panelVisible && !Boolean(navToggle?.checked);
    if (panelToggle.checked !== shouldOpenPanel) {
      panelToggle.checked = shouldOpenPanel;
    }
  }

  const navOpen = Boolean(navToggle?.checked);
  const panelOpen = Boolean(panelToggle?.checked);

  if (navOpen && panelToggle && panelToggle.checked) {
    panelToggle.checked = false;
  }

  if (panelOpen && navToggle && navToggle.checked) {
    navToggle.checked = false;
  }

  const navIsOpen = Boolean(navToggle?.checked);
  const panelIsOpen = Boolean(panelToggle?.checked);
  const drawerOpen = navIsOpen || panelIsOpen;

  if (top) {
    top.toggleAttribute("inert", drawerOpen);
  }
  if (main) {
    main.toggleAttribute("inert", drawerOpen);
  }
  if (nav) {
    nav.toggleAttribute("inert", !navIsOpen);
  }
  if (panel) {
    panel.toggleAttribute("inert", !panelIsOpen);
  }
}

function activeDrawer(shell) {
  const navToggle = shell.querySelector('input[id$="-nav-toggle"]');
  const panelToggle = shell.querySelector('input[id$="-panel-toggle"]');
  if (isMobile() && navToggle?.checked) {
    return shell.querySelector(".app-shell-nav");
  }
  if (isMobile() && panelToggle?.checked) {
    return shell.querySelector(".app-shell-panel");
  }
  return null;
}

function closePanelLikeCancel(shell) {
  const panel = shell.querySelector(".app-shell-panel");
  const panelToggle = shell.querySelector('input[id$="-panel-toggle"]');
  const cancelButton = findCancelButton(panel);

  if (cancelButton) {
    cancelButton.click();
    return true;
  }

  if (panelToggle?.checked) {
    panelToggle.checked = false;
  }
  applyState(shell);
  return false;
}

function trapFocusInActiveDrawer(event, shell) {
  if (event.key !== "Tab") {
    return;
  }

  const drawer = activeDrawer(shell);
  if (!drawer) {
    return;
  }

  const items = focusableElements(drawer);
  if (items.length === 0) {
    event.preventDefault();
    return;
  }

  const first = items[0];
  const last = items[items.length - 1];
  const active = document.activeElement;

  if (event.shiftKey && active === first) {
    event.preventDefault();
    last.focus();
    return;
  }

  if (!event.shiftKey && active === last) {
    event.preventDefault();
    first.focus();
    return;
  }

  if (!drawer.contains(active)) {
    event.preventDefault();
    first.focus();
  }
}

function observePanel(panel, sync) {
  if (!panel) {
    return;
  }

  const observer = new MutationObserver(sync);
  observer.observe(panel, {
    subtree: true,
    childList: true,
    attributes: true,
    attributeFilter: ["style", "class", "hidden", "data-show"],
  });
}

function wireShell(shell) {
  const navToggle = shell.querySelector('input[id$="-nav-toggle"]');
  const panelToggle = shell.querySelector('input[id$="-panel-toggle"]');
  const panel = shell.querySelector(".app-shell-panel");
  const panelBackdrop = shell.querySelector(".app-shell-backdrop-panel");

  function sync() {
    applyState(shell);
  }

  if (navToggle) {
    navToggle.addEventListener("change", function onNavChange() {
      sync();
      if (navToggle.checked) {
        focusFirst(shell.querySelector(".app-shell-nav"));
      }
    });
  }

  if (panelToggle) {
    panelToggle.addEventListener("change", function onPanelChange() {
      if (isMobile() && !panelToggle.checked && hasVisiblePanelContent(shell)) {
        closePanelLikeCancel(shell);
        return;
      }

      sync();
      if (panelToggle.checked) {
        focusFirst(panel);
      }
    });
  }

  if (panelBackdrop) {
    panelBackdrop.addEventListener("click", function onPanelBackdropClick(event) {
      event.preventDefault();
      closePanelLikeCancel(shell);
    });
  }

  observePanel(panel, sync);

  shell.addEventListener(
    "keydown",
    function onShellKeydown(event) {
      trapFocusInActiveDrawer(event, shell);
    },
    true,
  );

  sync();
}

function applyStateToAllShells() {
  const shells = document.querySelectorAll(".app-shell, .app-shell-main-only");
  shells.forEach(function(shell) {
    applyState(shell);
  });
}

function initDrawerFocus() {
  const shells = document.querySelectorAll(".app-shell, .app-shell-main-only");
  shells.forEach(function(shell) {
    wireShell(shell);
  });
}

function onViewportChange() {
  applyStateToAllShells();
}

initDrawerFocus();
mobileQuery.addEventListener("change", onViewportChange);
