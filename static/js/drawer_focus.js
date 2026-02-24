const mobileQuery = window.matchMedia("(max-width: 1020px)");

function isMobile() {
  return mobileQuery.matches;
}

function focusableElements(container) {
  if (!container) {
    return [];
  }
  return Array.from(
    container.querySelectorAll(
      'a[href], button:not([disabled]), input:not([disabled]):not([type="hidden"]), select:not([disabled]), textarea:not([disabled]), [tabindex]:not([tabindex="-1"])'
    )
  ).filter((el) => !el.hasAttribute("inert") && el.offsetParent !== null);
}

function focusFirst(container) {
  const items = focusableElements(container);
  if (items.length > 0) {
    items[0].focus();
  }
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

    [top, nav, main].forEach((el) => el?.removeAttribute("inert"));
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

function wireShell(shell) {
  const navToggle = shell.querySelector('input[id$="-nav-toggle"]');
  const panelToggle = shell.querySelector('input[id$="-panel-toggle"]');
  const panel = shell.querySelector(".app-shell-panel");

  const sync = () => applyState(shell);

  navToggle?.addEventListener("change", () => {
    sync();
    if (navToggle.checked) {
      focusFirst(shell.querySelector(".app-shell-nav"));
    }
  });

  panelToggle?.addEventListener("change", () => {
    sync();
    if (panelToggle.checked) {
      focusFirst(panel);
    }
  });

  if (panel) {
    const observer = new MutationObserver(sync);
    observer.observe(panel, {
      subtree: true,
      childList: true,
      attributes: true,
      attributeFilter: ["style", "class", "hidden", "data-show"],
    });
  }

  shell.addEventListener("keydown", (event) => {
    if (event.key === "Escape") {
      if (navToggle?.checked) {
        navToggle.checked = false;
      }
      if (panelToggle?.checked) {
        panelToggle.checked = false;
      }
      sync();
      return;
    }

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
    } else if (!event.shiftKey && active === last) {
      event.preventDefault();
      first.focus();
    } else if (!drawer.contains(active)) {
      event.preventDefault();
      first.focus();
    }
  });

  sync();
}

function init() {
  const shells = document.querySelectorAll(".app-shell, .app-shell-main-only");
  shells.forEach(wireShell);
}

mobileQuery.addEventListener("change", () => {
  document
    .querySelectorAll(".app-shell, .app-shell-main-only")
    .forEach((shell) => applyState(shell));
});

if (document.readyState === "loading") {
  document.addEventListener("DOMContentLoaded", init);
} else {
  init();
}
