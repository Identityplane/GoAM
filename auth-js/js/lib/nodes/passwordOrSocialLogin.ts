export async function initPasswordOrSocialLogin(): Promise<void> {

    console.log('initializing PasswordOrSocialLogin')

    function base64urlToBuffer(base64url: string) {
        const padding = '='.repeat((4 - base64url.length % 4) % 4);
        const base64 = (base64url + padding)
            .replace(/-/g, '+')
            .replace(/_/g, '/');
        const raw = atob(base64);
        return new Uint8Array([...raw].map(char => char.charCodeAt(0)));
    }

    function bufferToBase64url(buffer: Uint8Array) {
        const binary = String.fromCharCode.apply(null, new Uint8Array(buffer));
        const base64 = btoa(binary);
        return base64.replace(/\+/g, '-').replace(/\//g, '_').replace(/=/g, '');
    }

    // Availability of `window.PublicKeyCredential` means WebAuthn is usable.  
    if (window.PublicKeyCredential && PublicKeyCredential.isConditionalMediationAvailable) {
        // Check if conditional mediation is available.  
        const isCMA = await PublicKeyCredential.isConditionalMediationAvailable();
        if (isCMA) {


            const optionsInput = document.getElementById("passkeysLoginOptions").value;
            const options = JSON.parse(optionsInput);

            options.publicKey.challenge = base64urlToBuffer(options.publicKey.challenge);

            const credential = await navigator.credentials.get({
                publicKey: options.publicKey,
                mediation: 'conditional'
            });

            console.log(credential);
        }
    }

}