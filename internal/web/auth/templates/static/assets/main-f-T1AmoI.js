var commonjsGlobal = typeof globalThis !== "undefined" ? globalThis : typeof window !== "undefined" ? window : typeof global !== "undefined" ? global : typeof self !== "undefined" ? self : {};
var _freeGlobal;
var hasRequired_freeGlobal;
function require_freeGlobal() {
  if (hasRequired_freeGlobal) return _freeGlobal;
  hasRequired_freeGlobal = 1;
  var freeGlobal = typeof commonjsGlobal == "object" && commonjsGlobal && commonjsGlobal.Object === Object && commonjsGlobal;
  _freeGlobal = freeGlobal;
  return _freeGlobal;
}
var _root;
var hasRequired_root;
function require_root() {
  if (hasRequired_root) return _root;
  hasRequired_root = 1;
  var freeGlobal = require_freeGlobal();
  var freeSelf = typeof self == "object" && self && self.Object === Object && self;
  var root = freeGlobal || freeSelf || Function("return this")();
  _root = root;
  return _root;
}
var _Symbol;
var hasRequired_Symbol;
function require_Symbol() {
  if (hasRequired_Symbol) return _Symbol;
  hasRequired_Symbol = 1;
  var root = require_root();
  var Symbol = root.Symbol;
  _Symbol = Symbol;
  return _Symbol;
}
var _getRawTag;
var hasRequired_getRawTag;
function require_getRawTag() {
  if (hasRequired_getRawTag) return _getRawTag;
  hasRequired_getRawTag = 1;
  var Symbol = require_Symbol();
  var objectProto = Object.prototype;
  var hasOwnProperty = objectProto.hasOwnProperty;
  var nativeObjectToString = objectProto.toString;
  var symToStringTag = Symbol ? Symbol.toStringTag : void 0;
  function getRawTag(value) {
    var isOwn = hasOwnProperty.call(value, symToStringTag), tag = value[symToStringTag];
    try {
      value[symToStringTag] = void 0;
      var unmasked = true;
    } catch (e) {
    }
    var result = nativeObjectToString.call(value);
    if (unmasked) {
      if (isOwn) {
        value[symToStringTag] = tag;
      } else {
        delete value[symToStringTag];
      }
    }
    return result;
  }
  _getRawTag = getRawTag;
  return _getRawTag;
}
var _objectToString;
var hasRequired_objectToString;
function require_objectToString() {
  if (hasRequired_objectToString) return _objectToString;
  hasRequired_objectToString = 1;
  var objectProto = Object.prototype;
  var nativeObjectToString = objectProto.toString;
  function objectToString(value) {
    return nativeObjectToString.call(value);
  }
  _objectToString = objectToString;
  return _objectToString;
}
var _baseGetTag;
var hasRequired_baseGetTag;
function require_baseGetTag() {
  if (hasRequired_baseGetTag) return _baseGetTag;
  hasRequired_baseGetTag = 1;
  var Symbol = require_Symbol(), getRawTag = require_getRawTag(), objectToString = require_objectToString();
  var nullTag = "[object Null]", undefinedTag = "[object Undefined]";
  var symToStringTag = Symbol ? Symbol.toStringTag : void 0;
  function baseGetTag(value) {
    if (value == null) {
      return value === void 0 ? undefinedTag : nullTag;
    }
    return symToStringTag && symToStringTag in Object(value) ? getRawTag(value) : objectToString(value);
  }
  _baseGetTag = baseGetTag;
  return _baseGetTag;
}
var isObjectLike_1;
var hasRequiredIsObjectLike;
function requireIsObjectLike() {
  if (hasRequiredIsObjectLike) return isObjectLike_1;
  hasRequiredIsObjectLike = 1;
  function isObjectLike(value) {
    return value != null && typeof value == "object";
  }
  isObjectLike_1 = isObjectLike;
  return isObjectLike_1;
}
var isSymbol_1;
var hasRequiredIsSymbol;
function requireIsSymbol() {
  if (hasRequiredIsSymbol) return isSymbol_1;
  hasRequiredIsSymbol = 1;
  var baseGetTag = require_baseGetTag(), isObjectLike = requireIsObjectLike();
  var symbolTag = "[object Symbol]";
  function isSymbol(value) {
    return typeof value == "symbol" || isObjectLike(value) && baseGetTag(value) == symbolTag;
  }
  isSymbol_1 = isSymbol;
  return isSymbol_1;
}
var _baseToNumber;
var hasRequired_baseToNumber;
function require_baseToNumber() {
  if (hasRequired_baseToNumber) return _baseToNumber;
  hasRequired_baseToNumber = 1;
  var isSymbol = requireIsSymbol();
  var NAN = 0 / 0;
  function baseToNumber(value) {
    if (typeof value == "number") {
      return value;
    }
    if (isSymbol(value)) {
      return NAN;
    }
    return +value;
  }
  _baseToNumber = baseToNumber;
  return _baseToNumber;
}
var _arrayMap;
var hasRequired_arrayMap;
function require_arrayMap() {
  if (hasRequired_arrayMap) return _arrayMap;
  hasRequired_arrayMap = 1;
  function arrayMap(array, iteratee) {
    var index = -1, length = array == null ? 0 : array.length, result = Array(length);
    while (++index < length) {
      result[index] = iteratee(array[index], index, array);
    }
    return result;
  }
  _arrayMap = arrayMap;
  return _arrayMap;
}
var isArray_1;
var hasRequiredIsArray;
function requireIsArray() {
  if (hasRequiredIsArray) return isArray_1;
  hasRequiredIsArray = 1;
  var isArray = Array.isArray;
  isArray_1 = isArray;
  return isArray_1;
}
var _baseToString;
var hasRequired_baseToString;
function require_baseToString() {
  if (hasRequired_baseToString) return _baseToString;
  hasRequired_baseToString = 1;
  var Symbol = require_Symbol(), arrayMap = require_arrayMap(), isArray = requireIsArray(), isSymbol = requireIsSymbol();
  var symbolProto = Symbol ? Symbol.prototype : void 0, symbolToString = symbolProto ? symbolProto.toString : void 0;
  function baseToString(value) {
    if (typeof value == "string") {
      return value;
    }
    if (isArray(value)) {
      return arrayMap(value, baseToString) + "";
    }
    if (isSymbol(value)) {
      return symbolToString ? symbolToString.call(value) : "";
    }
    var result = value + "";
    return result == "0" && 1 / value == -Infinity ? "-0" : result;
  }
  _baseToString = baseToString;
  return _baseToString;
}
var _createMathOperation;
var hasRequired_createMathOperation;
function require_createMathOperation() {
  if (hasRequired_createMathOperation) return _createMathOperation;
  hasRequired_createMathOperation = 1;
  var baseToNumber = require_baseToNumber(), baseToString = require_baseToString();
  function createMathOperation(operator, defaultValue) {
    return function(value, other) {
      var result;
      if (value === void 0 && other === void 0) {
        return defaultValue;
      }
      if (value !== void 0) {
        result = value;
      }
      if (other !== void 0) {
        if (result === void 0) {
          return other;
        }
        if (typeof value == "string" || typeof other == "string") {
          value = baseToString(value);
          other = baseToString(other);
        } else {
          value = baseToNumber(value);
          other = baseToNumber(other);
        }
        result = operator(value, other);
      }
      return result;
    };
  }
  _createMathOperation = createMathOperation;
  return _createMathOperation;
}
var add_1;
var hasRequiredAdd;
function requireAdd() {
  if (hasRequiredAdd) return add_1;
  hasRequiredAdd = 1;
  var createMathOperation = require_createMathOperation();
  var add = createMathOperation(function(augend, addend) {
    return augend + addend;
  }, 0);
  add_1 = add;
  return add_1;
}
requireAdd();
function initEmailOTP() {
  var _a;
  const form = document.getElementById("otpForm");
  const basicInput = document.getElementById("otp");
  if (!form || !basicInput) {
    console.error("Required form elements not found");
    return;
  }
  const otpInputs = document.createElement("div");
  otpInputs.className = "otp-inputs";
  for (let i = 0; i < 6; i++) {
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
    input.addEventListener("keydown", function(e) {
      if (e.key === "Backspace" && !this.value && index > 0) {
        inputs[index - 1].focus();
      }
    });
    input.addEventListener("paste", function(e) {
      var _a2;
      e.preventDefault();
      const pastedData = ((_a2 = e.clipboardData) == null ? void 0 : _a2.getData("text").slice(0, 6)) || "";
      if (/^\d+$/.test(pastedData)) {
        pastedData.split("").forEach((digit, i) => {
          if (inputs[i]) {
            inputs[i].value = digit;
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
  form.addEventListener("submit", function(e) {
    e.preventDefault();
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
    const r = cred.response;
    if (r.clientDataJSON) json.response.clientDataJSON = bufferToBase64url(r.clientDataJSON);
    if (r.attestationObject) json.response.attestationObject = bufferToBase64url(r.attestationObject);
    if (r.authenticatorData) json.response.authenticatorData = bufferToBase64url(r.authenticatorData);
    if (r.signature) json.response.signature = bufferToBase64url(r.signature);
    if (r.userHandle) json.response.userHandle = bufferToBase64url(r.userHandle);
    if (r.publicKey) json.response.publicKey = r.publicKey;
    if (r.publicKeyAlgorithm) json.response.publicKeyAlgorithm = r.publicKeyAlgorithm;
    if (r.transports) json.response.transports = r.transports;
  }
  return json;
}
async function initPasswordOrSocialLogin() {
  console.log("initializing PasswordOrSocialLogin");
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
var I = (t, e, r) => e in t ? K(t, e, { enumerable: true, configurable: true, writable: true, value: r }) : t[e] = r, d = (t, e) => {
  for (var r in e || (e = {})) P.call(e, r) && I(t, r, e[r]);
  if (y) for (var r of y(e)) D.call(e, r) && I(t, r, e[r]);
  return t;
}, O = (t, e) => Y(t, V(e));
var M = (t, e) => {
  var r = {};
  for (var n in t) P.call(t, n) && e.indexOf(n) < 0 && (r[n] = t[n]);
  if (t != null && y) for (var n of y(t)) e.indexOf(n) < 0 && D.call(t, n) && (r[n] = t[n]);
  return r;
};
var l = (t, e, r) => (I(t, typeof e != "symbol" ? e + "" : e, r), r);
var x = (t, e, r) => new Promise((n, s) => {
  var a = (c) => {
    try {
      o(r.next(c));
    } catch (i) {
      s(i);
    }
  }, u = (c) => {
    try {
      o(r.throw(c));
    } catch (i) {
      s(i);
    }
  }, o = (c) => c.done ? n(c.value) : Promise.resolve(c.value).then(a, u);
  o((r = r.apply(t, e)).next());
});
var L = "hCaptcha-script", b = "hCaptchaOnLoad", _ = "script-error";
var g = "@hCaptcha/loader";
function U(t) {
  return Object.entries(t).filter(([, e]) => e || e === false).map(([e, r]) => `${encodeURIComponent(e)}=${encodeURIComponent(String(r))}`).join("&");
}
function R(t) {
  let e = t && t.ownerDocument || document, r = e.defaultView || e.parentWindow || window;
  return { document: e, window: r };
}
function S(t) {
  return t || document.head;
}
function F(t) {
  var e;
  t.setTag("source", g), t.setTag("url", document.URL), t.setContext("os", { UA: navigator.userAgent }), t.setContext("browser", d({}, z())), t.setContext("device", O(d({}, J()), { screen_width_pixels: screen.width, screen_height_pixels: screen.height, language: navigator.language, orientation: ((e = screen.orientation) == null ? void 0 : e.type) || "Unknown", processor_count: navigator.hardwareConcurrency, platform: navigator.platform }));
}
function z() {
  var n, s, a, u, o, c;
  let t = navigator.userAgent, e, r;
  return t.indexOf("Firefox") !== -1 ? (e = "Firefox", r = (n = t.match(/Firefox\/([\d.]+)/)) == null ? void 0 : n[1]) : t.indexOf("Edg") !== -1 ? (e = "Microsoft Edge", r = (s = t.match(/Edg\/([\d.]+)/)) == null ? void 0 : s[1]) : t.indexOf("Chrome") !== -1 && t.indexOf("Safari") !== -1 ? (e = "Chrome", r = (a = t.match(/Chrome\/([\d.]+)/)) == null ? void 0 : a[1]) : t.indexOf("Safari") !== -1 && t.indexOf("Chrome") === -1 ? (e = "Safari", r = (u = t.match(/Version\/([\d.]+)/)) == null ? void 0 : u[1]) : t.indexOf("Opera") !== -1 || t.indexOf("OPR") !== -1 ? (e = "Opera", r = (o = t.match(/(Opera|OPR)\/([\d.]+)/)) == null ? void 0 : o[2]) : t.indexOf("MSIE") !== -1 || t.indexOf("Trident") !== -1 ? (e = "Internet Explorer", r = (c = t.match(/(MSIE |rv:)([\d.]+)/)) == null ? void 0 : c[2]) : (e = "Unknown", r = "Unknown"), { name: e, version: r };
}
function J() {
  let t = navigator.userAgent, e;
  t.indexOf("Win") !== -1 ? e = "Windows" : t.indexOf("Mac") !== -1 ? e = "Mac" : t.indexOf("Linux") !== -1 ? e = "Linux" : t.indexOf("Android") !== -1 ? e = "Android" : t.indexOf("like Mac") !== -1 || t.indexOf("iPhone") !== -1 || t.indexOf("iPad") !== -1 ? e = "iOS" : e = "Unknown";
  let r;
  return /Mobile|iPhone|iPod|Android/i.test(t) ? r = "Mobile" : /Tablet|iPad/i.test(t) ? r = "Tablet" : r = "Desktop", { model: e, family: e, device: r };
}
var Q = class B {
  constructor(e) {
    l(this, "_parent");
    l(this, "breadcrumbs", []);
    l(this, "context", {});
    l(this, "extra", {});
    l(this, "tags", {});
    l(this, "request");
    l(this, "user");
    this._parent = e;
  }
  get parent() {
    return this._parent;
  }
  child() {
    return new B(this);
  }
  setRequest(e) {
    return this.request = e, this;
  }
  removeRequest() {
    return this.request = void 0, this;
  }
  addBreadcrumb(e) {
    return typeof e.timestamp > "u" && (e.timestamp = (/* @__PURE__ */ new Date()).toISOString()), this.breadcrumbs.push(e), this;
  }
  setExtra(e, r) {
    return this.extra[e] = r, this;
  }
  removeExtra(e) {
    return delete this.extra[e], this;
  }
  setContext(e, r) {
    return typeof r.type > "u" && (r.type = e), this.context[e] = r, this;
  }
  removeContext(e) {
    return delete this.context[e], this;
  }
  setTags(e) {
    return this.tags = d(d({}, this.tags), e), this;
  }
  setTag(e, r) {
    return this.tags[e] = r, this;
  }
  removeTag(e) {
    return delete this.tags[e], this;
  }
  setUser(e) {
    return this.user = e, this;
  }
  removeUser() {
    return this.user = void 0, this;
  }
  toBody() {
    let e = [], r = this;
    for (; r; ) e.push(r), r = r.parent;
    return e.reverse().reduce((n, s) => {
      var a;
      return n.breadcrumbs = [...(a = n.breadcrumbs) != null ? a : [], ...s.breadcrumbs], n.extra = d(d({}, n.extra), s.extra), n.contexts = d(d({}, n.contexts), s.context), n.tags = d(d({}, n.tags), s.tags), s.user && (n.user = s.user), s.request && (n.request = s.request), n;
    }, { breadcrumbs: [], extra: {}, contexts: {}, tags: {}, request: void 0, user: void 0 });
  }
  clear() {
    this.breadcrumbs = [], this.context = {}, this.tags = {}, this.user = void 0;
  }
}, Z = /^\s*at (?:(.*?) ?\()?((?:file|https?|blob|chrome-extension|address|native|eval|webpack|<anonymous>|[-a-z]+:|.*bundle|\/).*?)(?::(\d+))?(?::(\d+))?\)?\s*$/i, ee = /^\s*(.*?)(?:\((.*?)\))?(?:^|@)?((?:file|https?|blob|chrome|webpack|resource|moz-extension).*?:\/.*?|\[native code\]|[^@]*(?:bundle|\d+\.js))(?::(\d+))?(?::(\d+))?\s*$/i, te = /^\s*at (?:((?:\[object object\])?.+) )?\(?((?:file|ms-appx|https?|webpack|blob):.*?):(\d+)(?::(\d+))?\)?\s*$/i, re = /^(?:(\w+):)\/\/(?:(\w+)(?::(\w+))?@)([\w.-]+)(?::(\d+))?\/(.+)/, N = "?", k = "An unknown error occurred", ne = "0.0.4";
function se(t) {
  for (let e = 0; e < t.length; e++) t[e] = Math.floor(Math.random() * 256);
  return t;
}
function p(t) {
  return (t + 256).toString(16).substring(1);
}
function oe() {
  let t = se(new Array(16));
  return t[6] = t[6] & 15 | 64, t[8] = t[8] & 63 | 128, p(t[0]) + p(t[1]) + p(t[2]) + p(t[3]) + "-" + p(t[4]) + p(t[5]) + "-" + p(t[6]) + p(t[7]) + "-" + p(t[8]) + p(t[9]) + "-" + p(t[10]) + p(t[11]) + p(t[12]) + p(t[13]) + p(t[14]) + p(t[15]);
}
var ie = [[Z, "chrome"], [te, "winjs"], [ee, "gecko"]];
function ae(t) {
  var n, s, a, u;
  if (!t.stack) return null;
  let e = [], r = (a = (s = (n = t.stack).split) == null ? void 0 : s.call(n, `
`)) != null ? a : [];
  for (let o = 0; o < r.length; ++o) {
    let c = null, i = null, f = null;
    for (let [v, w] of ie) if (i = v.exec(r[o]), i) {
      f = w;
      break;
    }
    if (!(!i || !f)) {
      if (f === "chrome") c = { filename: (u = i[2]) != null && u.startsWith("address at ") ? i[2].substring(11) : i[2], function: i[1] || N, lineno: i[3] ? +i[3] : null, colno: i[4] ? +i[4] : null };
      else if (f === "winjs") c = { filename: i[2], function: i[1] || N, lineno: +i[3], colno: i[4] ? +i[4] : null };
      else if (f === "gecko") o === 0 && !i[5] && t.columnNumber !== void 0 && e.length > 0 && (e[0].column = t.columnNumber + 1), c = { filename: i[3], function: i[1] || N, lineno: i[4] ? +i[4] : null, colno: i[5] ? +i[5] : null };
      else continue;
      !c.function && c.lineno && (c.function = N), e.push(c);
    }
  }
  return e.length ? e.reverse() : null;
}
function ce(t) {
  let e = ae(t);
  return { type: t.name, value: t.message, stacktrace: { frames: e != null ? e : [] } };
}
function ue(t) {
  let e = re.exec(t), r = e ? e.slice(1) : [];
  if (r.length !== 6) throw new Error("Invalid DSN");
  let n = r[5].split("/"), s = n.slice(0, -1).join("/");
  return r[0] + "://" + r[3] + (r[4] ? ":" + r[4] : "") + (s ? "/" + s : "") + "/api/" + n.pop() + "/envelope/?sentry_version=7&sentry_key=" + r[1] + (r[2] ? "&sentry_secret=" + r[2] : "");
}
function le(t, e, r) {
  var s, a;
  let n = d({ event_id: oe().replaceAll("-", ""), platform: "javascript", sdk: { name: "@hcaptcha/sentry", version: ne }, environment: e, release: r, timestamp: Date.now() / 1e3 }, t.scope.toBody());
  if (t.type === "exception") {
    n.message = (a = (s = t.error) == null ? void 0 : s.message) != null ? a : "Unknown error", n.fingerprint = [n.message];
    let u = [], o = t.error;
    for (let c = 0; c < 5 && o && (u.push(ce(o)), !(!o.cause || !(o.cause instanceof Error))); c++) o = o.cause;
    n.exception = { values: u.reverse() };
  }
  return t.type === "message" && (n.message = t.message, n.level = t.level), n;
}
function de(t) {
  if (t instanceof Error) return t;
  if (typeof t == "string") return new Error(t);
  if (typeof t == "object" && t !== null && !Array.isArray(t)) {
    let r = t, { message: n } = r, s = M(r, ["message"]), a = new Error(typeof n == "string" ? n : k);
    return Object.assign(a, s);
  }
  let e = new Error(k);
  return Object.assign(e, { cause: t });
}
function pe(t, e, r) {
  return x(this, null, function* () {
    var n, s;
    try {
      if (typeof fetch < "u" && typeof AbortSignal < "u") {
        let a;
        if (r) {
          let c = new AbortController();
          a = c.signal, setTimeout(() => c.abort(), r);
        }
        let u = yield fetch(t, O(d({}, e), { signal: a })), o = yield u.text();
        return { status: u.status, body: o };
      }
      return yield new Promise((a, u) => {
        var c, i;
        let o = new XMLHttpRequest();
        if (o.open((c = e == null ? void 0 : e.method) != null ? c : "GET", t), o.onload = () => a({ status: o.status, body: o.responseText }), o.onerror = () => u(new Error("XHR Network Error")), e == null ? void 0 : e.headers) for (let [f, v] of Object.entries(e.headers)) o.setRequestHeader(f, v);
        if (r) {
          let f = setTimeout(() => {
            o.abort(), u(new Error("Request timed out"));
          }, r);
          o.onloadend = () => {
            clearTimeout(f);
          };
        }
        o.send((i = e == null ? void 0 : e.body) == null ? void 0 : i.toString());
      });
    } catch (a) {
      return { status: 0, body: (s = (n = a == null ? void 0 : a.toString) == null ? void 0 : n.call(a)) != null ? s : "Unknown error" };
    }
  });
}
var h, A = (h = class {
  constructor(e) {
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
    var r, n, s, a, u;
    this.environment = e.environment, this.release = e.release, this.sampleRate = (r = e.sampleRate) != null ? r : 1, this.debug = (n = e.debug) != null ? n : false, this._scope = (s = e.scope) != null ? s : new Q(), this.apiURL = ue(e.dsn), this.dsn = e.dsn, this.shouldBuffer = (a = e.buffer) != null ? a : false, this.bufferlimit = (u = e.bufferLimit) != null ? u : 20;
  }
  static init(e) {
    h._instance || (h._instance = new h(e));
  }
  static get instance() {
    if (!h._instance) throw new Error("Sentry has not been initialized");
    return h._instance;
  }
  log(...e) {
    this.debug && console.log(...e);
  }
  get scope() {
    return this._scope;
  }
  static get scope() {
    return h.instance.scope;
  }
  withScope(e) {
    let r = this._scope.child();
    e(r);
  }
  static withScope(e) {
    h.instance.withScope(e);
  }
  captureException(e, r) {
    this.captureEvent({ type: "exception", level: "error", error: de(e), scope: r != null ? r : this._scope });
  }
  static captureException(e, r) {
    h.instance.captureException(e, r);
  }
  captureMessage(e, r = "info", n) {
    this.captureEvent({ type: "message", level: r, message: e, scope: n != null ? n : this._scope });
  }
  static captureMessage(e, r = "info", n) {
    h.instance.captureMessage(e, r, n);
  }
  captureEvent(e) {
    if (Math.random() >= this.sampleRate) {
      this.log("Dropped event due to sample rate");
      return;
    }
    if (this.shouldBuffer) {
      if (this.buffer.length >= this.bufferlimit) return;
      this.buffer.push(e);
    } else this.sendEvent(e);
  }
  sendEvent(e, r = 5e3) {
    return x(this, null, function* () {
      try {
        this.log("Sending sentry event", e);
        let n = le(e, this.environment, this.release), s = { event_id: n.event_id, dsn: this.dsn }, a = { type: "event" }, u = JSON.stringify(s) + `
` + JSON.stringify(a) + `
` + JSON.stringify(n), o = yield pe(this.apiURL, { method: "POST", headers: { "Content-Type": "application/x-sentry-envelope" }, body: u }, r);
        this.log("Sentry response", o.status), o.status !== 200 && (console.log(o.body), console.error("Failed to send event to Sentry", o));
      } catch (n) {
        console.error("Failed to send event", n);
      }
    });
  }
  flush(e = 5e3) {
    return x(this, null, function* () {
      try {
        this.log("Flushing sentry events", this.buffer.length);
        let r = this.buffer.splice(0, this.buffer.length).map((n) => this.sendEvent(n, e));
        yield Promise.all(r), this.log("Flushed all events");
      } catch (r) {
        console.error("Failed to flush events", r);
      }
    });
  }
  static flush(e = 5e3) {
    return h.instance.flush(e);
  }
  static reset() {
    h._instance = void 0;
  }
}, l(h, "_instance"), h);
var he = "https://d233059272824702afc8c43834c4912d@sentry.hcaptcha.com/6", fe = "2.0.0", me = "production";
function j(t = true) {
  if (!t) return q();
  A.init({ dsn: he, release: fe, environment: me });
  let e = A.scope;
  return F(e), q(e);
}
function q(t = null) {
  return { addBreadcrumb: (e) => {
    t && t.addBreadcrumb(e);
  }, captureRequest: (e) => {
    t && t.setRequest(e);
  }, captureException: (e) => {
    t && A.captureException(e, t);
  } };
}
function H({ scriptLocation: t, query: e, loadAsync: r = true, crossOrigin: n, apihost: s = "https://js.hcaptcha.com", cleanup: a = true, secureApi: u = false, scriptSource: o = "" } = {}, c) {
  let i = S(t), f = R(i);
  return new Promise((v, w) => {
    let m = f.document.createElement("script");
    m.id = L, o ? m.src = `${o}?onload=${b}` : u ? m.src = `${s}/1/secure-api.js?onload=${b}` : m.src = `${s}/1/api.js?onload=${b}`, m.crossOrigin = n, m.async = r;
    let T = (E, X) => {
      try {
        !u && a && i.removeChild(m), X(E);
      } catch (G) {
        w(G);
      }
    };
    m.onload = (E) => T(E, v), m.onerror = (E) => {
      c && c(m.src), T(E, w);
    }, m.src += e !== "" ? `&${e}` : "", i.appendChild(m);
  });
}
var C = [];
function ge(t = { cleanup: true }, e) {
  try {
    e.addBreadcrumb({ category: g, message: "hCaptcha loader params", data: t });
    let r = S(t.scriptLocation), n = R(r), s = C.find(({ scope: u }) => u === n.window);
    if (s) return e.addBreadcrumb({ category: g, message: "hCaptcha already loaded" }), s.promise;
    let a = new Promise((u, o) => x(this, null, function* () {
      try {
        n.window[b] = () => {
          e.addBreadcrumb({ category: g, message: "hCaptcha script called onload function" }), u(n.window.hcaptcha);
        };
        let c = U({ custom: t.custom, render: t.render, sentry: t.sentry, assethost: t.assethost, imghost: t.imghost, reportapi: t.reportapi, endpoint: t.endpoint, host: t.host, recaptchacompat: t.recaptchacompat, hl: t.hl });
        yield H(d({ query: c }, t), (i) => {
          e.captureRequest({ url: i, method: "GET" });
        }), e.addBreadcrumb({ category: g, message: "hCaptcha loaded", data: s });
      } catch (c) {
        e.addBreadcrumb({ category: g, message: "hCaptcha failed to load" });
        let i = C.findIndex((f) => f.scope === n.window);
        i !== -1 && C.splice(i, 1), o(new Error(_));
      }
    }));
    return C.push({ promise: a, scope: n.window }), a;
  } catch (r) {
    return e.captureException(r), Promise.reject(new Error(_));
  }
}
function $(t, e, r = 0) {
  return x(this, null, function* () {
    let n = r < 2 ? "Retry loading hCaptcha Api" : "Exceeded maximum retries";
    try {
      return yield ge(t, e);
    } catch (s) {
      return e.addBreadcrumb({ category: g, message: n }), r >= 2 ? (e.captureException(s), Promise.reject(s)) : (r += 1, $(t, e, r));
    }
  });
}
function xe() {
  return x(this, arguments, function* (t = {}) {
    let e = j(t.sentry);
    return yield $(t, e);
  });
}
async function initHcaptcha() {
  console.log("initializing Hcaptcha");
  await xe();
  await xe();
  hcaptcha.render({
    sitekey: "<your_sitekey>"
  });
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
const nodeHandlers = {
  "emailOTP": initEmailOTP,
  "passwordOrSocialLogin": initPasswordOrSocialLogin,
  "hcaptcha": initHcaptcha,
  "verifyPasskey": initVerifyPasskey,
  "registerPasskey": initRegisterPasskey
  // Add more node handlers here as needed
};
document.addEventListener("DOMContentLoaded", function() {
  const mainContent = document.querySelector(".main-content");
  if (mainContent) {
    const nodeName = mainContent.dataset.node;
    const handler = nodeHandlers[nodeName];
    if (handler) {
      handler();
      console.log("Initialized node:", nodeName);
    }
  }
});
