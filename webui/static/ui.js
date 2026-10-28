/* HTTP2SMTP UI component library — hand-rolled, no external dependencies.
   The pieces every page shares: modal dialogs (alert/confirm/prompt, forms loaded
   from server-rendered pages, read-only page dialogs), toast messages, a loading mask
   and partial navigation (pjax — the EasyUI/Vue-style shell behaviour where only the
   content area changes). Shell-specific wiring (nav highlight, tab bar, sidebar) stays
   in app.js and subscribes through UI.onNavigate.

   Dialog button labels come from data-ui-ok / data-ui-cancel on <body>, so they speak
   the page's current language; English is only the fallback. */
window.UI = (function () {
  'use strict';

  function label(name, fallback) {
    var el = document.querySelector('[data-ui-' + name + ']');
    return (el && el.getAttribute('data-ui-' + name)) || fallback;
  }

  function el(tag, cls) {
    var e = document.createElement(tag);
    if (cls) e.className = cls;
    return e;
  }

  function parsePage(text) {
    return new DOMParser().parseFromString(text, 'text/html');
  }

  function pageTitle(doc) {
    var heading = doc.querySelector('.page-head h1');
    if (heading) {
      var text = heading.textContent.trim();
      if (text) return text;
    }
    return (doc.title || '').replace(/\s*-\s*HTTP2SMTP\s*$/, '');
  }

  /* ---------- toast ---------- */
  var toastBox = null;
  function toast(message, kind) {
    if (!toastBox) {
      toastBox = el('div', 'ui-toasts');
      document.body.appendChild(toastBox);
    }
    var t = el('div', 'ui-toast ui-toast-' + (kind || 'ok'));
    t.textContent = message;
    toastBox.appendChild(t);
    setTimeout(function () {
      t.classList.add('out');
      setTimeout(function () { t.remove(); }, 300);
    }, 3200);
  }

  /* ---------- modal ---------- */
  var modalStack = [];
  // A dialog shell: title bar with ×, body, footer. Returns a handle the dialog
  // builders below (and page scripts) drive. opts: title, width, className,
  // escClose, cancelValue, onClose.
  function modal(opts) {
    opts = opts || {};
    var overlay = el('div', 'ui-dialog-mask');
    var box = el('div', 'ui-dialog' + (opts.className ? ' ' + opts.className : ''));
    if (opts.width) box.style.width = opts.width;
    box.setAttribute('role', 'dialog');
    box.setAttribute('aria-modal', 'true');

    var head = el('div', 'ui-dialog-head');
    var titleEl = el('span', 'ui-dialog-title');
    titleEl.textContent = opts.title || '';
    var closeBtn = el('button', 'ui-dialog-close');
    closeBtn.type = 'button';
    closeBtn.innerHTML = '&times;';
    closeBtn.setAttribute('aria-label', label('close', 'Close'));
    head.appendChild(titleEl);
    head.appendChild(closeBtn);

    var body = el('div', 'ui-dialog-body');
    var foot = el('div', 'ui-dialog-foot');
    var loading = el('div', 'ui-dialog-loading');
    loading.appendChild(el('div', 'ui-spinner'));

    if (opts.title) box.appendChild(head);
    box.appendChild(body);
    box.appendChild(foot);

    var closed = false;
    function close(value) {
      if (closed) return;
      closed = true;
      document.removeEventListener('keydown', onKey, true);
      overlay.remove();
      var at = modalStack.indexOf(handle);
      if (at >= 0) modalStack.splice(at, 1);
      if (opts.onClose) opts.onClose(value);
    }
    function onKey(ev) {
      // Only the topmost dialog reacts to Esc; stacked pickers must not close the
      // form dialog underneath them as well.
      if (modalStack[modalStack.length - 1] !== handle) return;
      if (ev.key === 'Escape' && opts.escClose !== false) {
        ev.preventDefault();
        close(opts.cancelValue);
      }
    }
    closeBtn.addEventListener('click', function () { close(opts.cancelValue); });
    document.addEventListener('keydown', onKey, true);
    overlay.addEventListener('click', function (ev) {
      if (ev.target === overlay && opts.backdropClose) close(opts.cancelValue);
    });

    var handle = {
      root: overlay,
      box: box,
      body: body,
      footer: foot,
      close: close,
      setTitle: function (text) {
        titleEl.textContent = text || '';
        if (!head.parentNode && (text || opts.showClose)) {
          box.insertBefore(head, box.firstChild);
        }
      },
      setBody: function (node) {
        body.innerHTML = '';
        if (typeof node === 'string') body.innerHTML = node;
        else if (node) body.appendChild(node);
      },
      setLoading: function (on) {
        body.classList.toggle('is-loading', !!on);
        if (on) body.appendChild(loading);
        else loading.remove();
      },
      addButton: function (text, primary, fn) {
        var b = el('button', 'btn' + (primary ? ' btn-primary' : ''));
        b.type = 'button';
        b.textContent = text;
        b.addEventListener('click', fn);
        foot.appendChild(b);
        return b;
      }
    };
    overlay.appendChild(box);
    document.body.appendChild(overlay);
    modalStack.push(handle);
    handle.focus = function (target) {
      (target || foot.querySelector('.btn-primary') || closeBtn).focus();
    };
    return handle;
  }

  /* ---------- message dialogs ---------- */
  function openMessage(opts) {
    return new Promise(function (resolve) {
      var kind = opts.kind || 'alert';
      var dlg = modal({
        title: opts.title,
        className: 'ui-dialog-message',
        cancelValue: kind === 'confirm' ? false : null,
        onClose: resolve
      });
      var input = null;
      if (kind === 'prompt') {
        if (opts.message) {
          var msg = el('div', 'ui-dialog-msg');
          msg.textContent = opts.message;
          dlg.body.appendChild(msg);
        }
        input = el('input', 'ui-dialog-input');
        input.type = 'text';
        input.value = opts.value == null ? '' : opts.value;
        dlg.body.appendChild(input);
      } else {
        dlg.body.textContent = opts.message || '';
      }

      if (kind !== 'alert') {
        dlg.addButton(label('cancel', 'Cancel'), false, function () {
          dlg.close(kind === 'confirm' ? false : null);
        });
      }
      var ok = dlg.addButton(label('ok', 'OK'), true, function () {
        dlg.close(kind === 'prompt' ? input.value : kind === 'confirm' ? true : undefined);
      });
      if (input) {
        input.addEventListener('keydown', function (ev) {
          if (ev.key === 'Enter') { ev.preventDefault(); dlg.close(input.value); }
        });
      }
      dlg.focus(input || ok);
    });
  }

  function alert(message, title) {
    return openMessage({ kind: 'alert', message: message, title: title });
  }

  function confirm(message, title) {
    return openMessage({ kind: 'confirm', message: message, title: title });
  }

  function prompt(message, value, title) {
    return openMessage({ kind: 'prompt', message: message, value: value, title: title });
  }

  /* ---------- loading mask ---------- */
  function mask(target) {
    var targetEl = target || document.querySelector('.content') || document.body;
    var m = el('div', 'ui-mask');
    m.appendChild(el('div', 'ui-spinner'));
    targetEl.appendChild(m);
    return m;
  }

  function unmask(m) {
    if (m) m.remove();
  }

  /* ---------- partial navigation (pjax) ----------
     Menu, tab and in-page links fetch the target page and swap only .content, so
     the sidebar, tab bar and document survive navigation. The fetched page's own
     scripts arrive inside the fragment; innerHTML inserts them inert, so they are
     re-created here to run. A page without the app shell (login redirect, setup,
     refusal page) falls back to a normal navigation. Form posts stay full
     navigations — they go through PRG redirects and rewriting the page is harmless. */
  var listeners = [];
  function onNavigate(fn) { listeners.push(fn); }

  var seq = 0;
  var loaded = location.pathname + location.search;

  /* ---------- select pickers ----------
     A server-rendered <select data-picker> becomes the Vue-style tag + picker dialog.
     The select stays in the form as the source of truth; it is only made invisible
     (not display:none), so required validation still fires and FormData still carries
     the chosen values. */
  function initPickers(root) {
    root.querySelectorAll('select[data-picker]').forEach(function (select) {
      if (select.getAttribute('data-picker-ready')) return;
      select.setAttribute('data-picker-ready', '1');
      var multiple = select.multiple;

      var wrap = el('div', 'ui-picker');
      select.parentNode.insertBefore(wrap, select);
      wrap.appendChild(select);
      select.classList.add('ui-picker-native');

      var trigger = el('div', 'ui-picker-trigger');
      var tags = el('div', 'ui-picker-tags');
      var btn = el('button', 'btn btn-sm ui-picker-btn');
      btn.type = 'button';
      trigger.appendChild(tags);
      trigger.appendChild(btn);
      wrap.appendChild(trigger);

      function selectedOptions() {
        return Array.prototype.filter.call(select.options, function (o) {
          return o.selected && o.value !== '';
        });
      }

      function fireChange() {
        select.dispatchEvent(new Event('change', { bubbles: true }));
      }

      function render() {
        tags.innerHTML = '';
        var selected = selectedOptions();
        if (!selected.length) {
          var placeholder = el('span', 'ui-picker-placeholder');
          placeholder.textContent = select.getAttribute('data-placeholder') || '';
          tags.appendChild(placeholder);
        } else {
          selected.forEach(function (option) {
            var tag = el('span', 'ui-picker-tag');
            var text = el('span');
            text.textContent = option.textContent.trim();
            var close = el('button', 'ui-picker-tag-close');
            close.type = 'button';
            close.innerHTML = '&times;';
            close.title = '×';
            close.addEventListener('click', function (ev) {
              ev.preventDefault();
              ev.stopPropagation();
              option.selected = false;
              fireChange();
              render();
            });
            tag.appendChild(text);
            tag.appendChild(close);
            tags.appendChild(tag);
          });
        }
        if (selected.length) btn.textContent = label('change', 'Change');
        else btn.textContent = multiple ? label('add', 'Add') : label('select', 'Select');
      }

      function openPicker() {
        var dlg = modal({
          title: select.getAttribute('data-title') || '',
          width: '560px',
          cancelValue: null
        });
        var list = el('div', 'ui-picker-list');
        var choices = [];
        Array.prototype.forEach.call(select.options, function (option) {
          if (option.value === '') return;
          var row = el('label', 'ui-picker-row');
          var input = el('input');
          input.type = multiple ? 'checkbox' : 'radio';
          input.name = 'ui-picker-choice';
          input.value = option.value;
          input.checked = option.selected;
          var text = el('span', 'ui-picker-option');
          text.textContent = option.textContent.trim();
          row.appendChild(input);
          row.appendChild(text);
          list.appendChild(row);
          choices.push({ option: option, input: input, row: row });
        });
        var empty = el('div', 'ui-picker-empty');
        empty.textContent = label('no-options', 'No options');
        empty.style.display = choices.length ? 'none' : '';
        list.appendChild(empty);

        var body = el('div');
        // Multi-select pickers can pin the pending selection above the list, so paging
        // and searching never hide what has already been checked.
        var chosenBox = null;
        if (select.hasAttribute('data-selected-header')) {
          chosenBox = el('div', 'ui-picker-chosen');
          body.appendChild(chosenBox);
        }
        var search = null;
        if (select.hasAttribute('data-search')) {
          var filterBar = el('div', 'ui-picker-filter');
          search = el('input', 'ui-dialog-input ui-picker-search');
          search.type = 'search';
          search.placeholder = select.getAttribute('data-search-placeholder') || label('search', 'Search');
          filterBar.appendChild(search);
          body.appendChild(filterBar);
        }
        body.appendChild(list);

        var pageSize = parseInt(select.getAttribute('data-page-size') || '0', 10) || 0;
        var page = 1;
        var pagerWrap = null;
        var info = null;
        var pager = null;
        if (pageSize > 0) {
          pagerWrap = el('div', 'ui-picker-pager');
          info = el('span', 'page-info');
          pager = el('div', 'pagination');
          pagerWrap.appendChild(info);
          pagerWrap.appendChild(pager);
          body.appendChild(pagerWrap);
        }

        function renderChosenHeader() {
          if (!chosenBox) return;
          chosenBox.innerHTML = '';
          var picked = choices.filter(function (c) { return c.input.checked; });
          chosenBox.style.display = picked.length ? '' : 'none';
          picked.forEach(function (c) {
            var tag = el('span', 'ui-picker-tag');
            var text = el('span');
            text.textContent = c.option.textContent.trim();
            var close = el('button', 'ui-picker-tag-close');
            close.type = 'button';
            close.innerHTML = '&times;';
            close.addEventListener('click', function (ev) {
              ev.preventDefault();
              c.input.checked = false;
              renderChosenHeader();
            });
            tag.appendChild(text);
            tag.appendChild(close);
            chosenBox.appendChild(tag);
          });
        }

        function visibleChoices() {
          var q = search ? search.value.trim().toLowerCase() : '';
          if (!q) return choices;
          return choices.filter(function (c) {
            return c.option.textContent.toLowerCase().indexOf(q) >= 0;
          });
        }

        function renderPager(total) {
          if (!pagerWrap) return;
          var pages = Math.max(1, Math.ceil(total / pageSize));
          if (page > pages) page = pages;
          var start = (page - 1) * pageSize;
          var from = total ? start + 1 : 0;
          var to = Math.min(start + pageSize, total);
          var tpl = select.getAttribute('data-page-info') || '{from}-{to} / {total}';
          info.textContent = String(tpl)
            .replace('{from}', from).replace('{to}', to).replace('{total}', total);
          pager.innerHTML = '';
          var last = 0;
          function add(label, target, current) {
            var node;
            if (current) {
              node = el('span', 'cur');
            } else {
              node = el('a');
              node.href = '#';
              node.addEventListener('click', function (ev) {
                ev.preventDefault();
                page = target;
                applyView();
              });
            }
            node.textContent = label;
            pager.appendChild(node);
          }
          if (page > 1) add('‹', page - 1, false);
          for (var no = 1; no <= pages; no++) {
            if (!(no === 1 || no === pages || Math.abs(no - page) <= 2)) continue;
            if (last && no > last + 1) {
              var gap = el('span', 'gap');
              gap.textContent = '…';
              pager.appendChild(gap);
            }
            add(String(no), no, no === page);
            last = no;
          }
          if (page < pages) add('›', page + 1, false);
        }

        function applyView() {
          var visible = visibleChoices();
          choices.forEach(function (c) { c.row.style.display = 'none'; });
          var start = 0;
          var end = visible.length;
          if (pageSize > 0) {
            var pages = Math.max(1, Math.ceil(visible.length / pageSize));
            if (page > pages) page = pages;
            start = (page - 1) * pageSize;
            end = start + pageSize;
          }
          visible.slice(start, end).forEach(function (c) { c.row.style.display = ''; });
          empty.style.display = visible.length ? 'none' : '';
          renderPager(visible.length);
          renderChosenHeader();
        }

        if (search) {
          search.addEventListener('input', function () {
            page = 1;
            applyView();
          });
        }
        choices.forEach(function (c) {
          c.input.addEventListener('change', renderChosenHeader);
        });

        dlg.setBody(body);
        dlg.addButton(label('cancel', 'Cancel'), false, function () { dlg.close(); });
        var ok = dlg.addButton(label('ok', 'OK'), true, function () {
          choices.forEach(function (c) { c.option.selected = c.input.checked; });
          if (!multiple && !choices.some(function (c) { return c.input.checked; })) {
            var placeholderIndex = Array.prototype.findIndex.call(select.options, function (o) {
              return o.value === '';
            });
            select.selectedIndex = placeholderIndex;
          }
          fireChange();
          render();
          dlg.close();
        });
        if (!choices.length) ok.disabled = true;
        applyView();
        dlg.focus(search || list.querySelector('input:checked') || list.querySelector('input'));
      }

      // The select itself is invisible, so surface its validation failure on the
      // trigger the user can actually see.
      select.addEventListener('invalid', function () {
        trigger.classList.add('is-invalid');
        trigger.scrollIntoView({ block: 'center' });
      });
      select.addEventListener('change', function () {
        trigger.classList.remove('is-invalid');
      });

      // The control sits inside the field's <label>, so clicking the field caption would
      // otherwise activate the (invisible) native select and open its dropdown.
      var fieldLabel = select.closest('label');
      if (fieldLabel) {
        fieldLabel.addEventListener('click', function (ev) {
          if (ev.target.closest('.ui-picker')) return;
          ev.preventDefault();
          openPicker();
        });
      }

      // preventDefault also keeps the surrounding <label> from forwarding the click
      // to the native select and opening its dropdown behind the picker.
      trigger.addEventListener('click', function (ev) {
        if (ev.target.closest('.ui-picker-tag-close')) return;
        ev.preventDefault();
        openPicker();
      });
      render();
    });
  }

  function runScripts(root) {
    root.querySelectorAll('script').forEach(function (old) {
      var s = document.createElement('script');
      for (var i = 0; i < old.attributes.length; i++) {
        s.setAttribute(old.attributes[i].name, old.attributes[i].value);
      }
      s.textContent = old.textContent;
      old.parentNode.replaceChild(s, old);
    });
    initPickers(root);
  }

  function autoHideFlash(flash) {
    setTimeout(function () {
      flash.style.transition = 'opacity .4s';
      flash.style.opacity = '0';
      setTimeout(function () { flash.remove(); }, 400);
    }, 4000);
  }

  // A page fetched for navigation may carry a flash (PRG query params); it belongs
  // where the layout would have put it, right above the content.
  function swapFlash(doc) {
    var main = document.querySelector('.main');
    if (!main) return;
    var old = main.querySelector(':scope > .flash');
    if (old) old.remove();
    var next = doc.querySelector('.main > .flash');
    if (!next) return;
    var flash = document.importNode(next, true);
    main.insertBefore(flash, main.querySelector('.content'));
    if (flash.classList.contains('flash-ok')) autoHideFlash(flash);
  }

  function apply(text, finalUrl, push) {
    var doc = parsePage(text);
    var current = document.querySelector('.content');
    var next = doc.querySelector('.content');
    if (!current || !next || !doc.querySelector('.app')) {
      location.href = finalUrl;
      return;
    }
    swapFlash(doc);
    current.innerHTML = next.innerHTML;
    runScripts(current);
    document.title = doc.title;
    if (push) history.pushState({ pjax: true }, '', finalUrl);
    var landed = new URL(finalUrl, location.href);
    loaded = landed.pathname + landed.search;
    window.scrollTo(0, 0);
    listeners.forEach(function (fn) { fn(doc); });
  }

  function navigate(url, push) {
    if (push === undefined) push = true;
    if (!document.querySelector('.app')) {
      location.href = url;
      return;
    }
    var mySeq = ++seq;
    var handle = mask(document.querySelector('.content'));
    fetch(url, { credentials: 'same-origin', cache: 'no-store' })
      .then(function (resp) {
        var type = resp.headers.get('content-type') || '';
        if (!resp.ok || type.indexOf('text/html') < 0) throw new Error('not a page');
        return resp.text().then(function (text) { return { text: text, url: resp.url || url }; });
      })
      .then(function (r) {
        if (mySeq !== seq) return;
        apply(r.text, r.url, push);
      })
      .catch(function () {
        if (mySeq === seq) location.href = url;
      })
      .then(function () { unmask(handle); });
  }

  // Reload the current page's content area without touching history.
  function refresh() {
    navigate(location.pathname + location.search + location.hash, false);
  }

  /* ---------- file attachments ----------
     A .file-upload block pairs a hidden file input with a button-like trigger and a
     chip list. fileChips wires the chip rendering and per-file removal; readFiles
     resolves the selection as base64 data URLs for the JSON send APIs. */
  function fileBox(scope) {
    return (scope || document).querySelector('.file-upload');
  }

  function fileChips(scope) {
    var box = fileBox(scope);
    if (!box) return;
    var input = box.querySelector('input[type=file]');
    var list = box.querySelector('.file-chips');
    if (!input || !list) return;
    var removeLabel = box.getAttribute('data-remove-label') || '×';
    // Removing one picked file means rebuilding the input's FileList, which needs
    // DataTransfer; browsers without it show the names but no ×.
    var canRemove = (function () {
      try { return !!new DataTransfer(); } catch (e) { return false; }
    })();
    function remove(index) {
      var picked = Array.prototype.slice.call(input.files || []);
      picked.splice(index, 1);
      var dt = new DataTransfer();
      picked.forEach(function (f) { dt.items.add(f); });
      input.files = dt.files;
      render();
    }
    function render() {
      list.innerHTML = '';
      Array.prototype.slice.call(input.files || []).forEach(function (f, idx) {
        var chip = el('span', 'file-chip');
        var name = el('span', 'file-chip-name');
        name.textContent = f.name;
        chip.appendChild(name);
        if (canRemove) {
          var x = el('button', 'file-chip-x');
          x.type = 'button';
          x.textContent = '×';
          x.setAttribute('aria-label', removeLabel);
          x.addEventListener('click', function () { remove(idx); });
          chip.appendChild(x);
        }
        list.appendChild(chip);
      });
    }
    input.addEventListener('change', render);
  }

  function readFiles(scope) {
    var box = fileBox(scope);
    var input = box && box.querySelector('input[type=file]');
    var picked = input && input.files ? Array.prototype.slice.call(input.files) : [];
    return Promise.all(picked.map(function (f) {
      return new Promise(function (resolve, reject) {
        var reader = new FileReader();
        reader.onload = function () { resolve({ name: f.name, data: String(reader.result) }); };
        reader.onerror = reject;
        reader.readAsDataURL(f);
      });
    }));
  }

  /* ---------- page dialogs ----------
     Load a server-rendered page into a dialog. It is the same markup the page would
     serve standalone: the page head (back link) is dropped and a Cancel button is
     added to the form actions, if any. A page script reports success with
     UI.dialogDone(button, message), which closes the dialog, toasts (unless message is
     empty) and refreshes the list underneath — the Vue-dialog behaviour. */
  function prepareDialogContent(content, dlg) {
    var head = content.querySelector('.page-head');
    if (head) head.remove();
    var actions = content.querySelector('.form-actions');
    if (actions && !actions.querySelector('.js-dialog-cancel')) {
      var cancel = el('button', 'btn js-dialog-cancel');
      cancel.type = 'button';
      cancel.textContent = label('cancel', 'Cancel');
      cancel.addEventListener('click', function () { dlg.close(); });
      actions.appendChild(cancel);
    }
    return content;
  }

  function dialogFlash(doc, content) {
    var flash = doc.querySelector('.main > .flash');
    if (!flash) return;
    var f = el('div', flash.className + ' ui-dialog-flash');
    f.textContent = flash.textContent.trim();
    content.insertBefore(f, content.firstChild);
    if (f.classList.contains('flash-ok')) autoHideFlash(f);
  }

  function pageDialog(opts) {
    var dlg = modal({ width: opts.width || '720px' });
    dlg.setLoading(true);
    fetch(opts.url, { credentials: 'same-origin', cache: 'no-store' })
      .then(function (resp) {
        if (!resp.ok) throw new Error('not a page');
        return resp.text().then(function (text) { return { resp: resp, text: text }; });
      })
      .then(function (r) {
        var doc = parsePage(r.text);
        var content = doc.querySelector('.content');
        if (!content || !doc.querySelector('.app')) { location.href = r.resp.url || opts.url; return; }
        dlg.setTitle(pageTitle(doc));
        var node = prepareDialogContent(document.importNode(content, true), dlg);
        dlg.setLoading(false);
        dlg.setBody(node);
        runScripts(node);
      })
      .catch(function () { dlg.close(); location.href = opts.url; });

    dlg.root.addEventListener('ui-dialog-done', function (ev) {
      var message = ev.detail && ev.detail.message;
      dlg.close();
      if (message) toast(message, 'ok');
      if (opts.onDone) opts.onDone();
      else refresh();
    });
    return dlg;
  }

  // Signal a page dialog that its work finished; harmless outside a dialog.
  function dialogDone(target, message) {
    var node = target && target.dispatchEvent ? target : document.body;
    node.dispatchEvent(new CustomEvent('ui-dialog-done', {
      bubbles: true,
      detail: { message: message || '' }
    }));
  }

  /* ---------- form dialogs ----------
     The add/edit forms stay ordinary server-rendered pages; loading one into a
     dialog gives the Vue-style interaction. Submitting posts the same form with
     fetch: a redirect means the save succeeded (alert, then close and refresh the
     list), a 400 re-render means validation failed (show it inside the dialog). */
  function formDialog(opts) {
    var dlg = modal({ width: opts.width || '640px' });
    var url = opts.url;
    var successText = opts.successText || '';
    dlg.setLoading(true);

    function show(text, finalUrl) {
      var doc = parsePage(text);
      var content = doc.querySelector('.content');
      if (!content || !doc.querySelector('.app')) { location.href = finalUrl; return; }
      dlg.setTitle(pageTitle(doc));
      var node = prepareDialogContent(document.importNode(content, true), dlg);
      dlg.setLoading(false);
      dlg.setBody(node);
      runScripts(node);
    }

    fetch(url, { credentials: 'same-origin', cache: 'no-store' })
      .then(function (resp) {
        if (!resp.ok) throw new Error('not a page');
        return resp.text().then(function (text) { return { resp: resp, text: text }; });
      })
      .then(function (r) { show(r.text, r.resp.url || url); })
      .catch(function () { dlg.close(); location.href = url; });

    dlg.body.addEventListener('submit', function (ev) {
      var form = ev.target.closest('form');
      if (!form) return;
      ev.preventDefault();
      dlg.setLoading(true);
      var method = (form.getAttribute('method') || 'post').toUpperCase();
      // Submit urlencoded like the browser's own form post: the handlers use
      // Request.Form, which ParseForm does not fill for multipart bodies.
      var encoded = new URLSearchParams();
      new FormData(form).forEach(function (value, key) { encoded.append(key, value); });
      fetch(new URL(form.getAttribute('action') || url, location.href).href, {
        method: method,
        headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
        body: encoded,
        credentials: 'same-origin'
      }).then(function (resp) {
        return resp.text().then(function (text) { return { resp: resp, text: text }; });
      }).then(function (r) {
        var doc = parsePage(r.text);
        if (!doc.querySelector('.app')) { location.href = r.resp.url || url; return; }
        if (r.resp.redirected) {
          var done = function () {
            dlg.close();
            if (opts.onSaved) opts.onSaved();
            else refresh();
          };
          if (successText) alert(successText).then(done);
          else done();
          return;
        }
        var content = doc.querySelector('.content');
        if (!content) { location.href = r.resp.url || url; return; }
        var node = prepareDialogContent(document.importNode(content, true), dlg);
        dialogFlash(doc, node);
        dlg.setLoading(false);
        dlg.setBody(node);
        runScripts(node);
      }).catch(function () {
        dlg.setLoading(false);
        alert('request failed');
      });
    });
    return dlg;
  }

  /* ---------- declarative dialog triggers ---------- */
  document.addEventListener('click', function (ev) {
    var formBtn = ev.target.closest('[data-form-dialog]');
    if (formBtn) {
      ev.preventDefault();
      formDialog({
        url: formBtn.getAttribute('data-form-dialog'),
        successText: formBtn.getAttribute('data-success') || '',
        width: formBtn.getAttribute('data-dialog-width') || undefined
      });
      return;
    }
    var pageBtn = ev.target.closest('[data-page-dialog]');
    if (pageBtn) {
      ev.preventDefault();
      pageDialog({
        url: pageBtn.getAttribute('data-page-dialog'),
        width: pageBtn.getAttribute('data-dialog-width') || undefined
      });
    }
  });

  /* ---------- page forms ----------
     GET forms (the list filters) and POST forms marked data-pjax-form submit through
     fetch instead of reloading the whole shell. Errors re-render inside .content. */
  function submitPageForm(form, successText) {
    var method = (form.getAttribute('method') || 'get').toUpperCase();
    if (method === 'GET') {
      var params = new URLSearchParams();
      new FormData(form).forEach(function (value, key) {
        if (value !== '') params.append(key, value);
      });
      var action = new URL(form.getAttribute('action') || location.pathname, location.href);
      var qs = params.toString();
      navigate(action.pathname + (qs ? '?' + qs : ''), true);
      return;
    }
    var handle = mask(document.querySelector('.content'));
    var encoded = new URLSearchParams();
    new FormData(form).forEach(function (value, key) { encoded.append(key, value); });
    fetch(new URL(form.getAttribute('action') || location.pathname, location.href).href, {
      method: 'POST',
      headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
      body: encoded,
      credentials: 'same-origin'
    }).then(function (resp) {
      return resp.text().then(function (text) { return { resp: resp, text: text }; });
    }).then(function (r) {
      unmask(handle);
      var doc = parsePage(r.text);
      var content = doc.querySelector('.content');
      if (!content || !doc.querySelector('.app')) { location.href = r.resp.url || location.href; return; }
      if (r.resp.redirected) {
        if (successText) alert(successText).then(refresh);
        else refresh();
        return;
      }
      var current = document.querySelector('.content');
      var node = document.importNode(content, true);
      dialogFlash(doc, node);
      current.innerHTML = node.innerHTML;
      runScripts(current);
      document.title = doc.title;
      listeners.forEach(function (fn) { fn(doc); });
      window.scrollTo(0, 0);
    }).catch(function () {
      unmask(handle);
      alert('request failed');
    });
  }

  document.addEventListener('submit', function (ev) {
    var form = ev.target;
    if (ev.defaultPrevented) return;
    if (!document.querySelector('.app')) return;
    if (form.closest && form.closest('.ui-dialog')) return;
    var method = (form.getAttribute('method') || 'get').toUpperCase();
    if (method === 'GET') {
      ev.preventDefault();
      submitPageForm(form, '');
      return;
    }
    if (form.hasAttribute('data-pjax-form')) {
      ev.preventDefault();
      submitPageForm(form, form.getAttribute('data-success') || '');
    }
  });

  document.addEventListener('click', function (ev) {
    if (ev.defaultPrevented || ev.button !== 0) return;
    if (ev.metaKey || ev.ctrlKey || ev.shiftKey || ev.altKey) return;
    if (!document.querySelector('.app')) return;
    var a = ev.target.closest('a[href]');
    if (!a) return;
    if (a.target && a.target !== '_self') return;
    if (a.hasAttribute('download')) return;
    var href = a.getAttribute('href');
    if (!href || href.charAt(0) === '#') return;
    var url;
    try { url = new URL(a.href); } catch (e) { return; }
    if (url.origin !== location.origin) return;
    if (url.pathname.indexOf('/api/') === 0) return;
    if (url.pathname + url.search === location.pathname + location.search) {
      // Already showing this page: clicking its tab does nothing; a different
      // hash still gets the browser's normal in-page jump.
      if (!url.hash || url.hash === location.hash) ev.preventDefault();
      return;
    }
    ev.preventDefault();
    navigate(url.pathname + url.search + url.hash, true);
  });

  window.addEventListener('popstate', function () {
    var here = location.pathname + location.search;
    if (here === loaded) return; // hash-only change
    navigate(here + location.hash, false);
  });

  document.addEventListener('DOMContentLoaded', function () {
    document.querySelectorAll('.flash-ok').forEach(autoHideFlash);
    initPickers(document);
  });

  return {
    modal: modal,
    alert: alert,
    confirm: confirm,
    prompt: prompt,
    toast: toast,
    mask: mask,
    unmask: unmask,
    navigate: navigate,
    refresh: refresh,
    onNavigate: onNavigate,
    pageDialog: pageDialog,
    formDialog: formDialog,
    dialogDone: dialogDone,
    fileChips: fileChips,
    readFiles: readFiles
  };
})();
