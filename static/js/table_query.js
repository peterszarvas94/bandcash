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

function normalizePage(value, fallback, totalPages = 0) {
  const numeric = Number(value);
  if (!Number.isFinite(numeric) || numeric <= 0) {
    return fallback;
  }

  let page = Math.floor(numeric);
  const maxPage = Number(totalPages);
  if (Number.isFinite(maxPage) && maxPage > 0) {
    page = Math.min(page, Math.floor(maxPage));
  }

  return page;
}

function buildTableSearchURL(input) {
  const basePath = typeof input.basePath === "string" ? input.basePath : "";
  const defaultPageSize = normalizePageSize(input.defaultPageSize, 10);
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

function tableSearchAction(basePath, tableQuery, defaultPageSize = 10) {
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

function tablePageAction(basePath, tableQuery, totalPages = 0, defaultPageSize = 10) {
  const queryState = tableQuery && typeof tableQuery === "object" ? tableQuery : {};
  const page = normalizePage(queryState.page, 1, totalPages);
  const pageSize = normalizePageSize(queryState.pageSize, normalizePageSize(defaultPageSize, 10));
  const search = normalizeText(queryState.search);
  const sort = normalizeText(queryState.sort);
  const sortSet = Boolean(queryState.sortSet) && sort !== "";
  const dir = queryState.dir === "desc" ? "desc" : "asc";

  const baseURL = new URL(basePath, window.location.origin);
  const params = baseURL.searchParams;

  if (search !== "") {
    params.set("q", search);
  } else {
    params.delete("q");
  }

  if (sortSet) {
    params.set("sort", sort);
    params.set("dir", dir);
  } else {
    params.delete("sort");
    params.delete("dir");
  }

  if (page > 1) {
    params.set("page", String(page));
  } else {
    params.delete("page");
  }

  if (pageSize !== normalizePageSize(defaultPageSize, 10)) {
    params.set("pageSize", String(pageSize));
  } else {
    params.delete("pageSize");
  }

  const query = params.toString();
  const url = query === "" ? baseURL.pathname : `${baseURL.pathname}?${query}`;
  history.pushState(null, "", url);
  return url;
}

window.buildTableSearchURL = buildTableSearchURL;
window.tableSearchAction = tableSearchAction;
window.tablePageAction = tablePageAction;
