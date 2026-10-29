/* HTTP2SMTP management UI behaviour.
   The pages are server-rendered; this file wires the shell to the ui.js component
   library: sidebar collapse, tab bar, nav highlighting, breadcrumb, delete
   confirmation, password visibility, token copy and token view. Partial navigation (pjax),
   dialogs, toasts and the loading mask live in ui.js. */
(function () {
  'use strict';

  /* ---------- sidebar ----------
     Desktop collapses to an icon rail (remembered in localStorage); on mobile the
     sidebar slides in over the content with a backdrop, like the old Vue layout. */
  var shell = document.getElementById('app-shell');
  var toggle = document.getElementById('sidebar-toggle');
  var overlay = document.getElementById('sidebar-overlay');
  var COLLAPSE_KEY = 'http2smtp-sidebar-collapsed';

  function isMobile() { return window.innerWidth < 768; }
  function collapsedOnLoad() {
    try { return localStorage.getItem(COLLAPSE_KEY) === '1'; } catch (e) { return false; }
  }
  function syncShell() {
    if (!shell) return;
    if (isMobile()) {
      shell.classList.remove('sidebar-collapsed');
    } else {
      shell.classList.remove('sidebar-open');
      shell.classList.toggle('sidebar-collapsed', collapsedOnLoad());
    }
  }
  syncShell();
  window.addEventListener('resize', syncShell);

  if (toggle) {
    toggle.addEventListener('click', function () {
      if (!shell) return;
      if (isMobile()) {
        shell.classList.toggle('sidebar-open');
        return;
      }
      var collapsed = shell.classList.toggle('sidebar-collapsed');
      try { localStorage.setItem(COLLAPSE_KEY, collapsed ? '1' : '0'); } catch (e) { /* private mode */ }
    });
  }
  if (overlay) {
    overlay.addEventListener('click', function () { shell.classList.remove('sidebar-open'); });
  }
  document.querySelectorAll('.nav-item').forEach(function (a) {
    a.addEventListener('click', function () {
      if (isMobile() && shell) shell.classList.remove('sidebar-open');
    });
  });

  /* ---------- nav highlight ---------- */
  function markNav() {
    var path = location.pathname;
    document.querySelectorAll('.nav-item[data-path]').forEach(function (el) {
      var target = el.getAttribute('data-path');
      var on = path === target || path.indexOf(target + '/') === 0;
      el.classList.toggle('active', on);
    });
  }
  markNav();

  /* ---------- header ---------- */
  // The header is outside the swapped content, so the breadcrumb and the language
  // form's return target have to be updated from the new location on every pjax.
  var breadcrumb = document.getElementById('breadcrumb');
  function syncHeader() {
    if (breadcrumb) {
      var title = document.title.replace(/\s*-\s*HTTP2SMTP\s*$/, '');
      breadcrumb.innerHTML = '';
      var span = document.createElement('span');
      span.textContent = title;
      breadcrumb.appendChild(span);
    }
    var back = document.querySelector('.lang-form input[name=back]');
    if (back) back.value = location.pathname + location.search;
  }
  syncHeader();

  /* ---------- tab bar ----------
     Tabs are a client-side record of visited pages in localStorage, keyed by path so
     a filter change does not open a second tab (the first URL is kept, as the Vue
     version did). Switching a tab is a partial navigation (ui.js), so the bar and the
     sidebar stay put. The dashboard tab is always present and cannot be closed. Form
     pages and the password page never enter the bar. */
  var TAB_KEY = 'http2smtp-tabs';
  var tabbar = document.getElementById('tabbar');
  var HOME = {
    path: '/dashboard',
    fullPath: '/dashboard',
    title: (tabbar && tabbar.getAttribute('data-home-title')) || 'Dashboard'
  };
  var EXCLUDE = [/^\/(login|setup|password)$/, /\/form(\/|$)/, /\/send(\/|$)/, /^\/logs\/\d+$/];

  function loadTabs() {
    try {
      var raw = localStorage.getItem(TAB_KEY);
      var arr = raw ? JSON.parse(raw) : [];
      var seen = {};
      return (Array.isArray(arr) ? arr : []).reduce(function (out, t) {
        if (!t || !t.path) return out;
        var path = String(t.path).split('?')[0];
        if (path === HOME.path || seen[path]) return out;
        seen[path] = true;
        out.push({ path: path, fullPath: t.fullPath || t.path, title: t.title || path });
        return out;
      }, []);
    } catch (e) {
      return [];
    }
  }

  function saveTabs(tabs) {
    try { localStorage.setItem(TAB_KEY, JSON.stringify(tabs)); } catch (e) { /* private mode */ }
  }

  function currentTab() {
    return {
      path: location.pathname,
      fullPath: location.pathname + location.search,
      title: document.title.replace(/\s*-\s*HTTP2SMTP\s*$/, '')
    };
  }

  function excluded(p) {
    var bare = p.split('?')[0];
    return EXCLUDE.some(function (re) { return re.test(bare); });
  }

  var cur = currentTab();

  function rememberTab() {
    if (!tabbar || cur.path === HOME.path || excluded(cur.path)) return;
    var tabs = loadTabs();
    var existing = null;
    for (var i = 0; i < tabs.length; i++) {
      if (tabs[i].path === cur.path) { existing = tabs[i]; break; }
    }
    if (existing) {
      // The stored title belongs to the language it was recorded in; refresh it so the
      // fallback below never resurrects the old one after a language switch.
      existing.title = cur.title;
    } else {
      tabs.push(cur);
    }
    saveTabs(tabs);
  }

  function allTabs() {
    return [HOME].concat(loadTabs());
  }

  /* Labels come from the sidebar, which the server rendered in the current language;
     the title stored with the tab may predate a language switch. Pages without a nav
     entry keep their stored title. */
  function navLabel(path) {
    var items = document.querySelectorAll('.nav-item[data-path]');
    for (var i = 0; i < items.length; i++) {
      if (items[i].getAttribute('data-path') !== path) continue;
      var label = items[i].querySelector('.nav-label');
      if (label) return label.textContent.trim();
    }
    return '';
  }

  function renderTabs() {
    if (!tabbar) return;
    tabbar.innerHTML = '';
    allTabs().forEach(function (t, idx) {
      var a = document.createElement('a');
      a.className = 'tab' + (t.path === location.pathname ? ' active' : '');
      a.href = t.fullPath;
      a.setAttribute('data-path', t.path);
      a.textContent = navLabel(t.path) || t.title || t.path;
      if (idx > 0) {
        var btn = document.createElement('button');
        btn.type = 'button';
        btn.className = 'tab-close';
        btn.textContent = '×';
        btn.title = '×';
        btn.addEventListener('click', function (ev) {
          ev.preventDefault();
          ev.stopPropagation();
          closeTab(t);
        });
        a.appendChild(btn);
      }
      tabbar.appendChild(a);
    });
  }

  // Closing the active tab lands on its neighbour, as the Vue tabs did.
  function closeTab(t) {
    var tabs = allTabs();
    var idx = tabs.findIndex(function (o) { return o.path === t.path; });
    saveTabs(loadTabs().filter(function (o) { return o.path !== t.path; }));
    renderTabs();
    if (t.path === location.pathname) {
      var next = tabs[idx + 1] || tabs[idx - 1] || HOME;
      if (next) UI.navigate(next.fullPath, true);
    }
  }

  if (tabbar) {
    rememberTab();
    renderTabs();

    /* Close-others / close-all menu, shared by right-click and the tab-bar chevron. */
    var menu = null;
    function closeMenu() { if (menu) { menu.remove(); menu = null; } }
    function openTabMenu(x, y, path, fullPath) {
      closeMenu();
      menu = document.createElement('div');
      menu.className = 'tab-menu';
      menu.style.left = x + 'px';
      menu.style.top = y + 'px';
      function mk(label, fn) {
        var b = document.createElement('button');
        b.type = 'button';
        b.textContent = label;
        b.addEventListener('click', function () { closeMenu(); fn(); });
        menu.appendChild(b);
      }
      mk(tabbar.getAttribute('data-close-others') || 'Close Others', function () {
        saveTabs(loadTabs().filter(function (o) { return o.path === path; }));
        renderTabs();
        if (location.pathname !== path && path !== HOME.path) UI.navigate(fullPath, true);
      });
      mk(tabbar.getAttribute('data-close-all') || 'Close All', function () {
        saveTabs([]);
        renderTabs();
        if (location.pathname !== HOME.path) UI.navigate(HOME.fullPath, true);
      });
      document.body.appendChild(menu);
    }
    tabbar.addEventListener('contextmenu', function (ev) {
      var tab = ev.target.closest('.tab');
      if (!tab || !tab.getAttribute('href')) return;
      ev.preventDefault();
      openTabMenu(ev.clientX, ev.clientY, tab.getAttribute('data-path'), tab.getAttribute('href'));
    });
    var actions = document.getElementById('tab-actions');
    if (actions) {
      actions.addEventListener('click', function (ev) {
        ev.stopPropagation();
        var r = actions.getBoundingClientRect();
        openTabMenu(Math.max(8, r.right - 140), r.bottom + 4,
          location.pathname, location.pathname + location.search);
      });
    }
    document.addEventListener('click', closeMenu);
    document.addEventListener('scroll', closeMenu, true);
  }

  /* Partial navigation (ui.js) swaps .content and leaves the shell in place, so the
     nav highlight, tab bar and header have to follow the new location themselves. */
  UI.onNavigate(function () {
    cur = currentTab();
    markNav();
    rememberTab();
    renderTabs();
    syncHeader();
    if (isMobile() && shell) shell.classList.remove('sidebar-open');
  });

  /* ---------- logout confirmation ---------- */
  document.addEventListener('submit', function (ev) {
    var form = ev.target;
    if (!form.matches || !form.matches('form[data-confirm]')) return;
    ev.preventDefault();
    UI.confirm(form.getAttribute('data-confirm')).then(function (ok) {
      if (ok) form.submit(); // bypasses the submit event, so this handler does not loop
    });
  });

  /* ---------- delete / trigger actions ----------
     Row actions call the existing JSON API (same session cookie) and then refresh
     only the content area, so the shell survives and the page number is kept. */
  document.addEventListener('click', function (ev) {
    var el = ev.target.closest('[data-href]');
    if (!el) return;
    ev.preventDefault();
    var href = el.getAttribute('data-href');
    var method = el.getAttribute('data-method') || 'DELETE';
    var confirmMsg = el.getAttribute('data-confirm');
    var proceed = confirmMsg ? UI.confirm(confirmMsg) : Promise.resolve(true);
    proceed.then(function (ok) {
      if (!ok) return;
      fetch(href, { method: method, credentials: 'same-origin' })
        .then(function (resp) {
          if (resp.status === 401 || resp.status === 403) {
            location.href = '/login';
            return;
          }
          return resp.json().then(function (body) {
            if (body && (body.code === 0 || body.code === 200)) {
              // data-alert is for actions whose outcome deserves a dialog (trigger now);
              // plain deletes keep the lighter toast.
              var alertMsg = el.getAttribute('data-alert');
              var flash = el.getAttribute('data-flash');
              if (alertMsg) UI.alert(alertMsg);
              else if (flash) UI.toast(flash);
              UI.refresh();
            } else {
              UI.alert((body && body.msg) || 'error');
            }
          });
        })
        .catch(function () { UI.alert('request failed'); });
    });
  });

  /* ---------- list refresh ----------
     data-refresh re-fetches the current URL through the partial-navigation path, so
     the shell, the filters and the tab bar stay put and only the list is replaced. */
  document.addEventListener('click', function (ev) {
    var el = ev.target.closest('[data-refresh]');
    if (!el) return;
    ev.preventDefault();
    UI.refresh();
  });

  /* ---------- filter checkboxes ----------
     A checkbox marked data-auto-submit applies immediately: requestSubmit fires the
     regular submit event, so ui.js still turns the GET filter into a partial
     navigation instead of a full reload. */
  document.addEventListener('change', function (ev) {
    var input = ev.target.closest('[data-auto-submit]');
    if (!input || !input.form) return;
    if (typeof input.form.requestSubmit === 'function') input.form.requestSubmit();
    else input.form.submit();
  });

  /* ---------- password visibility ----------
     The eye button inside .password-wrap flips the input between password and text.
     The labels come from data attributes so the announcement stays in the page's
     language; aria-pressed drives which of the two icons CSS shows. */
  document.addEventListener('click', function (ev) {
    var btn = ev.target.closest('[data-password-toggle]');
    if (!btn) return;
    var box = btn.closest('.password-wrap');
    var input = box && box.querySelector('input');
    if (!input) return;
    var show = input.type === 'password';
    input.type = show ? 'text' : 'password';
    btn.setAttribute('aria-pressed', show ? 'true' : 'false');
    var label = btn.getAttribute(show ? 'data-hide-label' : 'data-show-label');
    if (label) {
      btn.setAttribute('aria-label', label);
      btn.title = label;
    }
  });

  /* ---------- login captcha ----------
     Clicking Login opens a modal with a fresh challenge from /login/captcha. The form
     only submits after the shapes have been clicked in order: the dialog's confirm sets
     the hidden answer fields and lets the submit through. Cancelling consumes nothing. */
  (function () {
    var form = document.querySelector('.login-form');
    if (!form) return;
    var idInput = form.querySelector('[data-captcha-id]');
    var xsInput = form.querySelector('[data-captcha-xs]');
    var ysInput = form.querySelector('[data-captcha-ys]');
    if (!idInput || !xsInput || !ysInput) return;
    var usernameInput = form.querySelector('[name=username]');
    var passwordInput = form.querySelector('[name=password]');
    var verified = false;
    var cancelLabel = document.body.getAttribute('data-ui-cancel') || 'Cancel';

    function attr(name, fallback) {
      var value = form.getAttribute(name);
      return value == null || value === '' ? fallback : value;
    }

    // The click order is shown as the shapes themselves, not translated names, so the
    // sequence reads the same in every language. Labels only feed aria-label/title.
    function shapeNode(shape) {
      var ns = 'http://www.w3.org/2000/svg';
      var node;
      if (shape === 'circle') {
        node = document.createElementNS(ns, 'circle');
        node.setAttribute('cx', '12');
        node.setAttribute('cy', '12');
        node.setAttribute('r', '8.5');
      } else if (shape === 'triangle') {
        node = document.createElementNS(ns, 'path');
        node.setAttribute('d', 'M12 3.8 21 20.2H3Z');
      } else if (shape === 'heart') {
        node = document.createElementNS(ns, 'path');
        node.setAttribute('d', 'M12 21.2 4.2 13.7C1.9 11.4 2.1 7.8 4.5 5.7 6.8 3.7 10 4.1 12 6.4 14 4.1 17.2 3.7 19.5 5.7 21.9 7.8 22.1 11.4 19.8 13.7Z');
      } else if (shape === 'arrow') {
        node = document.createElementNS(ns, 'path');
        node.setAttribute('d', 'M12 2.8 18 8.6h-3.5v6.6H18L12 21.2 6 15.2h3.5V8.6H6Z');
      } else {
        node = document.createElementNS(ns, 'path');
        node.setAttribute('d', 'M12 3.2 14.6 8.9 20.8 9.6 16.2 13.9 17.5 20 12 16.9 6.5 20 7.8 13.9 3.2 9.6 9.4 8.9Z');
      }
      return node;
    }

    function renderPrompt(target, data) {
      target.textContent = '';
      var shapes = data.shapes || [];
      if (!shapes.length) {
        target.textContent = data.prompt || '';
        return;
      }
      var labels = data.labels || [];
      var ns = 'http://www.w3.org/2000/svg';
      shapes.forEach(function (shape, i) {
        if (i) {
          var arrow = document.createElement('span');
          arrow.className = 'login-captcha-arrow';
          arrow.textContent = '→';
          target.appendChild(arrow);
        }
        var label = labels[i] || shape;
        var icon = document.createElementNS(ns, 'svg');
        icon.setAttribute('viewBox', '0 0 24 24');
        icon.setAttribute('class', 'login-captcha-icon');
        icon.setAttribute('role', 'img');
        icon.setAttribute('aria-label', label);
        icon.appendChild(shapeNode(shape));
        var title = document.createElementNS(ns, 'title');
        title.textContent = label;
        icon.appendChild(title);
        target.appendChild(icon);
      });
    }

    form.addEventListener('submit', function (event) {
      if (verified) return;
      event.preventDefault();
      if (form.reportValidity && !form.reportValidity()) return;
      openCaptcha();
    });

    function openCaptcha() {
      var cancelled = false;
      var autoTimer = null;
      // The challenge is bound to these exact values; the overlay blocks the form, so
      // they cannot change while the dialog is open.
      var username = usernameInput ? usernameInput.value : '';
      var password = passwordInput ? passwordInput.value : '';
      var dlg = UI.modal({
        title: attr('data-captcha-title', ''),
        width: '360px',
        cancelValue: null,
        onClose: function () {
          cancelled = true;
          clearTimeout(autoTimer);
        }
      });
      var box = document.createElement('div');
      box.className = 'login-captcha';
      box.innerHTML =
        '<div class="login-captcha-head"><span><span class="login-captcha-prefix"></span> ' +
        '<span class="login-captcha-order" data-order></span></span>' +
        '<button type="button" class="login-captcha-refresh" data-refresh></button></div>' +
        '<div class="login-captcha-stage" data-stage>' +
        '<img alt="" draggable="false" data-image>' +
        '<div class="login-captcha-marks" data-marks></div></div>' +
        '<div class="login-captcha-status" data-status></div>';
      var orderEl = box.querySelector('[data-order]');
      var refresh = box.querySelector('[data-refresh]');
      var stage = box.querySelector('[data-stage]');
      var image = box.querySelector('[data-image]');
      var marks = box.querySelector('[data-marks]');
      var status = box.querySelector('[data-status]');
      box.querySelector('.login-captcha-prefix').textContent = attr('data-captcha-prompt-label', '');
      refresh.textContent = attr('data-captcha-refresh-label', 'Refresh');

      dlg.setBody(box);
      var cancel = dlg.addButton(cancelLabel, false, function () { dlg.close(); });

      var challengeId = '';
      var width = 300;
      var height = 200;
      var total = 3;
      var points = [];

      function submitAnswer() {
        if (!challengeId || points.length !== total) return;
        idInput.value = challengeId;
        xsInput.value = points.map(function (p) { return p.x; }).join(',');
        ysInput.value = points.map(function (p) { return p.y; }).join(',');
        verified = true;
        dlg.close();
        if (typeof form.requestSubmit === 'function') form.requestSubmit();
        else form.submit();
      }
      function progressText(done) {
        return attr('data-captcha-progress', '{done}/{total}')
          .replace('{done}', done)
          .replace('{total}', total);
      }
      function render() {
        if (challengeId && points.length === total) {
          status.textContent = attr('data-captcha-done', '');
          // The last click submits on its own; the short delay lets the marker land.
          if (!autoTimer) {
            autoTimer = setTimeout(function () {
              if (!cancelled) submitAnswer();
            }, 300);
          }
        } else {
          status.textContent = progressText(points.length);
        }
      }
      function reset() {
        clearTimeout(autoTimer);
        autoTimer = null;
        points = [];
        marks.textContent = '';
        render();
      }
      function load() {
        refresh.disabled = true;
        status.textContent = '';
        fetch('/login/captcha', {
          method: 'POST',
          credentials: 'same-origin',
          headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
          body: new URLSearchParams({ username: username, password: password }).toString()
        })
          .then(function (response) { return response.json(); })
          .then(function (payload) {
            var data = payload && payload.data;
            if (!data || !data.id || !data.image) throw new Error('bad captcha');
            challengeId = data.id;
            width = Number(data.width) || width;
            height = Number(data.height) || height;
            total = Number(data.count) || total;
            image.src = data.image;
            renderPrompt(orderEl, data);
            reset();
          })
          .catch(function () {
            challengeId = '';
            status.textContent = attr('data-captcha-failed', '');
          })
          .then(function () { refresh.disabled = false; });
      }

      stage.addEventListener('click', function (event) {
        if (!challengeId || points.length >= total) return;
        var rect = image.getBoundingClientRect();
        if (!rect.width || !rect.height) return;
        var x = Math.round((event.clientX - rect.left) * width / rect.width);
        var y = Math.round((event.clientY - rect.top) * height / rect.height);
        if (x < 0 || y < 0 || x > width || y > height) return;
        points.push({ x: x, y: y });
        var mark = document.createElement('span');
        mark.className = 'login-captcha-mark';
        mark.textContent = String(points.length);
        mark.style.left = (x / width * 100) + '%';
        mark.style.top = (y / height * 100) + '%';
        marks.appendChild(mark);
        render();
      });
      refresh.addEventListener('click', load);
      load();
      dlg.focus(cancel);
    }
  })();

  /* ---------- copy to clipboard ----------
     navigator.clipboard only exists in a secure context; the management UI is
     normally reached over plain http on a LAN address, so fall back to the old
     textarea + execCommand path instead of silently doing nothing. */
  function legacyCopy(text) {
    var area = document.createElement('textarea');
    area.value = text;
    area.style.position = 'fixed';
    area.style.opacity = '0';
    document.body.appendChild(area);
    area.select();
    var ok = false;
    try { ok = document.execCommand('copy'); } catch (e) { ok = false; }
    area.remove();
    return ok;
  }

  document.addEventListener('click', function (ev) {
    var el = ev.target.closest('[data-copy], [data-copy-input]');
    if (!el) return;
    var text = el.getAttribute('data-copy');
    if (el.hasAttribute('data-copy-input')) {
      var wrap = el.closest('.copy-wrap');
      var input = wrap && wrap.querySelector('input');
      text = input ? input.value : '';
    }
    var ok = el.getAttribute('data-ok') || 'OK';
    function flash() {
      // Remember the original markup once: the field button is an icon, so restoring
      // textContent (the old implementation) would erase it, and a second click while
      // the checkmark shows must not capture the checkmark as the original.
      if (el._copyOriginal == null) el._copyOriginal = el.innerHTML;
      el.textContent = ok;
      clearTimeout(el._copyTimer);
      el._copyTimer = setTimeout(function () {
        el.innerHTML = el._copyOriginal;
      }, 1500);
    }
    function failed() {
      UI.alert(el.getAttribute('data-fail') || 'copy failed');
    }
    if (navigator.clipboard && navigator.clipboard.writeText) {
      navigator.clipboard.writeText(text).then(flash, function () {
        if (legacyCopy(text)) flash(); else failed();
      });
    } else if (legacyCopy(text)) {
      flash();
    } else {
      failed();
    }
  });

  /* ---------- view a truncated value ----------
     Buttons carry the full value in data-view (the list renders only a prefix); the
     dialog body uses textContent, so the value is shown verbatim and stays selectable. */
  document.addEventListener('click', function (ev) {
    var el = ev.target.closest('[data-view]');
    if (!el) return;
    UI.alert(el.getAttribute('data-view'), el.getAttribute('data-view-title') || '');
  });
})();
