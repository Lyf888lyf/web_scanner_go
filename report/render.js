let reportData = []; // 全部数据缓存

async function loadReport() {
  const res = await fetch("report.json");
  reportData = await res.json();
  applyFilters(); // 初始化应用筛选器
}

function applyFilters() {
  const keyword = document.getElementById("searchInput").value.trim().toLowerCase();
  const showEmpty = document.getElementById("toggleEmpty").checked;
  const selectedTypes = Array.from(document.querySelectorAll(".type-filter:checked")).map(cb => cb.value);

  const container = document.getElementById("report-container");
  container.innerHTML = "";

  const roots = reportData.filter(r => !r.parent);
  const childMap = {};
  reportData.forEach(r => {
    if (r.parent) {
      if (!childMap[r.parent]) childMap[r.parent] = [];
      childMap[r.parent].push(r);
    }
  });

  roots.forEach((root, idx) => {
    const filtered = filterResourceTypes(root, selectedTypes);
    if (!showEmpty && isEmpty(filtered)) return;
    if (keyword && !JSON.stringify(filtered).toLowerCase().includes(keyword)) return;

    const div = document.createElement("div");
    div.className = "target";
    div.innerHTML = `
      <h2>#${idx + 1} - ${highlight(root.title, keyword)}</h2>
      <p><b>URL:</b> ${highlight(root.url, keyword)} <button onclick="openLink('${root.url}')">访问</button></p>
      <p><b>Status:</b> ${root.status}</p>
      <p><b>时间:</b> ${root.timestamp}</p>
      ${renderToggleSection("HTML", filtered.htmlList, keyword)}
      ${renderToggleSection("JS", filtered.jsList, keyword)}
      ${renderToggleSection("CSS", filtered.cssList, keyword)}
      ${renderToggleSection("Images", filtered.imageList, keyword)}
      ${renderToggleAPISection(filtered.apiList, keyword)}
      ${renderToggleSection("Other", filtered.otherList, keyword)}
    `;

    if (childMap[root.url]) {
      const toggleId = "childs-" + Math.random().toString(36).slice(2);
      const toggle = document.createElement("div");
      toggle.className = "sub-group";
      toggle.innerHTML = `
        <h3 onclick="toggleSection('${toggleId}')">📁 子页面 (${childMap[root.url].length})</h3>
        <div id="${toggleId}" style="display:none;">
          ${childMap[root.url]
          .map(c => renderChild(c, keyword, selectedTypes, showEmpty))
          .filter(html => html).join("")}
        </div>
      `;
      div.appendChild(toggle);
    }

    container.appendChild(div);
  });
}

function filterResourceTypes(data, types) {
  return {
    ...data,
    htmlList: types.includes("htmlList") ? data.htmlList : [],
    jsList: types.includes("jsList") ? data.jsList : [],
    cssList: types.includes("cssList") ? data.cssList : [],
    imageList: types.includes("imageList") ? data.imageList : [],
    apiList: types.includes("apiList") ? data.apiList : [],
    otherList: types.includes("otherList") ? data.otherList : []
  };
}

function renderChild(c, keyword, types, showEmpty) {
  const filtered = filterResourceTypes(c, types);
  if (!showEmpty && isEmpty(filtered)) return "";

  return `
    <div class="target child">
      <h4>${highlight(c.title, keyword)}</h4>
      <p><b>URL:</b> ${highlight(c.url, keyword)} <button onclick="openLink('${c.url}')">访问</button></p>
      <p><b>Status:</b> ${c.status}</p>
      <p><b>时间:</b> ${c.timestamp}</p>
      ${renderToggleSection("HTML", filtered.htmlList, keyword)}
      ${renderToggleSection("JS", filtered.jsList, keyword)}
      ${renderToggleSection("CSS", filtered.cssList, keyword)}
      ${renderToggleSection("Images", filtered.imageList, keyword)}
      ${renderToggleAPISection(filtered.apiList, keyword)}
      ${renderToggleSection("Other", filtered.otherList, keyword)}
    </div>
  `;
}

function renderToggleSection(title, list, keyword) {
  list = list || [];
  if (list.length === 0) return "";
  const id = "section-" + Math.random().toString(36).substr(2, 9);

  return `
    <div class="section">
      <h3 onclick="toggleSection('${id}')">▶ ${title} (${list.length})</h3>
      <ul id="${id}" style="display:none;">
        ${list.map(item => {
    const safeItem = escapeHTML(item);
    return `
            <li>
              <div class="url-line">
                <div class="url-text">${highlight(safeItem, keyword)}</div>
                <div class="btn-group">
                  <button onclick="copyText(\`${item}\`)">复制</button>
                  ${renderLinkActions(item)}
                </div>
              </div>
            </li>`;
  }).join("")}
      </ul>
    </div>
  `;
}

function renderToggleAPISection(list, keyword) {
  list = list || [];
  if (list.length === 0) return "";
  const id = "api-" + Math.random().toString(36).substr(2, 9);

  return `
    <div class="section">
      <h3 onclick="toggleSection('${id}')">▶ API (${list.length})</h3>
      <ul id="${id}" style="display:none;">
        ${list.map(a => {
    const preIdResp = "resp-" + Math.random().toString(36).slice(2);
    const preIdPost = "post-" + Math.random().toString(36).slice(2);

    return `
            <li class="api-entry">
              <div class="api-main-line">
                <code>[${a.method}]</code>
                <div class="url-text">${highlight(a.url, keyword)}</div>
                <div class="btn-group">
                  <button onclick="copyText('${a.url}')">复制</button>
                  <button onclick="openLink('${a.url}')">访问</button>
                </div>
              </div>

              ${a.postData ? `
                <div class="apibox">
                  <b>请求体：</b>
                  <pre id="${preIdPost}">${highlightJSON(a.postData)}</pre>
                  <button class="json-toggle" onclick="toggleJson('${preIdPost}', this)">展开</button>
                </div>` : ""}

              ${a.response ? `
                <div class="apibox">
                  <b>响应体：</b>
                  <pre id="${preIdResp}">${highlightJSON(a.response)}</pre>
                  <button class="json-toggle" onclick="toggleJson('${preIdResp}', this)">展开</button>
                </div>` : ""}
            </li>`;
  }).join("")}
      </ul>
    </div>
  `;
}

function toggleSection(id) {
  const el = document.getElementById(id);
  el.style.display = (el.style.display === "none") ? "block" : "none";
}

function toggleJson(id, btn) {
  const pre = document.getElementById(id);
  const isExpanded = pre.classList.toggle("expanded");
  btn.innerText = isExpanded ? "收起" : "展开";
}

function toggleAllSections(expand) {
  document.querySelectorAll(".section ul, .apibox pre").forEach(el => {
    el.style.display = expand ? "block" : "none";
    if (expand) el.classList.add("expanded");
    else el.classList.remove("expanded");
  });
}

function highlight(str, keyword) {
  if (!keyword) return escapeHTML(str);
  return escapeHTML(str).replace(new RegExp(`(${keyword})`, 'gi'), `<mark>$1</mark>`);
}

function highlightJSON(jsonStr) {
  try {
    return JSON.stringify(JSON.parse(jsonStr), null, 2);
  } catch {
    return escapeHTML(jsonStr);
  }
}

function isEmpty(obj) {
  return (!obj.htmlList?.length && !obj.jsList?.length && !obj.cssList?.length &&
      !obj.imageList?.length && !obj.apiList?.length && !obj.otherList?.length);
}

function copyText(text) {
  navigator.clipboard.writeText(text).then(() => {
    alert("已复制: " + text);
  });
}

function openLink(url) {
  window.open(url, "_blank");
}

function escapeHTML(str) {
  return String(str).replace(/[<>&"]/g, c => ({
    "<": "&lt;", ">": "&gt;", "&": "&amp;", '"': "&quot;"
  }[c]));
}

function exportLinksTxt() {
  const links = [];
  document.querySelectorAll(".url-text").forEach(el => {
    const text = el.textContent.trim();
    if (text) links.push(text);
  });

  const blob = new Blob([links.join("\n")], { type: "text/plain" });
  const link = document.createElement("a");
  link.href = URL.createObjectURL(blob);
  link.download = "links.txt";
  link.click();
}

function toggleDarkMode() {
  document.body.classList.toggle("dark-mode");
}

function renderLinkActions(link) {
  const lower = link.toLowerCase();

  // 🖼️ 图片类
  if (link.startsWith("data:image/") || /\.(jpg|jpeg|png|gif|bmp|svg)$/.test(lower)) {
    return `
      <button onclick="previewDataUrl(\`${link}\`)">预览</button>
      <button onclick="downloadDataUrl(\`${link}\`)">下载</button>
    `;
  }

  // 📄 文本或 JSON
  if (link.startsWith("data:text/") || link.startsWith("data:application/json")) {
    return `
      <button onclick="previewTextData(\`${link}\`)">预览</button>
      <button onclick="downloadDataUrl(\`${link}\`)">下载</button>
    `;
  }

  // 🌐 HTTP 网页
  if (link.startsWith("http://") || link.startsWith("https://")) {
    return `<button onclick="openLink(\`${link}\`)">访问</button>`;
  }

  // 💾 其他未知二进制下载
  if (link.startsWith("data:application/") || link.startsWith("data:binary/")) {
    return `<button onclick="downloadDataUrl(\`${link}\`)">下载</button>`;
  }

  // 🛑 默认仅复制
  return `<span style="font-size:12px;">未知格式</span>`;
}

function previewDataUrl(dataUrl) {
  const w = window.open("");
  w.document.write(`<img src="${dataUrl}" style="max-width:100%;">`);
}

function previewTextData(dataUrl) {
  try {
    const text = decodeURIComponent(dataUrl.split(',')[1]);
    const w = window.open("");
    w.document.write(`<pre style="white-space:pre-wrap;">${escapeHTML(text)}</pre>`);
  } catch {
    alert("预览失败：无法解析文本内容");
  }
}

function downloadDataUrl(dataUrl) {
  const a = document.createElement("a");
  a.href = dataUrl;
  a.download = "downloaded_file"; // 可根据 MIME 类型生成扩展名
  document.body.appendChild(a);
  a.click();
  document.body.removeChild(a);
}

window.onload = () => {
  loadReport();
  document.getElementById("searchInput").addEventListener("input", applyFilters);
  document.querySelectorAll(".type-filter").forEach(cb =>
      cb.addEventListener("change", applyFilters)
  );
};
