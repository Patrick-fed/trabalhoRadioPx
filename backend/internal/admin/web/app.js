/* RadioPX Admin SPA — vanilla JS, token kept in memory only. */
(function () {
  "use strict";

  const API = ""; // same origin
  let token = null;
  let me = null;
  let statsTimer = null;
  let usersCache = [];
  let prevStats = null; // for delta badges (in-memory only)
  let map = null;
  let mapLayer = null;
  let mapMarkerCount = -1; // to re-fit bounds only when count changes

  const $ = (sel) => document.querySelector(sel);

  // ---------- utilities ----------

  function toast(msg, kind) {
    const el = $("#toast");
    el.textContent = msg;
    el.className = "toast " + (kind || "ok");
    el.hidden = false;
    clearTimeout(toast._t);
    toast._t = setTimeout(() => { el.hidden = true; }, 3200);
  }

  async function api(path, options) {
    const opts = options || {};
    const headers = Object.assign({ "Content-Type": "application/json" }, opts.headers || {});
    if (token) headers["Authorization"] = "Bearer " + token;

    let res;
    try {
      res = await fetch(API + path, Object.assign({}, opts, { headers }));
    } catch (e) {
      throw new Error("Não foi possível conectar ao servidor");
    }

    if (res.status === 401) {
      // Token expirado ou inválido → volta para o login.
      logout(false);
      throw new Error("Sessão expirada. Entre novamente.");
    }
    if (res.status === 429) {
      throw new Error("Muitas tentativas. Aguarde um momento.");
    }

    let body = null;
    const text = await res.text();
    if (text) {
      try { body = JSON.parse(text); } catch (_) { body = { error: text }; }
    }

    if (!res.ok) {
      const msg = (body && (body.error || body.message)) || ("Erro " + res.status);
      const err = new Error(msg);
      err.status = res.status;
      throw err;
    }
    return body;
  }

  // ---------- auth ----------

  async function login(email, password) {
    const data = await api("/api/v1/auth/login", {
      method: "POST",
      body: JSON.stringify({ email, password }),
    });
    token = data.token;

    // O /login devolve claims, não o flag is_admin — busca o perfil.
    // O endpoint de stats/users exige admin; se não for, recebe 403 e sai.
    const users = await api("/api/v1/admin/users");
    me = users.find(function (u) { return u.email === email.toLowerCase().trim(); });
    if (!me || !me.is_admin) {
      token = null;
      throw new Error("Acesso negado: apenas administradores");
    }
    return me;
  }

  function logout(callApi) {
    if (callApi !== false) {
      // apenas limpa estado local
    }
    token = null;
    me = null;
    stopStats();
    prevStats = null;
    mapMarkerCount = -1;
    $("#app-view").hidden = true;
    $("#login-view").hidden = false;
    $("#password").value = "";
  }

  // ---------- dashboard ----------

  function startStats() {
    refreshDashboard();
    stopStats();
    statsTimer = setInterval(refreshDashboard, 10000);
  }

  function stopStats() {
    if (statsTimer) { clearInterval(statsTimer); statsTimer = null; }
  }

  function setDelta(id, current, prev) {
    const el = $(id);
    if (!el) return;
    if (prev == null || current === prev) {
      el.hidden = true;
      return;
    }
    const d = current - prev;
    el.textContent = (d > 0 ? "+" : "−") + Math.abs(d);
    el.className = "delta " + (d > 0 ? "up" : "down");
    el.hidden = false;
  }

  function setSub(id, text) {
    const el = $(id);
    if (el) el.textContent = text || "";
  }

  async function refreshDashboard() {
    try {
      const [s, list] = await Promise.all([
        api("/api/v1/admin/stats"),
        api("/api/v1/admin/users"),
      ]);
      usersCache = list;

      const withLoc = list.filter(function (u) {
        return u.last_lat != null && u.last_lng != null;
      }).length;

      $("#stat-users").textContent = s.users_total;
      $("#stat-active").textContent = s.users_active;
      $("#stat-channels").textContent = s.channels_total;
      $("#stat-ws").textContent = s.ws_connections;
      $("#stat-rooms").textContent = s.live_rooms;
      $("#stat-tx").textContent = s.transmitting_channels;
      $("#stat-loc").textContent = withLoc;

      const suspended = s.users_total - s.users_active;
      setSub("#stat-users-sub", suspended > 0 ? suspended + " suspenso" + (suspended > 1 ? "s" : "") : "todos ativos");
      setSub("#stat-active-sub", s.users_total > 0 ? Math.round((s.users_active / s.users_total) * 100) + "% do total" : "sem usuários");
      setSub("#stat-channels-sub", s.channels_total > 0 ? "canais criados" : "nenhum canal");
      setSub("#stat-ws-sub", s.ws_connections > 0 ? "clientes conectados" : "nenhuma conexão");
      setSub("#stat-rooms-sub", s.live_rooms > 0 ? "salas com membros" : "nenhuma sala");
      setSub("#stat-tx-sub", s.transmitting_channels > 0 ? "canais no ar" : "silêncio no ar");
      setSub("#stat-loc-sub", "com posição conhecida");

      if (prevStats) {
        setDelta("#stat-users-delta", s.users_total, prevStats.users_total);
        setDelta("#stat-active-delta", s.users_active, prevStats.users_active);
        setDelta("#stat-channels-delta", s.channels_total, prevStats.channels_total);
        setDelta("#stat-loc-delta", withLoc, prevStats.withLoc);
      }
      prevStats = {
        users_total: s.users_total,
        users_active: s.users_active,
        channels_total: s.channels_total,
        withLoc: withLoc,
      };

      renderMap(list);
    } catch (e) {
      if (e.status !== 401) console.warn("dashboard:", e.message);
    }
  }

  // ---------- map ----------

  function initMap() {
    if (map || typeof L === "undefined") return;
    map = L.map("map", { scrollWheelZoom: true }).setView([-23.5505, -46.6333], 4);
    L.tileLayer("https://tile.openstreetmap.org/{z}/{x}/{y}.png", {
      maxZoom: 19,
      attribution: '&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a>',
    }).addTo(map);
    mapLayer = L.layerGroup().addTo(map);
    // Leaflet serves relative icon URLs from the CSS; with <base href="/admin/">
    // they resolve correctly, but pin them explicitly to the vendored path.
    L.Icon.Default.mergeOptions({
      iconUrl: "vendor/leaflet/images/marker-icon.png",
      iconRetinaUrl: "vendor/leaflet/images/marker-icon-2x.png",
      shadowUrl: "vendor/leaflet/images/marker-shadow.png",
    });
  }

  function popupHTML(u) {
    const name = escapeHTML(u.name || "—");
    const email = escapeHTML(u.email || "—");
    const when = escapeHTML(fmtWhen(u.last_location_at));
    const status = u.is_active
      ? '<span class="badge badge-ok">ativo</span>'
      : '<span class="badge badge-off">suspenso</span>';
    const lat = Number(u.last_lat).toFixed(5);
    const lng = Number(u.last_lng).toFixed(5);
    const osm = "https://www.openstreetmap.org/?mlat=" + lat + "&mlon=" + lng + "#map=15/" + lat + "/" + lng;
    return (
      "<b>" + name + "</b>" +
      email + "<br>" +
      status + " · " + when + "<br>" +
      '<a href="' + osm + '" target="_blank" rel="noopener">abrir no OpenStreetMap</a>'
    );
  }

  function renderMap(list) {
    if (typeof L === "undefined") return;
    initMap();

    const located = list.filter(function (u) {
      return u.last_lat != null && u.last_lng != null;
    });

    $("#map-empty").hidden = located.length > 0;

    mapLayer.clearLayers();
    const bounds = [];
    located.forEach(function (u) {
      const cls = u.is_active ? "map-marker-active" : "map-marker-inactive";
      const icon = L.divIcon({ className: cls, iconSize: [26, 26], iconAnchor: [13, 26], popupAnchor: [0, -24] });
      const m = L.marker([u.last_lat, u.last_lng], { icon: icon }).bindPopup(popupHTML(u));
      mapLayer.addLayer(m);
      bounds.push([u.last_lat, u.last_lng]);
    });

    // Re-fit only on first paint or when the number of markers changes —
    // otherwise the 10s refresh would steal pan/zoom from the admin.
    if (bounds.length && bounds.length !== mapMarkerCount) {
      map.fitBounds(bounds, { padding: [40, 40], maxZoom: 13 });
    }
    mapMarkerCount = bounds.length;
    // Leaflet needs a kick when the container was hidden (tab switch).
    setTimeout(function () { map.invalidateSize(); }, 100);
  }

  // ---------- users ----------

  function fmtWhen(iso) {
    if (!iso) return "—";
    try {
      const d = new Date(iso);
      return d.toLocaleString("pt-BR", { day: "2-digit", month: "2-digit", year: "2-digit", hour: "2-digit", minute: "2-digit" });
    } catch (_) { return iso; }
  }

  function locationHTML(u) {
    if (u.last_lat == null || u.last_lng == null) {
      return '<span class="loc-cell">—</span>';
    }
    const lat = Number(u.last_lat).toFixed(5);
    const lng = Number(u.last_lng).toFixed(5);
    const url = "https://www.openstreetmap.org/?mlat=" + lat + "&mlon=" + lng + "#map=15/" + lat + "/" + lng;
    return (
      '<span class="loc-cell">' + lat + ", " + lng +
      ' <a href="' + url + '" target="_blank" rel="noopener">ver no mapa</a>' +
      "<br><small>" + fmtWhen(u.last_location_at) + "</small></span>"
    );
  }

  function statusBadge(u) {
    return u.is_active
      ? '<span class="badge badge-ok">ativo</span>'
      : '<span class="badge badge-off">suspenso</span>';
  }

  function roleBadge(u) {
    return u.is_admin
      ? '<span class="badge badge-admin">admin</span>'
      : '<span class="badge badge-user">usuário</span>';
  }

  function actionsHTML(u) {
    const isMe = me && u.id === me.id;
    let html = '<div class="actions">';

    if (isMe) {
      html += '<span class="loc-cell">você</span></div>';
      return html;
    }

    if (u.is_active) {
      html += '<button class="btn-tiny warn" data-act="suspend" data-id="' + u.id + '">Suspender</button>';
    } else {
      html += '<button class="btn-tiny" data-act="reactivate" data-id="' + u.id + '">Reativar</button>';
    }

    if (u.is_admin) {
      html += '<button class="btn-tiny" data-act="demote" data-id="' + u.id + '">Remover admin</button>';
    } else {
      html += '<button class="btn-tiny" data-act="promote" data-id="' + u.id + '">Tornar admin</button>';
    }

    html += '<button class="btn-tiny danger" data-act="delete" data-id="' + u.id + '">Excluir</button>';
    html += "</div>";
    return html;
  }

  function renderUsers(list) {
    const tbody = $("#users-tbody");
    const q = ($("#user-search").value || "").toLowerCase().trim();
    const filtered = q
      ? list.filter(function (u) {
          return (u.name || "").toLowerCase().indexOf(q) >= 0 ||
                 (u.email || "").toLowerCase().indexOf(q) >= 0;
        })
      : list;

    if (!filtered.length) {
      tbody.innerHTML = '<tr><td colspan="6" class="empty">Nenhum usuário encontrado</td></tr>';
      return;
    }

    tbody.innerHTML = filtered.map(function (u) {
      return (
        "<tr>" +
        "<td>" + escapeHTML(u.name || "—") + "</td>" +
        "<td>" + escapeHTML(u.email || "—") + "</td>" +
        "<td>" + statusBadge(u) + "</td>" +
        "<td>" + roleBadge(u) + "</td>" +
        "<td>" + locationHTML(u) + "</td>" +
        "<td>" + actionsHTML(u) + "</td>" +
        "</tr>"
      );
    }).join("");
  }

  function escapeHTML(s) {
    return String(s)
      .replace(/&/g, "&amp;")
      .replace(/</g, "&lt;")
      .replace(/>/g, "&gt;")
      .replace(/"/g, "&quot;");
  }

  async function refreshUsers() {
    const errEl = $("#users-error");
    errEl.hidden = true;
    try {
      usersCache = await api("/api/v1/admin/users");
      renderUsers(usersCache);
    } catch (e) {
      if (e.status === 401) return;
      errEl.textContent = e.message;
      errEl.hidden = false;
    }
  }

  async function userAction(act, id) {
    const target = usersCache.find(function (u) { return u.id === id; });
    const name = target ? (target.email || id) : id;

    try {
      if (act === "suspend") {
        if (!confirm("Suspender " + name + "? Ele não poderá mais entrar.")) return;
        await api("/api/v1/admin/users/" + id, { method: "PATCH", body: JSON.stringify({ is_active: false }) });
        toast("Usuário suspenso");
      } else if (act === "reactivate") {
        await api("/api/v1/admin/users/" + id, { method: "PATCH", body: JSON.stringify({ is_active: true }) });
        toast("Usuário reativado");
      } else if (act === "promote") {
        if (!confirm("Promover " + name + " a administrador?")) return;
        await api("/api/v1/admin/users/" + id, { method: "PATCH", body: JSON.stringify({ is_admin: true }) });
        toast("Usuário promovido a admin");
      } else if (act === "demote") {
        if (!confirm("Remover privilégios de admin de " + name + "?")) return;
        await api("/api/v1/admin/users/" + id, { method: "PATCH", body: JSON.stringify({ is_admin: false }) });
        toast("Privilégios de admin removidos");
      } else if (act === "delete") {
        if (!confirm("Excluir definitivamente " + name + "? Essa ação não pode ser desfeita.")) return;
        await api("/api/v1/admin/users/" + id, { method: "DELETE" });
        toast("Usuário excluído");
      }
      await refreshUsers();
      refreshDashboard();
    } catch (e) {
      toast(e.message, "err");
    }
  }

  // ---------- tabs ----------

  function showTab(name) {
    document.querySelectorAll(".nav-btn").forEach(function (b) {
      b.classList.toggle("active", b.dataset.tab === name);
    });
    $("#dashboard-tab").hidden = name !== "dashboard";
    $("#users-tab").hidden = name !== "users";
    if (name === "users") refreshUsers();
    if (name === "dashboard") refreshDashboard();
  }

  // ---------- wiring ----------

  $("#login-form").addEventListener("submit", async function (e) {
    e.preventDefault();
    const btn = $("#login-btn");
    const errEl = $("#login-error");
    btn.disabled = true;
    errEl.hidden = true;
    try {
      const user = await login($("#email").value, $("#password").value);
      $("#admin-email").textContent = user.email;
      $("#login-view").hidden = true;
      $("#app-view").hidden = false;
      showTab("dashboard");
      startStats();
    } catch (err) {
      errEl.textContent = err.message;
      errEl.hidden = false;
    } finally {
      btn.disabled = false;
    }
  });

  $("#logout-btn").addEventListener("click", function () { logout(); });

  document.querySelectorAll(".nav-btn").forEach(function (b) {
    b.addEventListener("click", function () { showTab(b.dataset.tab); });
  });

  $("#refresh-users").addEventListener("click", refreshUsers);
  $("#user-search").addEventListener("input", function () { renderUsers(usersCache); });

  $("#users-tbody").addEventListener("click", function (e) {
    const btn = e.target.closest("button[data-act]");
    if (!btn) return;
    userAction(btn.dataset.act, btn.dataset.id);
  });
})();
