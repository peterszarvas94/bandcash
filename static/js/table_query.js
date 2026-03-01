function normalizeText(value) {
  if (typeof value !== "string") {
    return "";
  }
  return value.trim();
}

function normalizePageSize(value, fallback) {
  const numeric = Number(value);
  if (!Number.isFinite(numeric) || numeric <= 0) {
    return fallback;
  }
  return Math.floor(numeric);
}

function buildTableSearchURL(input) {
  const basePath = typeof input.basePath === "string" ? input.basePath : "";
  const defaultPageSize = normalizePageSize(input.defaultPageSize, 50);
  const pageSize = normalizePageSize(input.pageSize, defaultPageSize);
  const search = normalizeText(input.search);
  const sort = normalizeText(input.sort);
  const dir = input.dir === "desc" ? "desc" : "asc";
  const currentSearch = typeof input.currentSearch === "string" ? input.currentSearch : window.location.search;
  const current = new URLSearchParams(currentSearch);

  const params = new URLSearchParams();
  if (search !== "") {
    params.set("q", search);
  }

  if (current.has("sort") && sort !== "") {
    params.set("sort", sort);
    params.set("dir", dir);
  }

  if (pageSize !== defaultPageSize) {
    params.set("pageSize", String(pageSize));
  }

  const query = params.toString();
  if (query === "") {
    return basePath;
  }

  const separator = basePath.includes("?") ? "&" : "?";
  return `${basePath}${separator}${query}`;
}

function tableSearchAction(basePath, tableQuery, defaultPageSize = 50) {
  const queryState = tableQuery && typeof tableQuery === "object" ? tableQuery : {};
  queryState.page = 1;

  const url = buildTableSearchURL({
    basePath,
    search: queryState.search,
    sort: queryState.sort,
    dir: queryState.dir,
    pageSize: queryState.pageSize,
    defaultPageSize,
  });

  history.pushState(null, "", url);
  return url;
}

window.buildTableSearchURL = buildTableSearchURL;
window.tableSearchAction = tableSearchAction;
