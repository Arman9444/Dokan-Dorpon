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
// setMenuOpen opens or closes the hamburger menu in the top right corner.
function setMenuOpen(open) {
  document.querySelector(".menu").classList.toggle("open", open);
  document.getElementById("menu-button").setAttribute("aria-expanded", open);
}

function toggleMenu() {
  setMenuOpen(!document.querySelector(".menu").classList.contains("open"));
}
function showTab(name) {
  tabs.forEach((tab) => {
    document.getElementById(tab).classList.toggle("hidden", tab !== name);
  });
  document.querySelectorAll("nav button").forEach((button) => {
    button.classList.toggle("active", button.dataset.tab === name);
  });
    setMenuOpen(false);
  if(name==="dashboard"){
    loadDashboard();
  }
}

// ---------- Header ----------
async function loadOccasions() {
  try {
    const data = await api("/api/occasions");
    serverToday = data.today;
    // In demo mode (go run . -today=...) the server pretends it is another day, so we say so.
    document.getElementById("today").textContent = data.demo ? data.today + " (demo)" : data.today;

    const now = [];
    data.festivals.forEach((f) => now.push(f.name));
    data.seasons.forEach((s) => now.push(s.name));
    document.getElementById("occasions").textContent = now.length > 0 ? now.join(", ") : "None";

    const coming = data.upcoming.map((f) => `${f.name} (from ${f.start})`);
    document.getElementById("upcoming").textContent = coming.length > 0 ? coming.join(", ") : "None";
  } catch (err) {
    showMessage(err.message, true);
  }
}

async function loadDashboard(){
  try {
    const report=await api("/api/report?period="+serverToday);
    document.getElementById("today-income").textContent = money(report.total_income);
    document.getElementById("today-profit").textContent=money(report.total_profit);
    document.getElementById("today-items").textContent = report.total_qty;
  }catch(err){
    showMessage(err.message,true);
  }
}

// ---------- Start ----------

// A click outside the menu, or the Escape key, closes it.
document.addEventListener("click", (event) => {
  if (!event.target.closest(".menu")) {
    setMenuOpen(false);
  }
});
document.addEventListener("keydown", (event) => {
  if (event.key === "Escape") {
    setMenuOpen(false);
  }
});
function loadStockInProducts() {
  const sel = document.getElementById("si-product");
  sel.innerHTML = products.map(p =>
    `<option value="${p.id}">${safe(p.name)}</option>`
  ).join("");
}

document.getElementById("stock-in-form").addEventListener("submit", async function(e) {
  e.preventDefault();
  const productID = Number(document.getElementById("si-product").value);
  const qty = Number(document.getElementById("si-qty").value);
  const price = Number(document.getElementById("si-price").value);
  const expiry = document.getElementById("si-expiry").value;
  try {
    await api("/api/purchases", {
      product_id: productID,
      quantity: qty,
      buy_price: price,
      expiry_date: expiry,
    });
    showMessage("Stock added");
    document.getElementById("stock-in-form").reset();
  } catch (err) {
    showMessage(err.message, true);
  }
});


// A click outside the menu, or the Escape key, closes it.
document.addEventListener("click", (event) => {
  if (!event.target.closest(".menu")) {
    setMenuOpen(false);
  }
});
document.addEventListener("keydown", (event) => {
  if (event.key === "Escape") {
    setMenuOpen(false);
  }
});

async function start() {
  await loadOccasions(); // first, because everything below needs serverToday
  try {
    const prods = await api("/api/products");
    products = prods;
    loadSellProducts();
    loadStockInProducts();
    await loadRecentSales();
  } catch (err) {
    showMessage(err.message, true);
  }
  showTab("dashboard");
}
function loadSellProducts() {
  const sel = document.getElementById("sell-product");
  sel.innerHTML = products.map(p =>
    `<option value="${p.id}">${safe(p.name)} (Stock: ${p.stock})</option>`
  ).join("");
  updateSellTotal();
}

function updateSellTotal() {
  const sel = document.getElementById("sell-product");
  const qty = Number(document.getElementById("sell-qty").value) || 0;
  const p = findProduct(sel.value);
  const total = p ? p.sell_price * (1 - p.discount / 100) * qty : 0;
  document.getElementById("sell-total").textContent = "৳" + money(total);
}

document.getElementById("sell-product").addEventListener("change", updateSellTotal);
document.getElementById("sell-qty").addEventListener("input", updateSellTotal);

document.getElementById("sell-form").addEventListener("submit", async function(e) {
  e.preventDefault();
  const productID = Number(document.getElementById("sell-product").value);
  const qty = Number(document.getElementById("sell-qty").value);
  try {
    await api("/api/sales", { product_id: productID, quantity: qty });
    showMessage("Sale saved");
    document.getElementById("sell-form").reset();
    await loadRecentSales();
    products = await api("/api/products");
    loadSellProducts();
  } catch (err) {
    showMessage(err.message, true);
  }
});

async function loadRecentSales() {
  const sales = await api("/api/sales/recent");
  const tbody = document.getElementById("sell-table");
  if (sales.length === 0) {
    tbody.innerHTML = emptyRow(6, "No sales yet");
    return;
  }
  tbody.innerHTML = sales.map(s => `
    <tr>
      <td>${safe(s.product_name)}</td>
      <td>${s.quantity}</td>
      <td>৳${money(s.sell_price)}</td>
      <td>${s.discount}%</td>
      <td>৳${money(s.total)}</td>
      <td>${s.sale_date}</td>
    </tr>
  `).join("");
}

start();
