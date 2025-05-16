import { base64urlToBuffer, serializeCredential } from "../passkeys_helper";

export async function initPasswordOrSocialLogin(): Promise<void> {

    console.log('initializing PasswordOrSocialLogin')

    // Availability of `window.PublicKeyCredential` means WebAuthn is usable.  
    if (window.PublicKeyCredential && PublicKeyCredential.isConditionalMediationAvailable) {
        // Check if conditional mediation is available.  
        const isCMA = await PublicKeyCredential.isConditionalMediationAvailable();
        if (isCMA) {


            const optionsInput = document.getElementById("passkeysLoginOptions").value;
            const options = JSON.parse(optionsInput);

            options.publicKey.challenge = base64urlToBuffer(options.publicKey.challenge);

            const assertion = await navigator.credentials.get({
                publicKey: options.publicKey,
                mediation: 'conditional'
            });

            const serializedAssertion = serializeCredential(assertion);
            console.log(serializedAssertion);
            document.getElementById("passkeysFinishLoginJson").value = JSON.stringify(serializedAssertion);
            document.getElementById("passkey-form").submit();
        }
    }

}