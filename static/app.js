let serverToday = "";
let products = [];
let chartMonths = [];
let festivals = [];
let messageTimer = 0;

const tabs = ["dashboard", "products", "sell", "stock-in", "reports", "suggestions", "festivals"];
const monthNames = ["Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"];
const kindNames = { daily: "Daily", monthly: "Monthly", yearly: "Yearly" };

async function api(url, data) {
  let options = {};
  if (data) {
    options = {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(data),
    };
  }
  let res;
  try {
    res = await fetch(url, options);
  } catch (err) {
    throw new Error('Can not reach the server. Is "go run ." still running?');
  }
  if (!res.ok) {
    throw new Error((await res.text()).trim());
  }
  return res.json();
}

function showMessage(text, isError) {
  const box = document.getElementById("message");
  box.textContent = text;
  box.classList.toggle("error", isError === true);
  box.classList.remove("hidden");
  clearTimeout(messageTimer);
  messageTimer = setTimeout(() => box.classList.add("hidden"), 4000);
}

function safe(text) {
  return String(text)
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;");
}

function money(value) {
  return value.toLocaleString("en-IN", { minimumFractionDigits: 2, maximumFractionDigits: 2 });
}

function emptyRow(columns, text) {
  return `<tr><td colspan="${columns}" class="empty">${text}</td></tr>`;
}

function findProduct(id) {
  return products.find((p) => p.id === Number(id));
}

function showTab(name) {
  tabs.forEach((tab) => {
    document.getElementById(tab).classList.toggle("hidden", tab !== name);
  });
  document.querySelectorAll("nav button").forEach((button) => {
    button.classList.toggle("active", button.dataset.tab === name);
  });
}

async function start() {
  showTab("dashboard");
}

start();
