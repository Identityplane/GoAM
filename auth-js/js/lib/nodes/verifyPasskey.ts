import { base64urlToBuffer, serializeCredential } from "../passkeys_helper";

export async function initVerifyPasskey(): Promise<void> {

    console.log('initializing VerifyPasskey')

    const passkeyButton = document.getElementById('passkey-button')
    if (passkeyButton) {
        passkeyButton.addEventListener('click', async () => {
            await startPasskeyLogin()
        })
    }
}

async function startPasskeyLogin(): Promise<void> {

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