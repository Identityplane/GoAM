(function(factory) {
  typeof define === "function" && define.amd ? define(factory) : factory();
})((function() {
  "use strict";
  var __vite_style__ = document.createElement("style");
  __vite_style__.textContent = ".otp-group {\n  margin: 20px 0;\n}\n\n.otp-inputs {\n  display: none; /* Hidden by default, shown by JS */\n}\n\n.otp-inputs.enhanced {\n  display: flex;\n  gap: 10px;\n  justify-content: center;\n}\n\n.otp-inputs.enhanced .otp-input {\n  width: 40px;\n  max-width: 40px;\n  padding: 0;\n}\n\n/* Hide the basic input when enhanced version is active */\n.otp-inputs.enhanced ~ .otp-input {\n  display: none;\n}\n\n/* Basic input styling */\n.otp-input {\n  width: 100%;\n  max-width: 300px;\n  height: 40px;\n  text-align: center;\n  font-size: 20px;\n  border: 2px solid #ccc;\n  border-radius: 4px;\n  padding: 0 10px;\n}\n\n.otp-input:focus {\n  border-color: #007bff;\n  outline: none;\n} /*$vite$:1*/";
  document.head.appendChild(__vite_style__);
  function initEmailOTP() {
    var _a;
    const form = document.getElementById("otpForm");
    const basicInput = document.getElementById("otp");
    if (!form || !basicInput) {
      console.error("Required form elements not found");
      return;
    }
    initResendCountdown();
    const otpInputs = document.createElement("div");
    otpInputs.className = "otp-inputs";
    for (let i2 = 0; i2 < 6; i2++) {
      const input = document.createElement("input");
      input.type = "text";
      input.maxLength = 1;
      input.pattern = "[0-9]";
      input.inputMode = "numeric";
      input.className = "otp-input";
      input.required = true;
      otpInputs.appendChild(input);
    }
    (_a = basicInput.parentElement) == null ? void 0 : _a.insertBefore(otpInputs, basicInput);
    otpInputs.classList.add("enhanced");
    const inputs = otpInputs.querySelectorAll(".otp-input");
    inputs.forEach((input, index) => {
      input.addEventListener("input", function() {
        if (this.value.length === 1) {
          if (index < inputs.length - 1) {
            inputs[index + 1].focus();
          }
        }
      });
      input.addEventListener("keydown", function(e2) {
        if (e2.key === "Backspace" && !this.value && index > 0) {
          inputs[index - 1].focus();
        }
      });
      input.addEventListener("paste", function(e2) {
        var _a2;
        e2.preventDefault();
        const pastedData = ((_a2 = e2.clipboardData) == null ? void 0 : _a2.getData("text").slice(0, 6)) || "";
        if (/^\d+$/.test(pastedData)) {
          pastedData.split("").forEach((digit, i2) => {
            if (inputs[i2]) {
              inputs[i2].value = digit;
            }
          });
          updateBasicInput();
          if (pastedData.length === 6) {
            form.submit();
          }
        }
      });
    });
    function updateBasicInput() {
      if (!basicInput) return;
      const otp = Array.from(inputs).map((input) => input.value).join("");
      basicInput.value = otp;
    }
    form.addEventListener("submit", function(e2) {
      e2.preventDefault();
      updateBasicInput();
      if (basicInput.value.length === 6) {
        form.submit();
      }
    });
    inputs.forEach((input) => {
      input.addEventListener("input", function() {
        updateBasicInput();
        if (basicInput.value.length === 6) {
          form.submit();
        }
      });
    });
  }
  function initResendCountdown() {
    const resendButton = document.getElementById("resend-otp-button");
    if (!resendButton) {
      return;
    }
    const resendInSeconds = parseInt(resendButton.getAttribute("data-resend-in-seconds") || "0");
    if (resendInSeconds <= 0) {
      return;
    }
    let remainingSeconds = resendInSeconds;
    function updateButtonText() {
      if (remainingSeconds > 0) {
        resendButton.textContent = `Resend in (${remainingSeconds})`;
        resendButton.disabled = true;
        remainingSeconds--;
      } else {
        resendButton.textContent = "Resend";
        resendButton.disabled = false;
        return;
      }
    }
    updateButtonText();
    const countdownInterval = setInterval(() => {
      updateButtonText();
      if (remainingSeconds < 0) {
        clearInterval(countdownInterval);
      }
    }, 1e3);
  }
  function base64urlToBuffer(base64url) {
    const padding = "=".repeat((4 - base64url.length % 4) % 4);
    const base64 = (base64url + padding).replace(/-/g, "+").replace(/_/g, "/");
    const raw = atob(base64);
    return new Uint8Array([...raw].map((char) => char.charCodeAt(0)));
  }
  function bufferToBase64url(buffer) {
    const binary = String.fromCharCode.apply(null, new Uint8Array(buffer));
    const base64 = btoa(binary);
    return base64.replace(/\+/g, "-").replace(/\//g, "_").replace(/=/g, "");
  }
  function serializeCredential(cred) {
    var _a;
    if (!cred) return null;
    const json = {
      id: cred.id,
      type: cred.type,
      rawId: bufferToBase64url(cred.rawId),
      authenticatorAttachment: cred.authenticatorAttachment || null,
      clientExtensionResults: ((_a = cred.getClientExtensionResults) == null ? void 0 : _a.call(cred)) || {},
      response: {}
    };
    if (cred.response) {
      const r2 = cred.response;
      if (r2.clientDataJSON) json.response.clientDataJSON = bufferToBase64url(r2.clientDataJSON);
      if (r2.attestationObject) json.response.attestationObject = bufferToBase64url(r2.attestationObject);
      if (r2.authenticatorData) json.response.authenticatorData = bufferToBase64url(r2.authenticatorData);
      if (r2.signature) json.response.signature = bufferToBase64url(r2.signature);
      if (r2.userHandle) json.response.userHandle = bufferToBase64url(r2.userHandle);
      if (r2.publicKey) json.response.publicKey = r2.publicKey;
      if (r2.publicKeyAlgorithm) json.response.publicKeyAlgorithm = r2.publicKeyAlgorithm;
      if (r2.transports) json.response.transports = r2.transports;
    }
    return json;
  }
  async function initPasswordOrSocialLogin() {
    console.log("initializing PasswordOrSocialLogin");
    const forgotPasswordLink = document.getElementById("forgot-password-link");
    if (forgotPasswordLink) {
      forgotPasswordLink.addEventListener("click", (event) => {
        var _a;
        event.preventDefault();
        (_a = document.getElementById("forgotPasswordForm")) == null ? void 0 : _a.submit();
      });
    }
    if (window.PublicKeyCredential && PublicKeyCredential.isConditionalMediationAvailable) {
      const isCMA = await PublicKeyCredential.isConditionalMediationAvailable();
      if (isCMA) {
        const optionsInput = document.getElementById("passkeysLoginOptions").value;
        const options = JSON.parse(optionsInput);
        options.publicKey.challenge = base64urlToBuffer(options.publicKey.challenge);
        const assertion = await navigator.credentials.get({
          publicKey: options.publicKey,
          mediation: "conditional"
        });
        const serializedAssertion = serializeCredential(assertion);
        console.log(serializedAssertion);
        document.getElementById("passkeysFinishLoginJson").value = JSON.stringify(serializedAssertion);
        document.getElementById("passkey-form").submit();
      }
    }
  }
  var K = Object.defineProperty, Y = Object.defineProperties;
  var V = Object.getOwnPropertyDescriptors;
  var y = Object.getOwnPropertySymbols;
  var P = Object.prototype.hasOwnProperty, D = Object.prototype.propertyIsEnumerable;
  var I = (t2, e2, r2) => e2 in t2 ? K(t2, e2, { enumerable: true, configurable: true, writable: true, value: r2 }) : t2[e2] = r2, d = (t2, e2) => {
    for (var r2 in e2 || (e2 = {})) P.call(e2, r2) && I(t2, r2, e2[r2]);
    if (y) for (var r2 of y(e2)) D.call(e2, r2) && I(t2, r2, e2[r2]);
    return t2;
  }, O = (t2, e2) => Y(t2, V(e2));
  var M = (t2, e2) => {
    var r2 = {};
    for (var n2 in t2) P.call(t2, n2) && e2.indexOf(n2) < 0 && (r2[n2] = t2[n2]);
    if (t2 != null && y) for (var n2 of y(t2)) e2.indexOf(n2) < 0 && D.call(t2, n2) && (r2[n2] = t2[n2]);
    return r2;
  };
  var l = (t2, e2, r2) => (I(t2, typeof e2 != "symbol" ? e2 + "" : e2, r2), r2);
  var x = (t2, e2, r2) => new Promise((n2, s) => {
    var a2 = (c) => {
      try {
        o2(r2.next(c));
      } catch (i2) {
        s(i2);
      }
    }, u2 = (c) => {
      try {
        o2(r2.throw(c));
      } catch (i2) {
        s(i2);
      }
    }, o2 = (c) => c.done ? n2(c.value) : Promise.resolve(c.value).then(a2, u2);
    o2((r2 = r2.apply(t2, e2)).next());
  });
  var L = "hCaptcha-script", b = "hCaptchaOnLoad", _ = "script-error";
  var g = "@hCaptcha/loader";
  function U(t2) {
    return Object.entries(t2).filter(([, e2]) => e2 || e2 === false).map(([e2, r2]) => `${encodeURIComponent(e2)}=${encodeURIComponent(String(r2))}`).join("&");
  }
  function R(t2) {
    let e2 = t2 && t2.ownerDocument || document, r2 = e2.defaultView || e2.parentWindow || window;
    return { document: e2, window: r2 };
  }
  function S(t2) {
    return t2 || document.head;
  }
  function F(t2) {
    var e2;
    t2.setTag("source", g), t2.setTag("url", document.URL), t2.setContext("os", { UA: navigator.userAgent }), t2.setContext("browser", d({}, z())), t2.setContext("device", O(d({}, J()), { screen_width_pixels: screen.width, screen_height_pixels: screen.height, language: navigator.language, orientation: ((e2 = screen.orientation) == null ? void 0 : e2.type) || "Unknown", processor_count: navigator.hardwareConcurrency, platform: navigator.platform }));
  }
  function z() {
    var n2, s, a2, u2, o2, c;
    let t2 = navigator.userAgent, e2, r2;
    return t2.indexOf("Firefox") !== -1 ? (e2 = "Firefox", r2 = (n2 = t2.match(/Firefox\/([\d.]+)/)) == null ? void 0 : n2[1]) : t2.indexOf("Edg") !== -1 ? (e2 = "Microsoft Edge", r2 = (s = t2.match(/Edg\/([\d.]+)/)) == null ? void 0 : s[1]) : t2.indexOf("Chrome") !== -1 && t2.indexOf("Safari") !== -1 ? (e2 = "Chrome", r2 = (a2 = t2.match(/Chrome\/([\d.]+)/)) == null ? void 0 : a2[1]) : t2.indexOf("Safari") !== -1 && t2.indexOf("Chrome") === -1 ? (e2 = "Safari", r2 = (u2 = t2.match(/Version\/([\d.]+)/)) == null ? void 0 : u2[1]) : t2.indexOf("Opera") !== -1 || t2.indexOf("OPR") !== -1 ? (e2 = "Opera", r2 = (o2 = t2.match(/(Opera|OPR)\/([\d.]+)/)) == null ? void 0 : o2[2]) : t2.indexOf("MSIE") !== -1 || t2.indexOf("Trident") !== -1 ? (e2 = "Internet Explorer", r2 = (c = t2.match(/(MSIE |rv:)([\d.]+)/)) == null ? void 0 : c[2]) : (e2 = "Unknown", r2 = "Unknown"), { name: e2, version: r2 };
  }
  function J() {
    let t2 = navigator.userAgent, e2;
    t2.indexOf("Win") !== -1 ? e2 = "Windows" : t2.indexOf("Mac") !== -1 ? e2 = "Mac" : t2.indexOf("Linux") !== -1 ? e2 = "Linux" : t2.indexOf("Android") !== -1 ? e2 = "Android" : t2.indexOf("like Mac") !== -1 || t2.indexOf("iPhone") !== -1 || t2.indexOf("iPad") !== -1 ? e2 = "iOS" : e2 = "Unknown";
    let r2;
    return /Mobile|iPhone|iPod|Android/i.test(t2) ? r2 = "Mobile" : /Tablet|iPad/i.test(t2) ? r2 = "Tablet" : r2 = "Desktop", { model: e2, family: e2, device: r2 };
  }
  var Q = class B {
    constructor(e2) {
      l(this, "_parent");
      l(this, "breadcrumbs", []);
      l(this, "context", {});
      l(this, "extra", {});
      l(this, "tags", {});
      l(this, "request");
      l(this, "user");
      this._parent = e2;
    }
    get parent() {
      return this._parent;
    }
    child() {
      return new B(this);
    }
    setRequest(e2) {
      return this.request = e2, this;
    }
    removeRequest() {
      return this.request = void 0, this;
    }
    addBreadcrumb(e2) {
      return typeof e2.timestamp > "u" && (e2.timestamp = (/* @__PURE__ */ new Date()).toISOString()), this.breadcrumbs.push(e2), this;
    }
    setExtra(e2, r2) {
      return this.extra[e2] = r2, this;
    }
    removeExtra(e2) {
      return delete this.extra[e2], this;
    }
    setContext(e2, r2) {
      return typeof r2.type > "u" && (r2.type = e2), this.context[e2] = r2, this;
    }
    removeContext(e2) {
      return delete this.context[e2], this;
    }
    setTags(e2) {
      return this.tags = d(d({}, this.tags), e2), this;
    }
    setTag(e2, r2) {
      return this.tags[e2] = r2, this;
    }
    removeTag(e2) {
      return delete this.tags[e2], this;
    }
    setUser(e2) {
      return this.user = e2, this;
    }
    removeUser() {
      return this.user = void 0, this;
    }
    toBody() {
      let e2 = [], r2 = this;
      for (; r2; ) e2.push(r2), r2 = r2.parent;
      return e2.reverse().reduce((n2, s) => {
        var a2;
        return n2.breadcrumbs = [...(a2 = n2.breadcrumbs) != null ? a2 : [], ...s.breadcrumbs], n2.extra = d(d({}, n2.extra), s.extra), n2.contexts = d(d({}, n2.contexts), s.context), n2.tags = d(d({}, n2.tags), s.tags), s.user && (n2.user = s.user), s.request && (n2.request = s.request), n2;
      }, { breadcrumbs: [], extra: {}, contexts: {}, tags: {}, request: void 0, user: void 0 });
    }
    clear() {
      this.breadcrumbs = [], this.context = {}, this.tags = {}, this.user = void 0;
    }
  }, Z = /^\s*at (?:(.*?) ?\()?((?:file|https?|blob|chrome-extension|address|native|eval|webpack|<anonymous>|[-a-z]+:|.*bundle|\/).*?)(?::(\d+))?(?::(\d+))?\)?\s*$/i, ee = /^\s*(.*?)(?:\((.*?)\))?(?:^|@)?((?:file|https?|blob|chrome|webpack|resource|moz-extension).*?:\/.*?|\[native code\]|[^@]*(?:bundle|\d+\.js))(?::(\d+))?(?::(\d+))?\s*$/i, te = /^\s*at (?:((?:\[object object\])?.+) )?\(?((?:file|ms-appx|https?|webpack|blob):.*?):(\d+)(?::(\d+))?\)?\s*$/i, re = /^(?:(\w+):)\/\/(?:(\w+)(?::(\w+))?@)([\w.-]+)(?::(\d+))?\/(.+)/, N = "?", k = "An unknown error occurred", ne = "0.0.4";
  function se(t2) {
    for (let e2 = 0; e2 < t2.length; e2++) t2[e2] = Math.floor(Math.random() * 256);
    return t2;
  }
  function p(t2) {
    return (t2 + 256).toString(16).substring(1);
  }
  function oe() {
    let t2 = se(new Array(16));
    return t2[6] = t2[6] & 15 | 64, t2[8] = t2[8] & 63 | 128, p(t2[0]) + p(t2[1]) + p(t2[2]) + p(t2[3]) + "-" + p(t2[4]) + p(t2[5]) + "-" + p(t2[6]) + p(t2[7]) + "-" + p(t2[8]) + p(t2[9]) + "-" + p(t2[10]) + p(t2[11]) + p(t2[12]) + p(t2[13]) + p(t2[14]) + p(t2[15]);
  }
  var ie = [[Z, "chrome"], [te, "winjs"], [ee, "gecko"]];
  function ae(t2) {
    var n2, s, a2, u2;
    if (!t2.stack) return null;
    let e2 = [], r2 = (a2 = (s = (n2 = t2.stack).split) == null ? void 0 : s.call(n2, `
`)) != null ? a2 : [];
    for (let o2 = 0; o2 < r2.length; ++o2) {
      let c = null, i2 = null, f = null;
      for (let [v, w] of ie) if (i2 = v.exec(r2[o2]), i2) {
        f = w;
        break;
      }
      if (!(!i2 || !f)) {
        if (f === "chrome") c = { filename: (u2 = i2[2]) != null && u2.startsWith("address at ") ? i2[2].substring(11) : i2[2], function: i2[1] || N, lineno: i2[3] ? +i2[3] : null, colno: i2[4] ? +i2[4] : null };
        else if (f === "winjs") c = { filename: i2[2], function: i2[1] || N, lineno: +i2[3], colno: i2[4] ? +i2[4] : null };
        else if (f === "gecko") o2 === 0 && !i2[5] && t2.columnNumber !== void 0 && e2.length > 0 && (e2[0].column = t2.columnNumber + 1), c = { filename: i2[3], function: i2[1] || N, lineno: i2[4] ? +i2[4] : null, colno: i2[5] ? +i2[5] : null };
        else continue;
        !c.function && c.lineno && (c.function = N), e2.push(c);
      }
    }
    return e2.length ? e2.reverse() : null;
  }
  function ce(t2) {
    let e2 = ae(t2);
    return { type: t2.name, value: t2.message, stacktrace: { frames: e2 != null ? e2 : [] } };
  }
  function ue(t2) {
    let e2 = re.exec(t2), r2 = e2 ? e2.slice(1) : [];
    if (r2.length !== 6) throw new Error("Invalid DSN");
    let n2 = r2[5].split("/"), s = n2.slice(0, -1).join("/");
    return r2[0] + "://" + r2[3] + (r2[4] ? ":" + r2[4] : "") + (s ? "/" + s : "") + "/api/" + n2.pop() + "/envelope/?sentry_version=7&sentry_key=" + r2[1] + (r2[2] ? "&sentry_secret=" + r2[2] : "");
  }
  function le(t2, e2, r2) {
    var s, a2;
    let n2 = d({ event_id: oe().replaceAll("-", ""), platform: "javascript", sdk: { name: "@hcaptcha/sentry", version: ne }, environment: e2, release: r2, timestamp: Date.now() / 1e3 }, t2.scope.toBody());
    if (t2.type === "exception") {
      n2.message = (a2 = (s = t2.error) == null ? void 0 : s.message) != null ? a2 : "Unknown error", n2.fingerprint = [n2.message];
      let u2 = [], o2 = t2.error;
      for (let c = 0; c < 5 && o2 && (u2.push(ce(o2)), !(!o2.cause || !(o2.cause instanceof Error))); c++) o2 = o2.cause;
      n2.exception = { values: u2.reverse() };
    }
    return t2.type === "message" && (n2.message = t2.message, n2.level = t2.level), n2;
  }
  function de(t2) {
    if (t2 instanceof Error) return t2;
    if (typeof t2 == "string") return new Error(t2);
    if (typeof t2 == "object" && t2 !== null && !Array.isArray(t2)) {
      let r2 = t2, { message: n2 } = r2, s = M(r2, ["message"]), a2 = new Error(typeof n2 == "string" ? n2 : k);
      return Object.assign(a2, s);
    }
    let e2 = new Error(k);
    return Object.assign(e2, { cause: t2 });
  }
  function pe(t2, e2, r2) {
    return x(this, null, function* () {
      var n2, s;
      try {
        if (typeof fetch < "u" && typeof AbortSignal < "u") {
          let a2;
          if (r2) {
            let c = new AbortController();
            a2 = c.signal, setTimeout(() => c.abort(), r2);
          }
          let u2 = yield fetch(t2, O(d({}, e2), { signal: a2 })), o2 = yield u2.text();
          return { status: u2.status, body: o2 };
        }
        return yield new Promise((a2, u2) => {
          var c, i2;
          let o2 = new XMLHttpRequest();
          if (o2.open((c = e2 == null ? void 0 : e2.method) != null ? c : "GET", t2), o2.onload = () => a2({ status: o2.status, body: o2.responseText }), o2.onerror = () => u2(new Error("XHR Network Error")), e2 == null ? void 0 : e2.headers) for (let [f, v] of Object.entries(e2.headers)) o2.setRequestHeader(f, v);
          if (r2) {
            let f = setTimeout(() => {
              o2.abort(), u2(new Error("Request timed out"));
            }, r2);
            o2.onloadend = () => {
              clearTimeout(f);
            };
          }
          o2.send((i2 = e2 == null ? void 0 : e2.body) == null ? void 0 : i2.toString());
        });
      } catch (a2) {
        return { status: 0, body: (s = (n2 = a2 == null ? void 0 : a2.toString) == null ? void 0 : n2.call(a2)) != null ? s : "Unknown error" };
      }
    });
  }
  var h, A = (h = class {
    constructor(e2) {
      l(this, "apiURL");
      l(this, "dsn");
      l(this, "environment");
      l(this, "release");
      l(this, "sampleRate");
      l(this, "debug");
      l(this, "_scope");
      l(this, "shouldBuffer", false);
      l(this, "bufferlimit", 20);
      l(this, "buffer", []);
      var r2, n2, s, a2, u2;
      this.environment = e2.environment, this.release = e2.release, this.sampleRate = (r2 = e2.sampleRate) != null ? r2 : 1, this.debug = (n2 = e2.debug) != null ? n2 : false, this._scope = (s = e2.scope) != null ? s : new Q(), this.apiURL = ue(e2.dsn), this.dsn = e2.dsn, this.shouldBuffer = (a2 = e2.buffer) != null ? a2 : false, this.bufferlimit = (u2 = e2.bufferLimit) != null ? u2 : 20;
    }
    static init(e2) {
      h._instance || (h._instance = new h(e2));
    }
    static get instance() {
      if (!h._instance) throw new Error("Sentry has not been initialized");
      return h._instance;
    }
    log(...e2) {
      this.debug && console.log(...e2);
    }
    get scope() {
      return this._scope;
    }
    static get scope() {
      return h.instance.scope;
    }
    withScope(e2) {
      let r2 = this._scope.child();
      e2(r2);
    }
    static withScope(e2) {
      h.instance.withScope(e2);
    }
    captureException(e2, r2) {
      this.captureEvent({ type: "exception", level: "error", error: de(e2), scope: r2 != null ? r2 : this._scope });
    }
    static captureException(e2, r2) {
      h.instance.captureException(e2, r2);
    }
    captureMessage(e2, r2 = "info", n2) {
      this.captureEvent({ type: "message", level: r2, message: e2, scope: n2 != null ? n2 : this._scope });
    }
    static captureMessage(e2, r2 = "info", n2) {
      h.instance.captureMessage(e2, r2, n2);
    }
    captureEvent(e2) {
      if (Math.random() >= this.sampleRate) {
        this.log("Dropped event due to sample rate");
        return;
      }
      if (this.shouldBuffer) {
        if (this.buffer.length >= this.bufferlimit) return;
        this.buffer.push(e2);
      } else this.sendEvent(e2);
    }
    sendEvent(e2, r2 = 5e3) {
      return x(this, null, function* () {
        try {
          this.log("Sending sentry event", e2);
          let n2 = le(e2, this.environment, this.release), s = { event_id: n2.event_id, dsn: this.dsn }, a2 = { type: "event" }, u2 = JSON.stringify(s) + `
` + JSON.stringify(a2) + `
` + JSON.stringify(n2), o2 = yield pe(this.apiURL, { method: "POST", headers: { "Content-Type": "application/x-sentry-envelope" }, body: u2 }, r2);
          this.log("Sentry response", o2.status), o2.status !== 200 && (console.log(o2.body), console.error("Failed to send event to Sentry", o2));
        } catch (n2) {
          console.error("Failed to send event", n2);
        }
      });
    }
    flush(e2 = 5e3) {
      return x(this, null, function* () {
        try {
          this.log("Flushing sentry events", this.buffer.length);
          let r2 = this.buffer.splice(0, this.buffer.length).map((n2) => this.sendEvent(n2, e2));
          yield Promise.all(r2), this.log("Flushed all events");
        } catch (r2) {
          console.error("Failed to flush events", r2);
        }
      });
    }
    static flush(e2 = 5e3) {
      return h.instance.flush(e2);
    }
    static reset() {
      h._instance = void 0;
    }
  }, l(h, "_instance"), h);
  var he = "https://d233059272824702afc8c43834c4912d@sentry.hcaptcha.com/6", fe = "2.1.0", me = "production";
  function j(t2 = true) {
    if (!t2) return q();
    A.init({ dsn: he, release: fe, environment: me });
    let e2 = A.scope;
    return F(e2), q(e2);
  }
  function q(t2 = null) {
    return { addBreadcrumb: (e2) => {
      t2 && t2.addBreadcrumb(e2);
    }, captureRequest: (e2) => {
      t2 && t2.setRequest(e2);
    }, captureException: (e2) => {
      t2 && A.captureException(e2, t2);
    } };
  }
  function H({ scriptLocation: t2, query: e2, loadAsync: r2 = true, crossOrigin: n2 = "anonymous", apihost: s = "https://js.hcaptcha.com", cleanup: a2 = false, secureApi: u2 = false, scriptSource: o2 = "" } = {}, c) {
    let i2 = S(t2), f = R(i2);
    return new Promise((v, w) => {
      let m = f.document.createElement("script");
      m.id = L, o2 ? m.src = `${o2}?onload=${b}` : u2 ? m.src = `${s}/1/secure-api.js?onload=${b}` : m.src = `${s}/1/api.js?onload=${b}`, m.crossOrigin = n2, m.async = r2;
      let T = (E, X) => {
        try {
          !u2 && a2 && i2.removeChild(m), X(E);
        } catch (G) {
          w(G);
        }
      };
      m.onload = (E) => T(E, v), m.onerror = (E) => {
        c && c(m.src), T(E, w);
      }, m.src += e2 !== "" ? `&${e2}` : "", i2.appendChild(m);
    });
  }
  var C = [];
  function ge(t2 = { cleanup: false }, e2) {
    try {
      e2.addBreadcrumb({ category: g, message: "hCaptcha loader params", data: t2 });
      let r2 = S(t2.scriptLocation), n2 = R(r2), s = C.find(({ scope: u2 }) => u2 === n2.window);
      if (s) return e2.addBreadcrumb({ category: g, message: "hCaptcha already loaded" }), s.promise;
      let a2 = new Promise((u2, o2) => x(this, null, function* () {
        try {
          n2.window[b] = () => {
            e2.addBreadcrumb({ category: g, message: "hCaptcha script called onload function" }), u2(n2.window.hcaptcha);
          };
          let c = U({ custom: t2.custom, render: t2.render, sentry: t2.sentry, assethost: t2.assethost, imghost: t2.imghost, reportapi: t2.reportapi, endpoint: t2.endpoint, host: t2.host, recaptchacompat: t2.recaptchacompat, hl: t2.hl });
          yield H(d({ query: c }, t2), (i2) => {
            e2.captureRequest({ url: i2, method: "GET" });
          }), e2.addBreadcrumb({ category: g, message: "hCaptcha loaded", data: s });
        } catch (c) {
          e2.addBreadcrumb({ category: g, message: "hCaptcha failed to load" });
          let i2 = C.findIndex((f) => f.scope === n2.window);
          i2 !== -1 && C.splice(i2, 1), o2(new Error(_));
        }
      }));
      return C.push({ promise: a2, scope: n2.window }), a2;
    } catch (r2) {
      return e2.captureException(r2), Promise.reject(new Error(_));
    }
  }
  function $(t2, e2, r2 = 0) {
    return x(this, null, function* () {
      let n2 = r2 < 2 ? "Retry loading hCaptcha Api" : "Exceeded maximum retries";
      try {
        return yield ge(t2, e2);
      } catch (s) {
        return e2.addBreadcrumb({ category: g, message: n2 }), r2 >= 2 ? (e2.captureException(s), Promise.reject(s)) : (r2 += 1, $(t2, e2, r2));
      }
    });
  }
  function xe() {
    return x(this, arguments, function* (t2 = {}) {
      let e2 = j(t2.sentry);
      return yield $(t2, e2);
    });
  }
  async function initHcaptcha() {
    console.log("initializing Hcaptcha");
    await xe();
    hcaptcha.render();
    const { response } = await hcaptcha.execute({ async: true });
    console.log("hcaptcha response", response);
    const tokenInput = document.getElementById("hcaptcha-token");
    tokenInput.value = response;
    const form = document.getElementById("hcaptcha-form");
    form.submit();
  }
  async function initVerifyPasskey() {
    console.log("initializing VerifyPasskey");
    const passkeyButton = document.getElementById("passkey-button");
    if (passkeyButton) {
      passkeyButton.addEventListener("click", async () => {
        await startPasskeyLogin();
      });
    }
  }
  async function startPasskeyLogin() {
    const optionsInput = document.getElementById("passkeysLoginOptions").value;
    const options = JSON.parse(optionsInput);
    options.publicKey.challenge = base64urlToBuffer(options.publicKey.challenge);
    const assertion = await navigator.credentials.get({
      publicKey: options.publicKey
    });
    const serializedAssertion = serializeCredential(assertion);
    console.log(serializedAssertion);
    document.getElementById("passkeysFinishLoginJson").value = JSON.stringify(serializedAssertion);
    document.getElementById("passkey-form").submit();
  }
  async function initRegisterPasskey() {
    console.log("initializing RegisterPasskey");
    const passkeyButton = document.getElementById("passkey-button");
    if (passkeyButton) {
      passkeyButton.addEventListener("click", async () => {
        await startPasskeyRegistration();
      });
    }
  }
  async function startPasskeyRegistration() {
    try {
      const optionsInput = document.getElementById("passkeysOptions").value;
      const options = JSON.parse(optionsInput);
      options.publicKey.challenge = base64urlToBuffer(options.publicKey.challenge);
      options.publicKey.user.id = base64urlToBuffer(options.publicKey.user.id);
      const useMock = false;
      let cred;
      if (useMock) ;
      else {
        cred = await navigator.credentials.create({ publicKey: options.publicKey });
      }
      serializedCred = serializeCredential(cred);
      document.getElementById("passkeysFinishRegistrationJson").value = JSON.stringify(serializedCred);
      document.getElementById("passkey-form").submit();
    } catch (err) {
      alert("Passkey registration failed: " + err.message);
      console.error(err);
    }
  }
  async function initOnboardingWithPasskey() {
    console.log("initializing OnboardingWithPasskey");
    const passkeyButton = document.getElementById("passkey-button");
    if (passkeyButton) {
      passkeyButton.addEventListener("click", async () => {
        await startPasskeyOnboarding();
      });
    }
  }
  async function startPasskeyOnboarding() {
    var _a;
    try {
      const emailInput = document.getElementById("email");
      const email = emailInput == null ? void 0 : emailInput.value;
      if (!email) {
        emailInput.reportValidity();
        emailInput.focus();
        return;
      }
      const actionBtn = document.getElementById("action");
      actionBtn.value = "passkey";
      const optionsInput = (_a = document.getElementById("passkeysOptions")) == null ? void 0 : _a.value;
      const options = JSON.parse(optionsInput);
      options.publicKey.user.name = email;
      options.publicKey.user.displayName = email;
      options.publicKey.challenge = base64urlToBuffer(options.publicKey.challenge);
      options.publicKey.user.id = base64urlToBuffer(options.publicKey.user.id);
      const useMock = false;
      let cred;
      if (useMock) ;
      else {
        cred = await navigator.credentials.create({ publicKey: options.publicKey });
      }
      serializedCred = serializeCredential(cred);
      document.getElementById("passkeysFinishRegistrationJson").value = JSON.stringify(serializedCred);
      document.getElementById("onboarding-with-passkey-form").submit();
    } catch (err) {
      alert("Passkey registration failed: " + err.message);
      console.error(err);
    }
  }
  async function initTelegramLogin() {
    console.log("initializing TelegramLogin");
    const tgAuthResult = window.location.hash.split("tgAuthResult=")[1];
    if (tgAuthResult) {
      console.log("tgAuthResult", tgAuthResult);
      const tgAuthResultInput = document.getElementById("tgAuthResult");
      if (tgAuthResultInput) {
        tgAuthResultInput.value = tgAuthResult;
      } else {
        console.error("Telegram auth result input not found");
        return;
      }
      const form = document.getElementById("telegramLoginForm");
      if (form) {
        form.submit();
      } else {
        console.error("Telegram login form not found");
      }
    }
  }
  function initNodeHistory() {
    const nodeCurrent = document.querySelector(".main-content");
    if (!nodeCurrent) {
      console.log("no main content found");
      return;
    }
    const loginuri = nodeCurrent.dataset.loginuri;
    if (!loginuri) {
      console.log("no loginuri found");
      return;
    }
    window.history.pushState({}, "", loginuri);
    window.addEventListener("popstate", function(event) {
      location.reload(true);
    });
  }
  function t(e2) {
    return t = "function" == typeof Symbol && "symbol" == typeof Symbol.iterator ? function(t2) {
      return typeof t2;
    } : function(t2) {
      return t2 && "function" == typeof Symbol && t2.constructor === Symbol && t2 !== Symbol.prototype ? "symbol" : typeof t2;
    }, t(e2);
  }
  function e(t2, e2) {
    if (!(t2 instanceof e2)) throw new TypeError("Cannot call a class as a function");
  }
  function n(t2, e2) {
    for (var n2 = 0; n2 < e2.length; n2++) {
      var r2 = e2[n2];
      r2.enumerable = r2.enumerable || false, r2.configurable = true, "value" in r2 && (r2.writable = true), Object.defineProperty(t2, r2.key, r2);
    }
  }
  function r(t2, e2) {
    return (function(t3) {
      if (Array.isArray(t3)) return t3;
    })(t2) || (function(t3, e3) {
      var n2 = null == t3 ? null : "undefined" != typeof Symbol && t3[Symbol.iterator] || t3["@@iterator"];
      if (null == n2) return;
      var r2, a2, o2 = [], i2 = true, u2 = false;
      try {
        for (n2 = n2.call(t3); !(i2 = (r2 = n2.next()).done) && (o2.push(r2.value), !e3 || o2.length !== e3); i2 = true) ;
      } catch (t4) {
        u2 = true, a2 = t4;
      } finally {
        try {
          i2 || null == n2.return || n2.return();
        } finally {
          if (u2) throw a2;
        }
      }
      return o2;
    })(t2, e2) || (function(t3, e3) {
      if (!t3) return;
      if ("string" == typeof t3) return a(t3, e3);
      var n2 = Object.prototype.toString.call(t3).slice(8, -1);
      "Object" === n2 && t3.constructor && (n2 = t3.constructor.name);
      if ("Map" === n2 || "Set" === n2) return Array.from(t3);
      if ("Arguments" === n2 || /^(?:Ui|I)nt(?:8|16|32)(?:Clamped)?Array$/.test(n2)) return a(t3, e3);
    })(t2, e2) || (function() {
      throw new TypeError("Invalid attempt to destructure non-iterable instance.\nIn order to be iterable, non-array objects must have a [Symbol.iterator]() method.");
    })();
  }
  function a(t2, e2) {
    (null == e2 || e2 > t2.length) && (e2 = t2.length);
    for (var n2 = 0, r2 = new Array(e2); n2 < e2; n2++) r2[n2] = t2[n2];
    return r2;
  }
  var o = { INVALID_PARAM_LANGUAGE: function(e2) {
    return "Invalid parameter for `language` provided. Expected a string, but got ".concat(t(e2), ".");
  }, INVALID_PARAM_JSON: function(e2) {
    return "Invalid parameter for `json` provided. Expected an object, but got ".concat(t(e2), ".");
  }, EMPTY_PARAM_LANGUAGE: function() {
    return "The parameter for `language` can't be an empty string.";
  }, EMPTY_PARAM_JSON: function() {
    return "The parameter for `json` must have at least one key/value pair.";
  }, INVALID_PARAM_KEY: function(e2) {
    return "Invalid parameter for `key` provided. Expected a string, but got ".concat(t(e2), ".");
  }, NO_LANGUAGE_REGISTERED: function(t2) {
    return 'No translation for language "'.concat(t2, '" has been added, yet. Make sure to register that language using the `.add()` method first.');
  }, TRANSLATION_NOT_FOUND: function(t2, e2) {
    return 'No translation found for key "'.concat(t2, '" in language "').concat(e2, '". Is there a key/value in your translation file?');
  }, INVALID_PARAMETER_SOURCES: function(e2) {
    return "Invalid parameter for `sources` provided. Expected either a string or an array, but got ".concat(t(e2), ".");
  }, FETCH_ERROR: function(t2) {
    return 'Could not fetch "'.concat(t2.url, '": ').concat(t2.status, " (").concat(t2.statusText, ")");
  }, INVALID_ENVIRONMENT: function() {
    return "You are trying to execute the method `translatePageTo()`, which is only available in the browser. Your environment is most likely Node.js";
  }, MODULE_NOT_FOUND: function(t2) {
    return t2;
  }, MISMATCHING_ATTRIBUTES: function(t2, e2, n2) {
    return 'The attributes "data-i18n" and "data-i18n-attr" must contain the same number of keys.\n\nValues in `data-i18n`:      ('.concat(t2.length, ") `").concat(t2.join(" "), "`\nValues in `data-i18n-attr`: (").concat(e2.length, ") `").concat(e2.join(" "), "`\n\nThe HTML element is:\n").concat(n2.outerHTML);
  }, INVALID_OPTIONS: function(e2) {
    return "Invalid config passed to the `Translator` constructor. Expected an object, but got ".concat(t(e2), ". Using default config instead.");
  } };
  function i(t2) {
    return function(e2) {
      if (t2) try {
        for (var n2 = o[e2], a2 = arguments.length, i2 = new Array(a2 > 1 ? a2 - 1 : 0), u2 = 1; u2 < a2; u2++) i2[u2 - 1] = arguments[u2];
        throw new TypeError(n2 ? n2.apply(void 0, i2) : "Unhandled Error");
      } catch (t3) {
        var s = t3.stack.split(/\n/g)[1], l2 = s.split(/@/), c = r(l2, 2), g2 = c[0], f = c[1];
        console.error("".concat(t3.message, "\n\nThis error happened in the method `").concat(g2, "` from: `").concat(f, "`.\n\nIf you don't want to see these error messages, turn off debugging by passing `{ debug: false }` to the constructor.\n\nError code: ").concat(e2, "\n\nCheck out the documentation for more details about the API:\nhttps://github.com/andreasremdt/simple-translator#usage\n        "));
      }
    };
  }
  var u = (function() {
    function r2() {
      var n2 = arguments.length > 0 && void 0 !== arguments[0] ? arguments[0] : {};
      e(this, r2), this.debug = i(true), ("object" != t(n2) || Array.isArray(n2)) && (this.debug("INVALID_OPTIONS", n2), n2 = {}), this.languages = /* @__PURE__ */ new Map(), this.config = Object.assign(r2.defaultConfig, n2);
      var a3 = this.config, o3 = a3.debug, u3 = a3.registerGlobally, s = a3.detectLanguage;
      this.debug = i(o3), u3 && (this._globalObject[u3] = this.translateForKey.bind(this)), s && "browser" == this._env && this._detectLanguage();
    }
    var a2, o2, u2;
    return a2 = r2, o2 = [{ key: "_globalObject", get: function() {
      return "browser" == this._env ? window : global;
    } }, { key: "_env", get: function() {
      return "undefined" != typeof window ? "browser" : "undefined" != typeof module && module.exports ? "node" : "browser";
    } }, { key: "_detectLanguage", value: function() {
      var t2 = window.localStorage ? localStorage.getItem(this.config.persistKey) : void 0;
      if (t2) this.config.defaultLanguage = t2;
      else {
        var e2 = navigator.languages ? navigator.languages[0] : navigator.language;
        this.config.defaultLanguage = e2.substr(0, 2);
      }
    } }, { key: "_getValueFromJSON", value: function(t2, e2) {
      var n2 = this.languages.get(e2);
      return t2.split(".").reduce((function(t3, e3) {
        return t3 ? t3[e3] : null;
      }), n2);
    } }, { key: "_replace", value: function(t2, e2) {
      var n2, r3, a3 = this, o3 = null === (n2 = t2.getAttribute("data-i18n")) || void 0 === n2 ? void 0 : n2.split(/\s/g), i2 = null == t2 || null === (r3 = t2.getAttribute("data-i18n-attr")) || void 0 === r3 ? void 0 : r3.split(/\s/g);
      i2 && o3.length != i2.length && this.debug("MISMATCHING_ATTRIBUTES", o3, i2, t2), o3.forEach((function(n3, r4) {
        var o4 = a3._getValueFromJSON(n3, e2), u3 = i2 ? i2[r4] : "innerHTML";
        o4 ? "innerHTML" == u3 ? t2[u3] = o4 : t2.setAttribute(u3, o4) : a3.debug("TRANSLATION_NOT_FOUND", n3, e2);
      }));
    } }, { key: "translatePageTo", value: function() {
      var t2 = this, e2 = arguments.length > 0 && void 0 !== arguments[0] ? arguments[0] : this.config.defaultLanguage;
      if ("node" != this._env) if ("string" == typeof e2) if (0 != e2.length) if (this.languages.has(e2)) {
        var n2 = "string" == typeof this.config.selector ? Array.from(document.querySelectorAll(this.config.selector)) : this.config.selector;
        n2.length && n2.length > 0 ? n2.forEach((function(n3) {
          return t2._replace(n3, e2);
        })) : null == n2.length && this._replace(n2, e2), this._currentLanguage = e2, document.documentElement.lang = e2, this.config.persist && window.localStorage && localStorage.setItem(this.config.persistKey, e2);
      } else this.debug("NO_LANGUAGE_REGISTERED", e2);
      else this.debug("EMPTY_PARAM_LANGUAGE");
      else this.debug("INVALID_PARAM_LANGUAGE", e2);
      else this.debug("INVALID_ENVIRONMENT");
    } }, { key: "translateForKey", value: function(t2) {
      var e2 = arguments.length > 1 && void 0 !== arguments[1] ? arguments[1] : this.config.defaultLanguage;
      if ("string" != typeof t2) return this.debug("INVALID_PARAM_KEY", t2), null;
      if (!this.languages.has(e2)) return this.debug("NO_LANGUAGE_REGISTERED", e2), null;
      var n2 = this._getValueFromJSON(t2, e2);
      return n2 || (this.debug("TRANSLATION_NOT_FOUND", t2, e2), null);
    } }, { key: "add", value: function(e2, n2) {
      return "string" != typeof e2 ? (this.debug("INVALID_PARAM_LANGUAGE", e2), this) : 0 == e2.length ? (this.debug("EMPTY_PARAM_LANGUAGE"), this) : Array.isArray(n2) || "object" != t(n2) ? (this.debug("INVALID_PARAM_JSON", n2), this) : 0 == Object.keys(n2).length ? (this.debug("EMPTY_PARAM_JSON"), this) : (this.languages.set(e2, n2), this);
    } }, { key: "remove", value: function(t2) {
      return "string" != typeof t2 ? (this.debug("INVALID_PARAM_LANGUAGE", t2), this) : 0 == t2.length ? (this.debug("EMPTY_PARAM_LANGUAGE"), this) : (this.languages.delete(t2), this);
    } }, { key: "fetch", value: (function(t2) {
      function e2(e3) {
        return t2.apply(this, arguments);
      }
      return e2.toString = function() {
        return t2.toString();
      }, e2;
    })((function(t2) {
      var e2 = this, n2 = !(arguments.length > 1 && void 0 !== arguments[1]) || arguments[1];
      if (!Array.isArray(t2) && "string" != typeof t2) return this.debug("INVALID_PARAMETER_SOURCES", t2), null;
      Array.isArray(t2) || (t2 = [t2]);
      var r3 = t2.map((function(t3) {
        var n3 = t3.replace(/\.json$/, "").replace(/^\//, ""), r4 = e2.config.filesLocation.replace(/\/$/, "");
        return "".concat(r4, "/").concat(n3, ".json");
      }));
      return "browser" == this._env ? Promise.all(r3.map((function(t3) {
        return fetch(t3);
      }))).then((function(t3) {
        return Promise.all(t3.map((function(t4) {
          if (t4.ok) return t4.json();
          e2.debug("FETCH_ERROR", t4);
        })));
      })).then((function(r4) {
        return r4 = r4.filter((function(t3) {
          return t3;
        })), n2 && r4.forEach((function(n3, r5) {
          e2.add(t2[r5], n3);
        })), r4.length > 1 ? r4 : r4[0];
      })) : "node" == this._env ? new Promise((function(a3) {
        var o3 = [];
        r3.forEach((function(r4, a4) {
          try {
            var i2 = JSON.parse(require("fs").readFileSync(process.cwd() + r4, "utf-8"));
            n2 && e2.add(t2[a4], i2), o3.push(i2);
          } catch (t3) {
            e2.debug("MODULE_NOT_FOUND", t3.message);
          }
        })), a3(o3.length > 1 ? o3 : o3[0]);
      })) : void 0;
    })) }, { key: "setDefaultLanguage", value: function(t2) {
      if ("string" == typeof t2) {
        if (0 != t2.length) return this.languages.has(t2) ? void (this.config.defaultLanguage = t2) : (this.debug("NO_LANGUAGE_REGISTERED", t2), null);
        this.debug("EMPTY_PARAM_LANGUAGE");
      } else this.debug("INVALID_PARAM_LANGUAGE", t2);
    } }, { key: "currentLanguage", get: function() {
      return this._currentLanguage || this.config.defaultLanguage;
    } }, { key: "defaultLanguage", get: function() {
      return this.config.defaultLanguage;
    } }], u2 = [{ key: "defaultConfig", get: function() {
      return { defaultLanguage: "en", detectLanguage: true, selector: "[data-i18n]", debug: false, registerGlobally: "__", persist: false, persistKey: "preferred_language", filesLocation: "/i18n" };
    } }], o2 && n(a2.prototype, o2), u2 && n(a2, u2), r2;
  })();
  function initTranslator() {
    let locales = /* @__PURE__ */ new Map();
    var translator = new u({ debug: true });
    window.translator = translator;
    const links = document.head.querySelectorAll("link[i18n]");
    if (links.length > 0) {
      for (const link of links) {
        const href = link.href;
        const locale = link.getAttribute("i18n");
        locales.set(locale, href);
      }
    }
    console.log("locales", locales);
    translatePageWithLocaleCookie(translator, locales);
  }
  const translationCache = /* @__PURE__ */ new Map();
  async function translatePageWithLocaleCookie(translator, locales) {
    var _a;
    const locale = (_a = document.cookie.split("; ").find((row) => row.startsWith("locale="))) == null ? void 0 : _a.split("=")[1];
    if (locale) {
      const localeFile = locales.get(locale);
      if (localeFile) {
        const data = await loadTranslationFile(locale, localeFile);
        translator.add(locale, data).translatePageTo(locale);
      }
    }
  }
  async function loadTranslationFile(locale, localeFile) {
    try {
      if (translationCache.has(locale)) {
        return translationCache.get(locale);
      }
      const response = await fetch(localeFile, {
        cache: "force-cache"
        // Use browser cache
      });
      if (!response.ok) {
        throw new Error(`Failed to fetch translation file: ${response.status} ${response.statusText}`);
      }
      return await response.json();
    } catch (error) {
      console.error("Error loading translation file:", error);
      try {
        const response = await fetch(localeFile);
        return await response.json();
      } catch (fallbackError) {
        console.error("Fallback translation loading also failed:", fallbackError);
      }
    }
  }
  const nodeHandlers = {
    "emailOTP": initEmailOTP,
    "passwordOrSocialLogin": initPasswordOrSocialLogin,
    "hcaptcha": initHcaptcha,
    "verifyPasskey": initVerifyPasskey,
    "registerPasskey": initRegisterPasskey,
    "onboardingWithPasskey": initOnboardingWithPasskey,
    "telegramLogin": initTelegramLogin
    // Add more node handlers here as needed
  };
  document.addEventListener("DOMContentLoaded", function() {
    console.log("DOMContentLoaded");
    initTranslator();
    const mainContent = document.querySelector(".main-content");
    if (mainContent) {
      const nodeName = mainContent.dataset.node;
      const handler = nodeHandlers[nodeName];
      if (handler) {
        handler();
        console.log("Initialized node:", nodeName);
      }
    }
    initNodeHistory();
  });
}));
