const tableQueryKeys = new Set([
  "q",
  "sort",
  "dir",
  "page",
  "pageSize",
  "summary",
  "year",
  "dateMode",
  "from",
  "to",
]);

let hasPushedTableQuery = false;

function mergeTableQueryURL(targetURLString) {
  const currentURL = new URL(window.location.href);
  const targetURL = new URL(targetURLString, currentURL);

  if (
    targetURL.origin !== currentURL.origin ||
    targetURL.pathname !== currentURL.pathname
  ) {
    return targetURL.toString();
  }

  const mergedURL = new URL(currentURL.toString());
  for (const key of tableQueryKeys) {
    mergedURL.searchParams.delete(key);
  }

  for (const [key, value] of targetURL.searchParams.entries()) {
    if (tableQueryKeys.has(key)) {
      mergedURL.searchParams.append(key, value);
    }
  }

  mergedURL.hash = targetURL.hash;
  return mergedURL.toString();
}

function pushTableQueryURL(targetURLString) {
  const url = mergeTableQueryURL(targetURLString);
  window.history.pushState({ bandcashTableQuery: true }, "", url);
  hasPushedTableQuery = true;
  return url;
}

window.addEventListener("popstate", function onPopState() {
  if (!hasPushedTableQuery) {
    return;
  }

  window.location.assign(window.location.href);
});

window.bandcashTableQuery = {
  push: pushTableQueryURL,
};
