import { base64urlToBuffer, serializeCredential } from "../passkeys_helper";

export async function initOnboardingWithPasskey(): Promise<void> {

    console.log('initializing OnboardingWithPasskey')

    const passkeyButton = document.getElementById('passkey-button')
    if (passkeyButton) {
        passkeyButton.addEventListener('click', async () => {
            await startPasskeyOnboarding()
        })
    }
}

async function startPasskeyOnboarding(): Promise<void> {

    try {

        // Get the email input element
        const emailInput = document.getElementById("email") as HTMLInputElement;
        const email = emailInput?.value;
        if (!email) {
            // Use browser native validation highlighting
            emailInput.reportValidity();
            emailInput.focus();
            return;
        }

        // Get the action
        const actionBtn = (document.getElementById("action") as HTMLInputElement);
        actionBtn.value = "passkey";

        // Get the passkeys options
        const optionsInput = (document.getElementById("passkeysOptions") as HTMLInputElement)?.value;
        const options = JSON.parse(optionsInput);
    
        // Overwrite the username with the email
        options.publicKey.user.name = email;
        options.publicKey.user.displayName = email;

        options.publicKey.challenge = base64urlToBuffer(options.publicKey.challenge);
        options.publicKey.user.id = base64urlToBuffer(options.publicKey.user.id);
    
        // Toggle this flag to use the mock instead of the real API
        const useMock = false;
    
        let cred;
        if (useMock) {
          
          cred = JSON.parse('{"authenticatorAttachment":"platform","clientExtensionResults":{},"id":"LjWKi6SQpjaO1zxsK0JgVmnwyl4ptYcaHj6yWg7Fzp8","rawId":"LjWKi6SQpjaO1zxsK0JgVmnwyl4ptYcaHj6yWg7Fzp8","response":{"attestationObject":"o2NmbXRkbm9uZWdhdHRTdG10oGhhdXRoRGF0YVikSZYN5YgOjGh0NBcPZHZgW4_krrmihjLHmVzzuoMdl2NFAAAAAK3OAAI1vMYKZIsLJfHwVQMAIC41ioukkKY2jtc8bCtCYFZp8MpeKbWHGh4-sloOxc6fpQECAyYgASFYIOlLScYq6Jiu4v3-iHAqu7foa9UJqbEnWWSqUW07fucCIlgg-sOCWpkgmPMx5ypb9hOxB0IzA4SDUcqljI15KnL99TI","authenticatorData":"SZYN5YgOjGh0NBcPZHZgW4_krrmihjLHmVzzuoMdl2NFAAAAAK3OAAI1vMYKZIsLJfHwVQMAIC41ioukkKY2jtc8bCtCYFZp8MpeKbWHGh4-sloOxc6fpQECAyYgASFYIOlLScYq6Jiu4v3-iHAqu7foa9UJqbEnWWSqUW07fucCIlgg-sOCWpkgmPMx5ypb9hOxB0IzA4SDUcqljI15KnL99TI","clientDataJSON":"eyJ0eXBlIjoid2ViYXV0aG4uY3JlYXRlIiwiY2hhbGxlbmdlIjoiRFFRazZSb3ltQ3N5RWVjU3NXQ3JBOXk1MFFYazViemxrOElCbzNqVFVhZyIsIm9yaWdpbiI6Imh0dHA6Ly9sb2NhbGhvc3Q6ODA4MCIsImNyb3NzT3JpZ2luIjpmYWxzZX0","publicKey":"MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAE6UtJxiromK7i_f6IcCq7t-hr1QmpsSdZZKpRbTt-5wL6w4JamSCY8zHnKlv2E7EHQjMDhINRyqWMjXkqcv31Mg","publicKeyAlgorithm":-7,"transports":["internal"]},"type":"public-key"}')
    
        } else {
          cred = await navigator.credentials.create({ publicKey: options.publicKey });
        }
    
        serializedCred = serializeCredential(cred)
        document.getElementById("passkeysFinishRegistrationJson").value = JSON.stringify(serializedCred);
        document.getElementById("onboarding-with-passkey-form").submit();
    
      } catch (err) {
        alert("Passkey registration failed: " + err.message);
        console.error(err);
      }
}