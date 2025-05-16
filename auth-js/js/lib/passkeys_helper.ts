export function base64urlToBuffer(base64url: string) {
    const padding = '='.repeat((4 - base64url.length % 4) % 4);
    const base64 = (base64url + padding)
        .replace(/-/g, '+')
        .replace(/_/g, '/');
    const raw = atob(base64);
    return new Uint8Array([...raw].map(char => char.charCodeAt(0)));
}

export function bufferToBase64url(buffer: Uint8Array) {
    const binary = String.fromCharCode.apply(null, new Uint8Array(buffer));
    const base64 = btoa(binary);
    return base64.replace(/\+/g, '-').replace(/\//g, '_').replace(/=/g, '');
}

export function serializeCredential(cred) {
    if (!cred) return null;
  
    const json = {
      id: cred.id,
      type: cred.type,
      rawId: bufferToBase64url(cred.rawId),
      authenticatorAttachment: cred.authenticatorAttachment || null,
      clientExtensionResults: cred.getClientExtensionResults?.() || {},
      response: {}
    };
  
    if (cred.response) {
      const r = cred.response;
      if (r.clientDataJSON) json.response.clientDataJSON = bufferToBase64url(r.clientDataJSON);
      if (r.attestationObject) json.response.attestationObject = bufferToBase64url(r.attestationObject); // registration
      if (r.authenticatorData) json.response.authenticatorData = bufferToBase64url(r.authenticatorData); // login
      if (r.signature) json.response.signature = bufferToBase64url(r.signature);                         // login
      if (r.userHandle) json.response.userHandle = bufferToBase64url(r.userHandle);                     // login
      if (r.publicKey) json.response.publicKey = r.publicKey;                                           // registration
      if (r.publicKeyAlgorithm) json.response.publicKeyAlgorithm = r.publicKeyAlgorithm;               // registration
      if (r.transports) json.response.transports = r.transports;                                        // registration
    }
  
    return json;
  }