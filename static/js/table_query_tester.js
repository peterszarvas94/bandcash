function nowTime() {
  return new Date().toLocaleTimeString();
}

function appendLine(output, line) {
  output.textContent = `${line}\n${output.textContent}`;
}

async function runQueryTest(output, label, path) {
  appendLine(output, `[${nowTime()}] ${label}: requesting ${path}`);

  const res = await fetch(path, { credentials: "same-origin" });
  const text = await res.text();

  let body = text;
  try {
    const asJSON = JSON.parse(text);
    body = JSON.stringify(asJSON, null, 2);
  } catch (_) {
  }

  appendLine(output, `[${nowTime()}] ${label}: status ${res.status}`);
  appendLine(output, body);
}

function initTableQueryTester() {
  if (window.__bandcashTableQueryTester) {
    return;
  }
  window.__bandcashTableQueryTester = true;

  const output = document.getElementById("query-test-output");
  const eventsValid = document.getElementById("query-test-events-valid");
  const eventsInvalid = document.getElementById("query-test-events-invalid");
  const membersValid = document.getElementById("query-test-members-valid");
  const expensesValid = document.getElementById("query-test-expenses-valid");

  if (!output || !eventsValid || !eventsInvalid || !membersValid || !expensesValid) {
    return;
  }

  eventsValid.addEventListener("click", function onEventsValidClick() {
    runQueryTest(output, "events-valid", "/dev/query-test/events?q=concert&sort=time&dir=asc&page=1&pageSize=50");
  });

  eventsInvalid.addEventListener("click", function onEventsInvalidClick() {
    runQueryTest(output, "events-invalid", "/dev/query-test/events?q=&sort=hack&dir=up&page=-3&pageSize=9999&strict=1");
  });

  membersValid.addEventListener("click", function onMembersValidClick() {
    runQueryTest(output, "members-valid", "/dev/query-test/members?q=&sort=name&dir=asc&page=1&pageSize=50");
  });

  expensesValid.addEventListener("click", function onExpensesValidClick() {
    runQueryTest(output, "expenses-valid", "/dev/query-test/expenses?q=&sort=amount&dir=desc&page=1&pageSize=50");
  });
}

initTableQueryTester();
