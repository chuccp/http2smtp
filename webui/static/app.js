/* HTTP2SMTP management UI behaviour.
   The pages are server-rendered; this file carries the small bits of interactivity:
   the tab bar, nav highlighting, delete confirmation, token copy and flash hiding. */
(function () {
  'use strict';

  /* ---------- nav highlight ---------- */
  var path = location.pathname;
  document.querySelectorAll('.nav-item[data-path]').forEach(function (el) {
    var target = el.getAttribute('data-path');
    if (path === target || path.indexOf(target + '/') === 0) {
      el.classList.add('active');
    }
  });

  /* ---------- tab bar ----------
     Tabs are a client-side record of visited pages in localStorage; switching is a
     normal navigation (the pages re-render server-side anyway). The dashboard tab is
     always present and cannot be closed. Form pages and the password page never enter
     the bar, so an unsaved form is not a tab someone can lose data through. */
  var TAB_KEY = 'http2smtp-tabs';
  var FIXED = [{ path: '/dashboard', title: 'Dashboard' }];
  var EXCLUDE = [/^\/(login|setup|password)$/, /\/form(\/|$)/, /^\/logs\/\d+$/];

  function loadTabs() {
    try {
      var raw = localStorage.getItem(TAB_KEY);
      var arr = raw ? JSON.parse(raw) : [];
      return Array.isArray(arr) ? arr : [];
    } catch (e) {
      return [];
    }
  }

  function saveTabs(tabs) {
    try { localStorage.setItem(TAB_KEY, JSON.stringify(tabs)); } catch (e) { /* private mode */ }
  }

  function currentTab() {
    return { path: location.pathname + location.search, title: document.title.replace(/\s*-\s*HTTP2SMTP\s*$/, '') };
  }

  function excluded(p) {
    var bare = p.split('?')[0];
    return EXCLUDE.some(function (re) { return re.test(bare); });
  }

  var tabbar = document.getElementById('tabbar');
  if (tabbar) {
    var cur = currentTab();
    if (!excluded(cur.path)) {
      var tabs = loadTabs().filter(function (t) { return t && t.path && !excluded(t.path) && t.path.indexOf('/dashboard') !== 0; });
      if (!tabs.some(function (t) { return t.path === cur.path; })) {
        tabs.push(cur);
      }
      saveTabs(tabs);
    }

    function renderTabs() {
      var open = loadTabs().filter(function (t) { return t && t.path && !excluded(t.path) && t.path.indexOf('/dashboard') !== 0; });
      tabbar.innerHTML = '';
      FIXED.concat(open).forEach(function (t, idx) {
        var a = document.createElement('a');
        a.className = 'tab' + (t.path === cur.path ? ' active' : '');
        a.href = t.path;
        a.textContent = t.title || t.path;
        if (idx >= FIXED.length) {
          var btn = document.createElement('button');
          btn.type = 'button';
          btn.className = 'tab-close';
          btn.textContent = '×';
          btn.title = '×';
          btn.addEventListener('click', function (ev) {
            ev.preventDefault();
            ev.stopPropagation();
            saveTabs(loadTabs().filter(function (o) { return o.path !== t.path; }));
            renderTabs();
          });
          a.appendChild(btn);
        }
        tabbar.appendChild(a);
      });
    }
    renderTabs();

    /* Right-click menu: close others / close all, mirroring the old SPA's tab menu. */
    var menu = null;
    function closeMenu() { if (menu) { menu.remove(); menu = null; } }
    tabbar.addEventListener('contextmenu', function (ev) {
      var tab = ev.target.closest('.tab');
      if (!tab || !tab.getAttribute('href')) return;
      ev.preventDefault();
      closeMenu();
      var href = tab.getAttribute('href');
      menu = document.createElement('div');
      menu.className = 'tab-menu';
      menu.style.left = ev.clientX + 'px';
      menu.style.top = ev.clientY + 'px';
      function mk(label, fn) {
        var b = document.createElement('button');
        b.type = 'button';
        b.textContent = label;
        b.addEventListener('click', function () { closeMenu(); fn(); });
        menu.appendChild(b);
      }
      mk(tabbar.getAttribute('data-close-others') || 'Close Others', function () {
        saveTabs(loadTabs().filter(function (o) { return o.path === href || o.path.indexOf('/dashboard') === 0; }));
        renderTabs();
      });
      mk(tabbar.getAttribute('data-close-all') || 'Close All', function () {
        saveTabs([]);
        renderTabs();
      });
      document.body.appendChild(menu);
    });
    document.addEventListener('click', closeMenu);
    document.addEventListener('scroll', closeMenu, true);
  }

  /* ---------- delete confirmation ----------
     Row deletes call the existing JSON API (same session cookie) and reload, so the
     page always shows the post-delete state including the current page number. */
  document.addEventListener('click', function (ev) {
    var el = ev.target.closest('[data-href]');
    if (!el) return;
    ev.preventDefault();
    var confirmMsg = el.getAttribute('data-confirm');
    if (confirmMsg && !window.confirm(confirmMsg)) return;
    var href = el.getAttribute('data-href');
    var method = el.getAttribute('data-method') || 'DELETE';
    fetch(href, { method: method, credentials: 'same-origin' })
      .then(function (resp) {
        if (resp.status === 401 || resp.status === 403) {
          location.href = '/login';
          return;
        }
        return resp.json().then(function (body) {
          if (body && (body.code === 0 || body.code === 200)) {
            var flash = el.getAttribute('data-flash');
            if (flash) window.alert(flash);
            location.reload();
          } else {
            window.alert((body && body.msg) || 'error');
          }
        });
      })
      .catch(function () { window.alert('request failed'); });
  });

  /* ---------- copy to clipboard ---------- */
  document.addEventListener('click', function (ev) {
    var el = ev.target.closest('[data-copy]');
    if (!el) return;
    var text = el.getAttribute('data-copy');
    var ok = el.getAttribute('data-ok') || 'OK';
    function flash() {
      var old = el.textContent;
      el.textContent = ok;
      setTimeout(function () { el.textContent = old; }, 1500);
    }
    if (navigator.clipboard && navigator.clipboard.writeText) {
      navigator.clipboard.writeText(text).then(flash, function () {});
    }
  });

  /* ---------- flash auto-hide ---------- */
  document.querySelectorAll('.flash-ok').forEach(function (el) {
    setTimeout(function () {
      el.style.transition = 'opacity .4s';
      el.style.opacity = '0';
      setTimeout(function () { el.remove(); }, 400);
    }, 4000);
  });
})();
